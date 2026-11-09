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
	providerError "github.com/IBM/ibmcloud-volume-interface/lib/utils"
	"github.com/IBM/ibmcloud-volume-vpc/common/vpcclient/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Create a group response with two member snapshots for the tests.
func newGroupResponse(lifecycleState, secondMemberState string) *models.SnapshotConsistencyGroupResponse {
	createdAt := time.Date(2026, time.January, 2, 3, 4, 5, 0, time.UTC)
	return &models.SnapshotConsistencyGroupResponse{
		ID:             "group-snapshot-id",
		CRN:            "group-snapshot-crn",
		Href:           "group-snapshot-href",
		LifecycleState: lifecycleState,
		CreatedAt:      &createdAt,
		Snapshots: []*models.Snapshot{
			{
				ID:              "snapshot-id-1",
				CRN:             "snapshot-crn-1",
				Href:            "snapshot-href-1",
				MinimumCapacity: 10,
				CreatedAt:       &createdAt,
				LifecycleState:  snapshotReadyState,
				SourceVolume:    &models.SourceVolume{ID: "volume-id-1"},
			},
			{
				ID:              "snapshot-id-2",
				CRN:             "snapshot-crn-2",
				MinimumCapacity: 20,
				CreatedAt:       &createdAt,
				LifecycleState:  secondMemberState,
				SourceVolume:    &models.SourceVolume{ID: "volume-id-2"},
			},
		},
	}
}

func TestGetGroupSnapshot(t *testing.T) {
	notFoundErr := &models.Error{Errors: []models.ErrorItem{
		{Code: SnapshotConsistencyGroupNotFound},
	}}
	internalErr := &models.Error{Errors: []models.ErrorItem{
		{Code: "internal_error", Status: "500 Internal Server Error"},
	}}

	testCases := []struct {
		name       string
		groupID    string
		group      *models.SnapshotConsistencyGroupResponse
		groupErr   error
		maxRetries int // 0 keeps the defaults; a positive value sets total attempts, including the first call.

		wantErrCode     string
		wantErrText     string
		wantBackendText string
		wantReady       bool
		wantVolumeIDs   []string
		wantGroupCalls  int
	}{
		{
			name:           "stable group with ready members",
			groupID:        "group-snapshot-id",
			group:          newGroupResponse(snapshotReadyState, snapshotReadyState),
			wantReady:      true,
			wantVolumeIDs:  []string{"volume-id-1", "volume-id-2"},
			wantGroupCalls: 1,
		},
		{
			name:           "pending member makes the group not ready",
			groupID:        "group-snapshot-id",
			group:          newGroupResponse(snapshotReadyState, "pending"),
			wantReady:      false,
			wantVolumeIDs:  []string{"volume-id-1", "volume-id-2"},
			wantGroupCalls: 1,
		},
		{
			name:           "pending group is not ready even when members are ready",
			groupID:        "group-snapshot-id",
			group:          newGroupResponse("pending", snapshotReadyState),
			wantReady:      false,
			wantVolumeIDs:  []string{"volume-id-1", "volume-id-2"},
			wantGroupCalls: 1,
		},
		{
			name:            "backend error is wrapped as GroupSnapshotRetrievalFailed",
			groupID:         "group-snapshot-id",
			groupErr:        internalErr,
			maxRetries:      1,
			wantErrCode:     "GroupSnapshotRetrievalFailed",
			wantErrText:     "group-snapshot-id",
			wantBackendText: "internal_error",
			wantGroupCalls:  1,
		},
		{
			name:            "not-found error is not retried",
			groupID:         "group-snapshot-id",
			groupErr:        notFoundErr,
			maxRetries:      3,
			wantErrCode:     "GroupSnapshotRetrievalFailed",
			wantErrText:     "group-snapshot-id",
			wantBackendText: SnapshotConsistencyGroupNotFound,
			wantGroupCalls:  1,
		},
		{
			name:           "empty backend response is invalid",
			groupID:        "group-snapshot-id",
			wantErrCode:    "InvalidGroupSnapshotResponse",
			wantErrText:    "the response is empty",
			wantGroupCalls: 1,
		},
		{
			name:           "group without member snapshots",
			groupID:        "group-snapshot-id",
			group:          &models.SnapshotConsistencyGroupResponse{ID: "group-snapshot-id"},
			wantErrCode:    "GroupSnapshotMemberDetailsUnavailable",
			wantErrText:    "group-snapshot-id",
			wantGroupCalls: 1,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			logger, teardown := GetTestLogger(t)
			defer teardown()

			vpcs, uc, _, err := GetTestOpenSession(t, logger)
			require.NoError(t, err)

			// Set this after opening the session, which loads the retry settings from config.
			if tc.maxRetries > 0 {
				limitRetries(t, tc.maxRetries)
			}

			groupService := &fakeSnapshotConsistencyGroupManager{getResult: tc.group, getErr: tc.groupErr}
			uc.SnapshotConsistencyGroupServiceReturns(groupService)

			result, err := vpcs.GetGroupSnapshot(tc.groupID)

			assert.Equal(t, tc.wantGroupCalls, groupService.getCalls)
			assert.Zero(t, uc.SnapshotServiceCallCount())
			assert.Equal(t, tc.groupID, groupService.getGroupID)

			switch {
			case tc.wantErrCode != "":
				assert.Nil(t, result)
				requireProviderError(t, err, tc.wantErrCode, tc.wantErrText, tc.wantBackendText)
			default:
				require.NoError(t, err)
				require.NotNil(t, result)
				assert.Equal(t, tc.group.ID, result.GroupSnapshotID)
				assert.Equal(t, tc.group.CRN, result.GroupSnapshotCRN)
				assert.Equal(t, tc.wantReady, result.ReadyToUse)
				var gotVolumeIDs []string
				require.Len(t, result.Snapshots, len(tc.group.Snapshots))
				for i, member := range result.Snapshots {
					gotVolumeIDs = append(gotVolumeIDs, member.VolumeID)
					assert.Equal(t, tc.group.Snapshots[i].LifecycleState == snapshotReadyState, member.ReadyToUse)
				}
				assert.Equal(t, tc.wantVolumeIDs, gotVolumeIDs)
			}
		})
	}
}

func TestGetGroupSnapshotByName(t *testing.T) {
	notFoundErr := &models.Error{Errors: []models.ErrorItem{
		{Code: SnapshotConsistencyGroupNotFound},
	}}
	forbiddenErr := &models.Error{Errors: []models.ErrorItem{
		{Code: "snapshots_not_authorized", Status: "403 Forbidden"},
	}}

	testCases := []struct {
		name          string
		groupName     string
		resourceGroup string
		group         *models.SnapshotConsistencyGroupResponse
		groupErr      error
		maxRetries    int // 0 keeps the defaults; a positive value sets total attempts, including the first call.

		wantErrCode       string // Expected provider error code; empty means no error is expected.
		wantBackendText   string
		wantNilResult     bool
		wantVolumeIDs     []string
		wantByNameCalls   int
		wantQueriedByName string
	}{
		{
			name:              "name and resource group are forwarded to the backend",
			groupName:         "my-group",
			resourceGroup:     "my-resource-group",
			group:             newGroupResponse(snapshotReadyState, snapshotReadyState),
			wantVolumeIDs:     []string{"volume-id-1", "volume-id-2"},
			wantByNameCalls:   1,
			wantQueriedByName: "my-group",
		},
		{
			name:              "no matching group returns nil without an error",
			groupName:         "nonexistent-group",
			resourceGroup:     "my-resource-group",
			wantNilResult:     true,
			wantByNameCalls:   1,
			wantQueriedByName: "nonexistent-group",
		},
		{
			name:            "empty name is rejected without calling the backend",
			groupName:       "",
			resourceGroup:   "my-resource-group",
			wantErrCode:     "InvalidGroupSnapshotName",
			wantByNameCalls: 0,
		},
		{
			name:              "backend error is wrapped as GroupSnapshotNameLookupFailed",
			groupName:         "group-name",
			resourceGroup:     "my-resource-group",
			groupErr:          forbiddenErr,
			maxRetries:        1,
			wantErrCode:       "GroupSnapshotNameLookupFailed",
			wantBackendText:   "snapshots_not_authorized",
			wantByNameCalls:   1,
			wantQueriedByName: "group-name",
		},
		{
			name:              "not-found error is not retried",
			groupName:         "group-name",
			resourceGroup:     "my-resource-group",
			groupErr:          notFoundErr,
			maxRetries:        3,
			wantErrCode:       "GroupSnapshotNameLookupFailed",
			wantBackendText:   SnapshotConsistencyGroupNotFound,
			wantByNameCalls:   1,
			wantQueriedByName: "group-name",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			logger, teardown := GetTestLogger(t)
			defer teardown()

			vpcs, uc, _, err := GetTestOpenSession(t, logger)
			require.NoError(t, err)

			// Set this after opening the session, which loads the retry settings from config.
			if tc.maxRetries > 0 {
				limitRetries(t, tc.maxRetries)
			}

			groupService := &fakeSnapshotConsistencyGroupManager{getByNameResult: tc.group, getByNameErr: tc.groupErr}
			uc.SnapshotConsistencyGroupServiceReturns(groupService)

			result, err := vpcs.GetGroupSnapshotByName(tc.groupName, tc.resourceGroup)

			assert.Equal(t, tc.wantByNameCalls, groupService.getByNameCalls)
			assert.Zero(t, uc.SnapshotServiceCallCount())
			if tc.wantByNameCalls > 0 {
				assert.Equal(t, tc.wantQueriedByName, groupService.getByNameName)
				assert.Equal(t, tc.resourceGroup, groupService.getByNameResourceGroupID)
			}

			switch {
			case tc.wantErrCode != "":
				assert.Nil(t, result)
				requireProviderError(t, err, tc.wantErrCode, "", tc.wantBackendText)
			case tc.wantNilResult:
				require.NoError(t, err)
				assert.Nil(t, result)
			default:
				require.NoError(t, err)
				require.NotNil(t, result)
				assert.Equal(t, tc.group.ID, result.GroupSnapshotID)
				var gotVolumeIDs []string
				for _, member := range result.Snapshots {
					gotVolumeIDs = append(gotVolumeIDs, member.VolumeID)
				}
				assert.Equal(t, tc.wantVolumeIDs, gotVolumeIDs)
			}
		})
	}
}

func TestGroupSnapshotReturnsAvailableMemberDetails(t *testing.T) {
	testCases := []struct {
		name        string
		changeGroup func(*models.SnapshotConsistencyGroupResponse)
		wantMembers int
		wantReady   bool
		checkMember func(*testing.T, *provider.Snapshot)
	}{
		{
			name: "null member",
			changeGroup: func(group *models.SnapshotConsistencyGroupResponse) {
				group.Snapshots[1] = nil
			},
			wantMembers: 1,
		},
		{
			name: "member reference without details",
			changeGroup: func(group *models.SnapshotConsistencyGroupResponse) {
				group.Snapshots[1] = &models.Snapshot{ID: "snapshot-id-2", CRN: "snapshot-crn-2"}
			},
			wantMembers: 2,
			checkMember: func(t *testing.T, member *provider.Snapshot) {
				assert.Equal(t, "snapshot-id-2", member.SnapshotID)
				assert.Equal(t, "snapshot-crn-2", member.SnapshotCRN)
				assert.Empty(t, member.VolumeID)
				assert.True(t, member.SnapshotCreationTime.IsZero())
				assert.Zero(t, member.SnapshotSize)
				assert.False(t, member.ReadyToUse)
			},
		},
		{
			name: "missing source volume ID",
			changeGroup: func(group *models.SnapshotConsistencyGroupResponse) {
				group.Snapshots[1].SourceVolume.ID = ""
			},
			wantMembers: 2,
			checkMember: func(t *testing.T, member *provider.Snapshot) {
				assert.Empty(t, member.VolumeID)
				assert.False(t, member.ReadyToUse)
			},
		},
		{
			name: "missing member CRN",
			changeGroup: func(group *models.SnapshotConsistencyGroupResponse) {
				group.Snapshots[1].CRN = ""
			},
			wantMembers: 2,
			wantReady:   true,
			checkMember: func(t *testing.T, member *provider.Snapshot) {
				assert.Empty(t, member.SnapshotCRN)
				assert.Equal(t, "snapshot-id-2", member.SnapshotID)
			},
		},
		{
			name: "missing member creation time",
			changeGroup: func(group *models.SnapshotConsistencyGroupResponse) {
				group.Snapshots[1].CreatedAt = nil
			},
			wantMembers: 2,
			wantReady:   true,
			checkMember: func(t *testing.T, member *provider.Snapshot) {
				assert.True(t, member.SnapshotCreationTime.IsZero())
			},
		},
		{
			name: "missing member lifecycle state",
			changeGroup: func(group *models.SnapshotConsistencyGroupResponse) {
				group.Snapshots[1].LifecycleState = ""
			},
			wantMembers: 2,
			checkMember: func(t *testing.T, member *provider.Snapshot) {
				assert.False(t, member.ReadyToUse)
			},
		},
	}

	for _, tc := range testCases {
		for _, operation := range groupSnapshotOperations {
			t.Run(tc.name+"/"+operation.name, func(t *testing.T) {
				logger, teardown := GetTestLogger(t)
				defer teardown()
				session, client, _, err := GetTestOpenSession(t, logger)
				require.NoError(t, err)

				group := newGroupResponse(snapshotReadyState, snapshotReadyState)
				tc.changeGroup(group)
				groupService := &fakeSnapshotConsistencyGroupManager{
					createResult: group, getResult: group, getByNameResult: group,
				}
				client.SnapshotConsistencyGroupServiceReturns(groupService)

				result, err := operation.call(session)

				require.NoError(t, err)
				require.NotNil(t, result)
				require.Len(t, result.Snapshots, tc.wantMembers)
				assert.Equal(t, "snapshot-id-1", result.Snapshots[0].SnapshotID)
				assert.Equal(t, tc.wantReady, result.ReadyToUse)
				if tc.checkMember != nil {
					tc.checkMember(t, result.Snapshots[1])
				}
				assert.Equal(t, 1, groupService.createCalls+groupService.getCalls+groupService.getByNameCalls)
				assert.Zero(t, client.SnapshotServiceCallCount())
			})
		}
	}
}

// Malformed group responses are errors, not missing groups.
func TestGroupSnapshotInvalidResponse(t *testing.T) {
	testCases := []struct {
		name       string
		group      *models.SnapshotConsistencyGroupResponse
		wantReason string
	}{
		{name: "empty response", wantReason: "the response is empty"},
		{
			name:       "missing group ID",
			group:      &models.SnapshotConsistencyGroupResponse{LifecycleState: snapshotReadyState},
			wantReason: "the group ID is missing",
		},
	}
	for _, tc := range testCases {
		for _, operation := range groupSnapshotOperations {
			t.Run(tc.name+"/"+operation.name, func(t *testing.T) {
				logger, teardown := GetTestLogger(t)
				defer teardown()
				session, client, _, err := GetTestOpenSession(t, logger)
				require.NoError(t, err)

				groupService := &fakeSnapshotConsistencyGroupManager{
					createResult: tc.group, getResult: tc.group, getByNameResult: tc.group,
				}
				client.SnapshotConsistencyGroupServiceReturns(groupService)

				result, err := operation.call(session)

				assert.Nil(t, result)
				assert.Equal(t, 1, groupService.createCalls+groupService.getCalls+groupService.getByNameCalls)
				assert.Zero(t, client.SnapshotServiceCallCount())
				// A name lookup uses nil to mean no matching group.
				if tc.group == nil && operation.name == "get by name" {
					require.NoError(t, err)
					return
				}

				requireProviderError(t, err, "InvalidGroupSnapshotResponse", tc.wantReason, "")
				var responseError providerError.Message
				require.ErrorAs(t, err, &responseError)
				assert.Equal(t, 500, responseError.RC)
				assert.Equal(t, providerError.RetrivalFailed, responseError.Type)
				assert.NotEqual(t, providerError.EntityNotFound, responseError.Type)
			})
		}
	}
}
