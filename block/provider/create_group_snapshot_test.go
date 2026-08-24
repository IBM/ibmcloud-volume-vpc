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

package provider

import (
	"testing"
	"time"

	"github.com/IBM/ibmcloud-volume-interface/lib/provider"
	"github.com/IBM/ibmcloud-volume-vpc/common/vpcclient/models"
	serviceFakes "github.com/IBM/ibmcloud-volume-vpc/common/vpcclient/vpcvolume/fakes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

type fakeSnapshotConsistencyGroupManager struct {
	createRequest *models.SnapshotConsistencyGroupRequest
	createResult  *models.SnapshotConsistencyGroup
	createErr     error

	deleteGroupID string
	deleteErr     error

	getGroupID string
	getResult  *models.SnapshotConsistencyGroup
	getErr     error

	getByNameName            string
	getByNameResourceGroupID string
	getByNameResult          *models.SnapshotConsistencyGroup
	getByNameErr             error

	listFilters *models.ListSnapshotConsistencyGroupFilters
	listResult  *models.SnapshotConsistencyGroupList
	listErr     error
}

func (fake *fakeSnapshotConsistencyGroupManager) CreateSnapshotConsistencyGroup(req *models.SnapshotConsistencyGroupRequest, logger *zap.Logger) (*models.SnapshotConsistencyGroup, error) {
	fake.createRequest = req
	return fake.createResult, fake.createErr
}

func (fake *fakeSnapshotConsistencyGroupManager) DeleteSnapshotConsistencyGroup(groupID string, logger *zap.Logger) error {
	fake.deleteGroupID = groupID
	return fake.deleteErr
}

func (fake *fakeSnapshotConsistencyGroupManager) GetSnapshotConsistencyGroup(groupID string, logger *zap.Logger) (*models.SnapshotConsistencyGroup, error) {
	fake.getGroupID = groupID
	return fake.getResult, fake.getErr
}

func (fake *fakeSnapshotConsistencyGroupManager) GetSnapshotConsistencyGroupByName(name, resourceGroupID string, logger *zap.Logger) (*models.SnapshotConsistencyGroup, error) {
	fake.getByNameName = name
	fake.getByNameResourceGroupID = resourceGroupID
	return fake.getByNameResult, fake.getByNameErr
}

func (fake *fakeSnapshotConsistencyGroupManager) ListSnapshotConsistencyGroups(limit int, start string, filters *models.ListSnapshotConsistencyGroupFilters, logger *zap.Logger) (*models.SnapshotConsistencyGroupList, error) {
	fake.listFilters = filters
	return fake.listResult, fake.listErr
}

func TestCreateGroupSnapshotCreatesConsistencyGroupAndListsMembers(t *testing.T) {
	logger, teardown := GetTestLogger(t)
	defer teardown()

	vpcs, uc, _, err := GetTestOpenSession(t, logger)
	require.NoError(t, err)

	now := time.Now()
	groupService := &fakeSnapshotConsistencyGroupManager{
		createResult: &models.SnapshotConsistencyGroup{
			ID:             "group-snapshot-id",
			CRN:            "group-snapshot-crn",
			Href:           "group-snapshot-href",
			LifecycleState: snapshotReadyState,
			CreatedAt:      &now,
		},
	}
	snapshotService := &serviceFakes.SnapshotManager{}
	snapshotService.ListSnapshotsReturns(&models.SnapshotList{
		Snapshots: []*models.Snapshot{
			{
				ID:              "snapshot-id-1",
				CRN:             "snapshot-crn-1",
				MinimumCapacity: 10,
				LifecycleState:  snapshotReadyState,
				SourceVolume:    &models.SourceVolume{ID: "volume-id-1"},
			},
			{
				ID:              "snapshot-id-2",
				CRN:             "snapshot-crn-2",
				MinimumCapacity: 20,
				LifecycleState:  snapshotReadyState,
				SourceVolume:    &models.SourceVolume{ID: "volume-id-2"},
			},
		},
	}, nil)
	uc.SnapshotConsistencyGroupServiceReturns(groupService)
	uc.SnapshotServiceReturns(snapshotService)

	groupSnapshot, err := vpcs.CreateGroupSnapshot([]string{"volume-id-1", "volume-id-2"}, provider.GroupSnapshotParameters{
		Name:          "group-snapshot-name",
		ResourceGroup: "resource-group-id",
	})

	require.NoError(t, err)
	require.NotNil(t, groupSnapshot)
	assert.Equal(t, "group-snapshot-id", groupSnapshot.GroupSnapshotID)
	assert.Equal(t, "group-snapshot-crn", groupSnapshot.GroupSnapshotCRN)
	assert.True(t, groupSnapshot.ReadyToUse)
	assert.Len(t, groupSnapshot.Snapshots, 2)
	assert.Equal(t, "snapshot-id-1", groupSnapshot.Snapshots[0].SnapshotID)
	assert.Equal(t, "volume-id-1", groupSnapshot.Snapshots[0].VolumeID)
	assert.Equal(t, "group-snapshot-name", groupService.createRequest.Name)
	assert.Equal(t, "resource-group-id", groupService.createRequest.ResourceGroup.ID)
	assert.True(t, groupService.createRequest.DeleteSnapshotsOnDelete)
	require.Len(t, groupService.createRequest.Snapshots, 2)
	assert.Equal(t, "volume-id-1", groupService.createRequest.Snapshots[0].SourceVolume.ID)
	assert.Equal(t, "volume-id-2", groupService.createRequest.Snapshots[1].SourceVolume.ID)
	_, _, filters, _ := snapshotService.ListSnapshotsArgsForCall(0)
	assert.Equal(t, "group-snapshot-id", filters.SnapshotConsistencyGroupID)
}

func TestDeleteGroupSnapshotDeletesConsistencyGroupByID(t *testing.T) {
	logger, teardown := GetTestLogger(t)
	defer teardown()

	vpcs, uc, _, err := GetTestOpenSession(t, logger)
	require.NoError(t, err)

	groupService := &fakeSnapshotConsistencyGroupManager{}
	uc.SnapshotConsistencyGroupServiceReturns(groupService)

	err = vpcs.DeleteGroupSnapshot("group-snapshot-id", []string{"snapshot-id-1", "snapshot-id-2"})

	require.NoError(t, err)
	assert.Equal(t, "group-snapshot-id", groupService.deleteGroupID)
}

func TestGetGroupSnapshotGetsConsistencyGroupAndListsMembers(t *testing.T) {
	logger, teardown := GetTestLogger(t)
	defer teardown()

	vpcs, uc, _, err := GetTestOpenSession(t, logger)
	require.NoError(t, err)

	groupService := &fakeSnapshotConsistencyGroupManager{
		getResult: &models.SnapshotConsistencyGroup{
			ID:             "group-snapshot-id",
			CRN:            "group-snapshot-crn",
			LifecycleState: snapshotReadyState,
			Snapshots: []models.SnapshotReference{
				{ID: "snapshot-id-1", CRN: "snapshot-crn-1"},
			},
		},
	}
	snapshotService := &serviceFakes.SnapshotManager{}
	snapshotService.ListSnapshotsReturns(&models.SnapshotList{
		Snapshots: []*models.Snapshot{
			{
				ID:              "snapshot-id-1",
				CRN:             "snapshot-crn-1",
				MinimumCapacity: 10,
				LifecycleState:  snapshotReadyState,
				SourceVolume:    &models.SourceVolume{ID: "volume-id-1"},
			},
		},
	}, nil)
	uc.SnapshotConsistencyGroupServiceReturns(groupService)
	uc.SnapshotServiceReturns(snapshotService)

	groupSnapshot, err := vpcs.GetGroupSnapshot("group-snapshot-id")

	require.NoError(t, err)
	require.NotNil(t, groupSnapshot)
	assert.Equal(t, "group-snapshot-id", groupService.getGroupID)
	assert.Equal(t, "group-snapshot-id", groupSnapshot.GroupSnapshotID)
	assert.True(t, groupSnapshot.ReadyToUse)
	require.Len(t, groupSnapshot.Snapshots, 1)
	assert.Equal(t, "volume-id-1", groupSnapshot.Snapshots[0].VolumeID)
	_, _, filters, _ := snapshotService.ListSnapshotsArgsForCall(0)
	assert.Equal(t, "group-snapshot-id", filters.SnapshotConsistencyGroupID)
}

func TestFromProviderToLibGroupSnapshotReadiness(t *testing.T) {
	logger, teardown := GetTestLogger(t)
	defer teardown()

	testCases := []struct {
		name            string
		groupState      string
		snapshotDetails []*models.Snapshot
		expectedReady   bool
	}{
		{
			name:       "stable group with all stable members",
			groupState: snapshotReadyState,
			snapshotDetails: []*models.Snapshot{
				{ID: "snapshot-id-1", LifecycleState: snapshotReadyState, SourceVolume: &models.SourceVolume{ID: "volume-id-1"}},
				{ID: "snapshot-id-2", LifecycleState: snapshotReadyState, SourceVolume: &models.SourceVolume{ID: "volume-id-2"}},
			},
			expectedReady: true,
		},
		{
			name:       "stable group with a pending member",
			groupState: snapshotReadyState,
			snapshotDetails: []*models.Snapshot{
				{ID: "snapshot-id-1", LifecycleState: snapshotReadyState, SourceVolume: &models.SourceVolume{ID: "volume-id-1"}},
				{ID: "snapshot-id-2", LifecycleState: "pending", SourceVolume: &models.SourceVolume{ID: "volume-id-2"}},
			},
			expectedReady: false,
		},
		{
			name:       "pending group with stable members",
			groupState: "pending",
			snapshotDetails: []*models.Snapshot{
				{ID: "snapshot-id-1", LifecycleState: snapshotReadyState, SourceVolume: &models.SourceVolume{ID: "volume-id-1"}},
			},
			expectedReady: false,
		},
		{
			name:            "stable group without full member details",
			groupState:      snapshotReadyState,
			snapshotDetails: nil,
			expectedReady:   false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			group := &models.SnapshotConsistencyGroup{
				ID:             "group-snapshot-id",
				LifecycleState: tc.groupState,
				Snapshots: []models.SnapshotReference{
					{ID: "snapshot-id-1"},
				},
			}

			result := FromProviderToLibGroupSnapshot(group, tc.snapshotDetails, logger)

			require.NotNil(t, result)
			assert.Equal(t, tc.expectedReady, result.ReadyToUse)
		})
	}
}
