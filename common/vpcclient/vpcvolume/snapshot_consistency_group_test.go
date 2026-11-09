/**
 * Copyright 2026 IBM Corp.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package vpcvolume_test

import (
	"encoding/json"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/IBM/ibmcloud-volume-vpc/common/vpcclient/models"
	"github.com/IBM/ibmcloud-volume-vpc/common/vpcclient/riaas/test"
	"github.com/IBM/ibmcloud-volume-vpc/common/vpcclient/vpcvolume"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const groupResponseWithMember = `{
	"id": "group-snapshot-id",
	"crn": "group-snapshot-crn",
	"lifecycle_state": "stable",
	"snapshots": [{
		"id": "snapshot-id-1",
		"crn": "snapshot-crn-1",
		"source_volume": {"id": "volume-id-1"},
		"created_at": "2026-01-02T03:04:05Z",
		"minimum_capacity": 10,
		"lifecycle_state": "stable"
	}]
}`

func assertGroupMember(t *testing.T, group *models.SnapshotConsistencyGroupResponse) {
	t.Helper()
	require.NotNil(t, group)
	assert.Equal(t, "group-snapshot-id", group.ID)
	require.Len(t, group.Snapshots, 1)
	member := group.Snapshots[0]
	require.NotNil(t, member)
	assert.Equal(t, "snapshot-id-1", member.ID)
	assert.Equal(t, "snapshot-crn-1", member.CRN)
	require.NotNil(t, member.SourceVolume)
	assert.Equal(t, "volume-id-1", member.SourceVolume.ID)
	require.NotNil(t, member.CreatedAt)
	assert.Equal(t, time.Date(2026, time.January, 2, 3, 4, 5, 0, time.UTC), *member.CreatedAt)
	assert.EqualValues(t, 10, member.MinimumCapacity)
	assert.Equal(t, "stable", member.LifecycleState)
}

func TestCreateSnapshotConsistencyGroup(t *testing.T) {
	logger, _ := GetTestContextLogger()
	defer logger.Sync()

	mux, client, teardown := test.SetupServer(t)
	defer teardown()

	test.SetupMuxResponse(t, mux, vpcvolume.Version+"/snapshot_consistency_groups", http.MethodPost, nil, http.StatusCreated, groupResponseWithMember, func(t *testing.T, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		require.NoError(t, err)

		var requestBody models.SnapshotConsistencyGroupRequest
		require.NoError(t, json.Unmarshal(body, &requestBody))
		assert.Equal(t, "group-snapshot-name", requestBody.Name)
		assert.Equal(t, "resource-group-id", requestBody.ResourceGroup.ID)
		assert.True(t, requestBody.DeleteSnapshotsOnDelete)
		require.Len(t, requestBody.Snapshots, 2)
		assert.Equal(t, "volume-id-1", requestBody.Snapshots[0].SourceVolume.ID)
		assert.Equal(t, "volume-id-2", requestBody.Snapshots[1].SourceVolume.ID)
	})

	groupService := vpcvolume.NewSnapshotConsistencyGroupManager(client)

	group, err := groupService.CreateSnapshotConsistencyGroup(&models.SnapshotConsistencyGroupRequest{
		Name:                    "group-snapshot-name",
		ResourceGroup:           &models.ResourceGroup{ID: "resource-group-id"},
		DeleteSnapshotsOnDelete: true,
		Snapshots: []models.GroupSnapshotTemplate{
			{SourceVolume: &models.SourceVolume{ID: "volume-id-1"}},
			{SourceVolume: &models.SourceVolume{ID: "volume-id-2"}},
		},
	}, logger)

	require.NoError(t, err)
	assertGroupMember(t, group)
}

func TestDeleteSnapshotConsistencyGroup(t *testing.T) {
	logger, _ := GetTestContextLogger()
	defer logger.Sync()

	mux, client, teardown := test.SetupServer(t)
	defer teardown()

	test.SetupMuxResponse(t, mux, vpcvolume.Version+"/snapshot_consistency_groups/group-snapshot-id", http.MethodDelete, nil, http.StatusNoContent, "", nil)

	groupService := vpcvolume.NewSnapshotConsistencyGroupManager(client)

	err := groupService.DeleteSnapshotConsistencyGroup("group-snapshot-id", logger)

	require.NoError(t, err)
}

func TestGetSnapshotConsistencyGroup(t *testing.T) {
	logger, _ := GetTestContextLogger()
	defer logger.Sync()

	mux, client, teardown := test.SetupServer(t)
	defer teardown()

	test.SetupMuxResponse(t, mux, vpcvolume.Version+"/snapshot_consistency_groups/group-snapshot-id", http.MethodGet, nil, http.StatusOK, groupResponseWithMember, nil)

	groupService := vpcvolume.NewSnapshotConsistencyGroupManager(client)

	group, err := groupService.GetSnapshotConsistencyGroup("group-snapshot-id", logger)

	require.NoError(t, err)
	assertGroupMember(t, group)
	assert.Equal(t, "group-snapshot-crn", group.CRN)
}

func TestGetSnapshotConsistencyGroupByName(t *testing.T) {
	logger, _ := GetTestContextLogger()
	defer logger.Sync()

	mux, client, teardown := test.SetupServer(t)
	defer teardown()

	response := `{"snapshot_consistency_groups":[` + groupResponseWithMember + `]}`
	test.SetupMuxResponse(t, mux, vpcvolume.Version+"/snapshot_consistency_groups", http.MethodGet, nil, http.StatusOK, response, func(t *testing.T, r *http.Request) {
		assert.Equal(t, "group-snapshot-name", r.URL.Query().Get("name"))
		assert.Equal(t, "resource-group-id", r.URL.Query().Get("resource_group.id"))
	})

	groupService := vpcvolume.NewSnapshotConsistencyGroupManager(client)
	group, err := groupService.GetSnapshotConsistencyGroupByName("group-snapshot-name", "resource-group-id", logger)

	require.NoError(t, err)
	assertGroupMember(t, group)
}

func TestGetSnapshotConsistencyGroupByNameEmptyOrInvalidResponse(t *testing.T) {
	logger, _ := GetTestContextLogger()
	defer logger.Sync()

	testCases := []struct {
		name     string
		response string
		wantErr  string
	}{
		{name: "empty list is no match", response: `{"snapshot_consistency_groups":[]}`},
		{
			name:     "null group is invalid",
			response: `{"snapshot_consistency_groups":[null]}`,
			wantErr:  "contained a null group",
		},
		{
			name:     "null list is invalid",
			response: `{"snapshot_consistency_groups":null}`,
			wantErr:  "did not include a group list",
		},
		{name: "missing list is invalid", response: `{}`, wantErr: "did not include a group list"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mux, client, teardown := test.SetupServer(t)
			defer teardown()
			test.SetupMuxResponse(t, mux, vpcvolume.Version+"/snapshot_consistency_groups", http.MethodGet, nil, http.StatusOK, tc.response, nil)

			service := vpcvolume.NewSnapshotConsistencyGroupManager(client)
			group, err := service.GetSnapshotConsistencyGroupByName("group-snapshot-name", "resource-group-id", logger)

			assert.Nil(t, group)
			if tc.wantErr != "" {
				require.ErrorContains(t, err, tc.wantErr)
				return
			}
			require.NoError(t, err)
		})
	}
}
