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
	"errors"
	"testing"
	"time"

	"github.com/IBM/ibmcloud-volume-interface/lib/provider"
	providerError "github.com/IBM/ibmcloud-volume-interface/lib/utils"
	userError "github.com/IBM/ibmcloud-volume-vpc/common/messages"
	"github.com/IBM/ibmcloud-volume-vpc/common/vpcclient/models"
	serviceFakes "github.com/IBM/ibmcloud-volume-vpc/common/vpcclient/vpcvolume/fakes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func init() {
	userError.MessagesEn = userError.InitMessages()
}

type fakeSnapshotConsistencyGroupManager struct {
	createCalls, deleteCalls, getCalls, getByNameCalls int

	createRequest *models.SnapshotConsistencyGroupRequest
	createResult  *models.SnapshotConsistencyGroupResponse
	createErr     error

	deleteGroupID string
	deleteErr     error

	getGroupID string
	getResult  *models.SnapshotConsistencyGroupResponse
	getErr     error

	getByNameName            string
	getByNameResourceGroupID string
	getByNameResult          *models.SnapshotConsistencyGroupResponse
	getByNameErr             error

	listFilters *models.ListSnapshotConsistencyGroupFilters
	listResult  *models.SnapshotConsistencyGroupList
	listErr     error
}

func (fake *fakeSnapshotConsistencyGroupManager) CreateSnapshotConsistencyGroup(req *models.SnapshotConsistencyGroupRequest, logger *zap.Logger) (*models.SnapshotConsistencyGroupResponse, error) {
	fake.createCalls++
	fake.createRequest = req
	return fake.createResult, fake.createErr
}

func (fake *fakeSnapshotConsistencyGroupManager) DeleteSnapshotConsistencyGroup(groupID string, logger *zap.Logger) error {
	fake.deleteCalls++
	fake.deleteGroupID = groupID
	return fake.deleteErr
}

func (fake *fakeSnapshotConsistencyGroupManager) GetSnapshotConsistencyGroup(groupID string, logger *zap.Logger) (*models.SnapshotConsistencyGroupResponse, error) {
	fake.getCalls++
	fake.getGroupID = groupID
	return fake.getResult, fake.getErr
}

func (fake *fakeSnapshotConsistencyGroupManager) GetSnapshotConsistencyGroupByName(name, resourceGroupID string, logger *zap.Logger) (*models.SnapshotConsistencyGroupResponse, error) {
	fake.getByNameCalls++
	fake.getByNameName = name
	fake.getByNameResourceGroupID = resourceGroupID
	return fake.getByNameResult, fake.getByNameErr
}

func (fake *fakeSnapshotConsistencyGroupManager) ListSnapshotConsistencyGroups(limit int, start string, filters *models.ListSnapshotConsistencyGroupFilters, logger *zap.Logger) (*models.SnapshotConsistencyGroupList, error) {
	fake.listFilters = filters
	return fake.listResult, fake.listErr
}

// attempts sets the max num of tries, including the first call: 1 means no retries
// Restore the previous retry settings when the test finishes.
func limitRetries(t *testing.T, attempts int) {
	t.Helper()
	previousAttempts, previousGap := maxRetryAttempt, maxRetryGap
	t.Cleanup(func() { SetRetryParameters(previousAttempts, previousGap) })
	SetRetryParameters(attempts, 1)
}

// requireProviderError asserts that err is a provider msg with the reason code, description, and backend error.
func requireProviderError(t *testing.T, err error, wantCode, wantDescription, wantBackend string) {
	t.Helper()
	require.Error(t, err)
	var userMsg providerError.Message
	require.ErrorAs(t, err, &userMsg)
	assert.Equal(t, wantCode, userMsg.Code)
	assert.Contains(t, userMsg.Description, wantDescription)
	assert.Contains(t, userMsg.BackendError, wantBackend)
}

func TestCreateGroupSnapshotCreatesConsistencyGroupAndListsMembers(t *testing.T) {
	logger, teardown := GetTestLogger(t)
	defer teardown()

	vpcs, uc, _, err := GetTestOpenSession(t, logger)
	require.NoError(t, err)

	now := time.Now()
	groupService := &fakeSnapshotConsistencyGroupManager{
		createResult: &models.SnapshotConsistencyGroupResponse{
			ID:             "group-snapshot-id",
			CRN:            "group-snapshot-crn",
			Href:           "group-snapshot-href",
			LifecycleState: snapshotReadyState,
			CreatedAt:      &now,
			// Match the VPC create response: member references precede full details.
			Snapshots: []models.SnapshotReference{
				{ID: "snapshot-id-1", CRN: "snapshot-crn-1"},
				{ID: "snapshot-id-2", CRN: "snapshot-crn-2"},
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

// groupSnapshotOperations is helper used by Create, Get and GetByName to fetch member snapshot details.
var groupSnapshotOperations = []struct {
	name string
	call func(*VPCSession) (*provider.GroupSnapshot, error)
}{
	{
		name: "create",
		call: func(session *VPCSession) (*provider.GroupSnapshot, error) {
			return session.CreateGroupSnapshot([]string{"volume-id-1", "volume-id-2"}, provider.GroupSnapshotParameters{
				Name: "group-snapshot-name", ResourceGroup: "resource-group-id",
			})
		},
	},
	{
		name: "get by ID",
		call: func(session *VPCSession) (*provider.GroupSnapshot, error) {
			return session.GetGroupSnapshot("group-snapshot-id")
		},
	},
	{
		name: "get by name",
		call: func(session *VPCSession) (*provider.GroupSnapshot, error) {
			return session.GetGroupSnapshotByName("group-snapshot-name", "resource-group-id")
		},
	},
}

// Check how Create, Get and GetByName handle missing member details and backend errors
func TestGroupSnapshotMemberLookup(t *testing.T) {
	fullMemberDetails := &models.SnapshotList{Snapshots: []*models.Snapshot{
		{ID: "snapshot-id-1", CRN: "snapshot-crn-1", LifecycleState: snapshotReadyState, SourceVolume: &models.SourceVolume{ID: "volume-id-1"}},
		{ID: "snapshot-id-2", CRN: "snapshot-crn-2", LifecycleState: snapshotReadyState, SourceVolume: &models.SourceVolume{ID: "volume-id-2"}},
	}}
	backendError := func(code, status string) error {
		return &models.Error{Trace: "member-lookup-trace", Errors: []models.ErrorItem{{Code: models.ErrorCode(code), Status: status}}}
	}
	memberLookups := []struct {
		name           string
		result         *models.SnapshotList
		err            error
		wantReady      bool
		wantIncomplete bool
		wantNilPage    bool
	}{
		{name: "full details", result: fullMemberDetails, wantReady: true},
		{name: "service unavailable", err: backendError("snapshots_service_unavailable", "503 Service Unavailable")},
		{name: "permission denied", err: backendError("snapshots_not_authorized", "403 Forbidden")},
		{name: "authentication failed", err: backendError(string(models.ErrorCodeTokenInvalid), "401 Unauthorized")},
		{name: "rate limited", err: backendError("snapshots_too_many_requests", "429 Too Many Requests")},
		{name: "internal backend error", err: backendError("internal_error", "500 Internal Server Error")},
		{name: "transport error", err: errors.New("snapshot endpoint connection reset")},
		// A nil response is a lookup failure, not a valid page with missing members
		{name: "nil page returned", wantNilPage: true},
		{name: "empty details returned", result: &models.SnapshotList{}, wantIncomplete: true},
		{name: "partial details returned", result: &models.SnapshotList{Snapshots: fullMemberDetails.Snapshots[:1]}, wantIncomplete: true},
		{
			name: "source volume missing from full details", wantIncomplete: true,
			result: &models.SnapshotList{Snapshots: []*models.Snapshot{
				{ID: "snapshot-id-1", CRN: "snapshot-crn-1"},
				fullMemberDetails.Snapshots[1],
			}},
		},
		{
			name: "complete details but snapshots are not ready",
			result: &models.SnapshotList{Snapshots: []*models.Snapshot{
				{ID: "snapshot-id-1", CRN: "snapshot-crn-1", LifecycleState: "pending", SourceVolume: &models.SourceVolume{ID: "volume-id-1"}},
				fullMemberDetails.Snapshots[1],
			}},
		},
	}

	for _, operation := range groupSnapshotOperations {
		for _, lookup := range memberLookups {
			t.Run(operation.name+"/"+lookup.name, func(t *testing.T) {
				logger, teardown := GetTestLogger(t)
				defer teardown()
				session, client, _, err := GetTestOpenSession(t, logger)
				require.NoError(t, err)

				group := &models.SnapshotConsistencyGroupResponse{
					ID: "group-snapshot-id", LifecycleState: snapshotReadyState,
					Snapshots: []models.SnapshotReference{
						{ID: "snapshot-id-1", CRN: "snapshot-crn-1"},
						{ID: "snapshot-id-2", CRN: "snapshot-crn-2"},
					},
				}
				client.SnapshotConsistencyGroupServiceReturns(&fakeSnapshotConsistencyGroupManager{
					createResult: group, getResult: group, getByNameResult: group,
				})
				snapshotService := &serviceFakes.SnapshotManager{}
				snapshotService.ListSnapshotsReturns(lookup.result, lookup.err)
				client.SnapshotServiceReturns(snapshotService)

				response, err := operation.call(session)

				// These responses have no next page, so only one member-list call should happen.
				// If the call fails, return the error without retrying it here.
				require.Equal(t, 1, snapshotService.ListSnapshotsCallCount())
				_, _, filters, _ := snapshotService.ListSnapshotsArgsForCall(0)
				assert.Equal(t, group.ID, filters.SnapshotConsistencyGroupID)

				if lookup.err != nil || lookup.wantIncomplete || lookup.wantNilPage {
					require.Error(t, err)
					assert.Nil(t, response)
					var memberLookupError providerError.Message
					require.ErrorAs(t, err, &memberLookupError)
					switch {
					case lookup.wantIncomplete:
						assert.Equal(t, "GroupSnapshotMemberDetailsUnavailable", memberLookupError.Code)
					case lookup.wantNilPage:
						assert.Equal(t, "GroupSnapshotMemberLookupFailed", memberLookupError.Code)
						assert.Contains(t, memberLookupError.BackendError, "empty snapshot page")
					default:
						assert.Equal(t, "GroupSnapshotMemberLookupFailed", memberLookupError.Code)
						assert.Equal(t, lookup.err.Error(), memberLookupError.BackendError)
					}
					assert.Contains(t, memberLookupError.Description, group.ID)

					// Simulate VPC returning full details on a later Get request. The same group should now be ready.
					snapshotService.ListSnapshotsReturns(fullMemberDetails, nil)
					response, err = session.GetGroupSnapshot(group.ID)
					require.NoError(t, err)
					require.NotNil(t, response)
					assert.Equal(t, group.ID, response.GroupSnapshotID)
					assert.True(t, response.ReadyToUse)
					assert.Equal(t, 2, snapshotService.ListSnapshotsCallCount())
					return
				}

				require.NoError(t, err)
				require.NotNil(t, response)
				assert.Equal(t, group.ID, response.GroupSnapshotID)
				assert.Equal(t, lookup.wantReady, response.ReadyToUse)
				require.Len(t, response.Snapshots, 2)
				assert.Equal(t, "snapshot-id-1", response.Snapshots[0].SnapshotID)
				assert.Equal(t, "snapshot-id-2", response.Snapshots[1].SnapshotID)
				assert.Equal(t, "volume-id-1", response.Snapshots[0].VolumeID)
				assert.Equal(t, "volume-id-2", response.Snapshots[1].VolumeID)
			})
		}
	}
}

func TestFromProviderToLibGroupSnapshotReadiness(t *testing.T) {
	logger, teardown := GetTestLogger(t)
	defer teardown()

	testCases := []struct {
		name            string
		groupState      string
		groupMembers    []models.SnapshotReference
		snapshotDetails []*models.Snapshot
		expectedReady   bool
	}{
		{
			name:       "stable group with all stable members",
			groupState: snapshotReadyState,
			groupMembers: []models.SnapshotReference{
				{ID: "snapshot-id-1"},
				{ID: "snapshot-id-2"},
			},
			snapshotDetails: []*models.Snapshot{
				{ID: "snapshot-id-1", LifecycleState: snapshotReadyState, SourceVolume: &models.SourceVolume{ID: "volume-id-1"}},
				{ID: "snapshot-id-2", LifecycleState: snapshotReadyState, SourceVolume: &models.SourceVolume{ID: "volume-id-2"}},
			},
			expectedReady: true,
		},
		{
			name:       "stable group with a pending member",
			groupState: snapshotReadyState,
			groupMembers: []models.SnapshotReference{
				{ID: "snapshot-id-1"},
				{ID: "snapshot-id-2"},
			},
			snapshotDetails: []*models.Snapshot{
				{ID: "snapshot-id-1", LifecycleState: snapshotReadyState, SourceVolume: &models.SourceVolume{ID: "volume-id-1"}},
				{ID: "snapshot-id-2", LifecycleState: "pending", SourceVolume: &models.SourceVolume{ID: "volume-id-2"}},
			},
			expectedReady: false,
		},
		{
			name:       "pending group with stable members",
			groupState: "pending",
			groupMembers: []models.SnapshotReference{
				{ID: "snapshot-id-1"},
			},
			snapshotDetails: []*models.Snapshot{
				{ID: "snapshot-id-1", LifecycleState: snapshotReadyState, SourceVolume: &models.SourceVolume{ID: "volume-id-1"}},
			},
			expectedReady: false,
		},
		{
			name:       "stable group without full member details",
			groupState: snapshotReadyState,
			groupMembers: []models.SnapshotReference{
				{ID: "snapshot-id-1"},
			},
			snapshotDetails: nil,
			expectedReady:   false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			group := &models.SnapshotConsistencyGroupResponse{
				ID:             "group-snapshot-id",
				LifecycleState: tc.groupState,
				Snapshots:      tc.groupMembers,
			}

			result := FromProviderToLibGroupSnapshot(group, tc.snapshotDetails, logger)

			require.NotNil(t, result)
			assert.Equal(t, tc.expectedReady, result.ReadyToUse)
		})
	}
}

// The group must stay not ready if any member's details are missing
// Once both members have complete details and are ready, the group becomes ready
func TestFromProviderToLibGroupSnapshotPartialMembers(t *testing.T) {
	logger, teardown := GetTestLogger(t)
	defer teardown()

	group := &models.SnapshotConsistencyGroupResponse{
		ID:             "group-snapshot-id",
		LifecycleState: snapshotReadyState,
		Snapshots: []models.SnapshotReference{
			{ID: "snapshot-id-1", CRN: "snapshot-crn-1"},
			{ID: "snapshot-id-2", CRN: "snapshot-crn-2"},
		},
	}

	// The group has two members, but VPC has returned full details for only one
	snapshotDetails := []*models.Snapshot{
		{
			ID:             "snapshot-id-1",
			CRN:            "snapshot-crn-1",
			LifecycleState: snapshotReadyState,
			SourceVolume:   &models.SourceVolume{ID: "volume-id-1"},
		},
	}

	result := FromProviderToLibGroupSnapshot(group, snapshotDetails, logger)

	require.NotNil(t, result)
	require.Len(t, result.Snapshots, 2)
	assert.Equal(t, "volume-id-1", result.Snapshots[0].VolumeID)
	assert.Equal(t, "snapshot-crn-1", result.Snapshots[0].SnapshotCRN)
	assert.True(t, result.Snapshots[0].ReadyToUse)
	assert.Equal(t, "snapshot-id-2", result.Snapshots[1].SnapshotID)
	assert.Empty(t, result.Snapshots[1].VolumeID)
	assert.Equal(t, "snapshot-crn-2", result.Snapshots[1].SnapshotCRN)
	assert.False(t, result.Snapshots[1].ReadyToUse)
	assert.False(t, result.ReadyToUse)

	// Simulate a later response with the second member's details
	// Convert it again and check that both members and the group are ready
	snapshotDetails = append(snapshotDetails, &models.Snapshot{
		ID:             "snapshot-id-2",
		CRN:            "snapshot-crn-2",
		LifecycleState: snapshotReadyState,
		SourceVolume:   &models.SourceVolume{ID: "volume-id-2"},
	})
	result = FromProviderToLibGroupSnapshot(group, snapshotDetails, logger)
	require.NotNil(t, result)
	require.Len(t, result.Snapshots, 2)
	assert.Equal(t, "volume-id-2", result.Snapshots[1].VolumeID)
	assert.True(t, result.Snapshots[1].ReadyToUse)
	assert.True(t, result.ReadyToUse)
}

// Missing source-volume details must not panic or produce a ready snapshot
func TestFromProviderToLibSnapshotNilSourceVolume(t *testing.T) {
	logger, teardown := GetTestLogger(t)
	defer teardown()

	for _, tc := range []struct {
		name         string
		sourceVolume *models.SourceVolume
	}{
		{name: "nil source volume"},
		{name: "empty source volume ID", sourceVolume: &models.SourceVolume{}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			result := FromProviderToLibSnapshot(&models.Snapshot{
				ID:             "snapshot-id-1",
				CRN:            "snapshot-crn-1",
				LifecycleState: snapshotReadyState,
				SourceVolume:   tc.sourceVolume,
			}, logger)

			require.NotNil(t, result)
			assert.Empty(t, result.VolumeID)
			assert.False(t, result.ReadyToUse)
		})
	}
}

// Deleting an already missing group must return its backend error after one attempt
func TestGroupNotFoundSkipsRetry(t *testing.T) {
	logger, teardown := GetTestLogger(t)
	defer teardown()

	previousAttempts, previousGap := maxRetryAttempt, maxRetryGap
	t.Cleanup(func() {
		SetRetryParameters(previousAttempts, previousGap)
	})
	SetRetryParameters(2, 1)

	notFoundErr := &models.Error{
		Errors: []models.ErrorItem{
			{Code: models.ErrorCode("snapshot_consistency_groups_not_found")},
		},
	}

	attempts := 0
	err := retry(logger, func() error {
		attempts++
		return notFoundErr
	})

	assert.Equal(t, 1, attempts, "retry must stop after the first attempt for a skip-listed error code")
	assert.Same(t, notFoundErr, err)
}

// If the create call fails, return FailedToCreateGroupSnapshot and don't try to fetch member snapshot details
func TestCreateGroupSnapshotWrapsBackendError(t *testing.T) {
	logger, teardown := GetTestLogger(t)
	defer teardown()

	vpcs, uc, _, err := GetTestOpenSession(t, logger)
	require.NoError(t, err)
	limitRetries(t, 1)

	groupService := &fakeSnapshotConsistencyGroupManager{createErr: &models.Error{Errors: []models.ErrorItem{
		{Code: models.ErrorCode("snapshots_source_volume_not_found"), Status: "404 Not Found"},
	}}}
	snapshotService := &serviceFakes.SnapshotManager{}
	uc.SnapshotConsistencyGroupServiceReturns(groupService)
	uc.SnapshotServiceReturns(snapshotService)

	result, err := vpcs.CreateGroupSnapshot([]string{"volume-id-1"}, provider.GroupSnapshotParameters{
		Name: "group-snapshot-name", ResourceGroup: "resource-group-id",
	})

	assert.Nil(t, result)
	requireProviderError(t, err, "FailedToCreateGroupSnapshot", "", "snapshots_source_volume_not_found")
	assert.Equal(t, 1, groupService.createCalls)
	assert.Equal(t, 0, snapshotService.ListSnapshotsCallCount())
}
