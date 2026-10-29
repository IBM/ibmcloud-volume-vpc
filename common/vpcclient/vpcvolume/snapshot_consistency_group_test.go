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

	"github.com/IBM/ibmcloud-volume-vpc/common/vpcclient/models"
	"github.com/IBM/ibmcloud-volume-vpc/common/vpcclient/riaas/test"
	"github.com/IBM/ibmcloud-volume-vpc/common/vpcclient/vpcvolume"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateSnapshotConsistencyGroup(t *testing.T) {
	logger, _ := GetTestContextLogger()
	defer logger.Sync()

	mux, client, teardown := test.SetupServer(t)
	defer teardown()

	test.SetupMuxResponse(t, mux, vpcvolume.Version+"/snapshot_consistency_groups", http.MethodPost, nil, http.StatusOK, `{"id":"group-snapshot-id","lifecycle_state":"stable"}`, func(t *testing.T, r *http.Request) {
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
	require.NotNil(t, group)
	assert.Equal(t, "group-snapshot-id", group.ID)
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

	test.SetupMuxResponse(t, mux, vpcvolume.Version+"/snapshot_consistency_groups/group-snapshot-id", http.MethodGet, nil, http.StatusOK, `{"id":"group-snapshot-id","crn":"group-snapshot-crn","lifecycle_state":"stable","snapshots":[{"id":"snapshot-id-1"}]}`, nil)

	groupService := vpcvolume.NewSnapshotConsistencyGroupManager(client)

	group, err := groupService.GetSnapshotConsistencyGroup("group-snapshot-id", logger)

	require.NoError(t, err)
	require.NotNil(t, group)
	assert.Equal(t, "group-snapshot-id", group.ID)
	assert.Equal(t, "group-snapshot-crn", group.CRN)
	require.Len(t, group.Snapshots, 1)
	assert.Equal(t, "snapshot-id-1", group.Snapshots[0].ID)
}
