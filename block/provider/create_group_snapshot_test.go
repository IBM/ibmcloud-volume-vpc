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

	"github.com/IBM/ibmcloud-volume-interface/lib/provider"
	providerError "github.com/IBM/ibmcloud-volume-interface/lib/utils"
	userError "github.com/IBM/ibmcloud-volume-vpc/common/messages"
	"github.com/IBM/ibmcloud-volume-vpc/common/vpcclient/models"
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
	getStub    func(string, *zap.Logger) (*models.SnapshotConsistencyGroupResponse, error)

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
	if fake.getStub != nil {
		return fake.getStub(groupID, logger)
	}
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

func TestCreateGroupSnapshotReturnsGroupAndMembers(t *testing.T) {
	logger, teardown := GetTestLogger(t)
	defer teardown()

	vpcs, uc, _, err := GetTestOpenSession(t, logger)
	require.NoError(t, err)

	group := newGroupResponse(snapshotReadyState, snapshotReadyState)
	groupService := &fakeSnapshotConsistencyGroupManager{createResult: group}
	uc.SnapshotConsistencyGroupServiceReturns(groupService)

	groupSnapshot, err := vpcs.CreateGroupSnapshot([]string{"volume-id-1", "volume-id-2"}, provider.GroupSnapshotParameters{
		Name:          "group-snapshot-name",
		ResourceGroup: "resource-group-id",
	})

	require.NoError(t, err)
	require.NotNil(t, groupSnapshot)
	assert.Equal(t, "group-snapshot-id", groupSnapshot.GroupSnapshotID)
	assert.Equal(t, "group-snapshot-crn", groupSnapshot.GroupSnapshotCRN)
	assert.Equal(t, "group-snapshot-href", groupSnapshot.GroupSnapshotMetadata.Href)
	assert.Equal(t, *group.CreatedAt, groupSnapshot.GroupSnapshotCreationTime)
	assert.True(t, groupSnapshot.ReadyToUse)
	require.Len(t, groupSnapshot.Snapshots, 2)
	assert.Equal(t, &provider.Snapshot{
		SnapshotID:           "snapshot-id-1",
		SnapshotCRN:          "snapshot-crn-1",
		SnapshotCreationTime: *group.Snapshots[0].CreatedAt,
		SnapshotSize:         10 * GiB,
		VPC:                  provider.VPC{Href: "snapshot-href-1"},
		VolumeID:             "volume-id-1",
		GroupSnapshotID:      group.ID,
		GroupSnapshotCRN:     group.CRN,
		ReadyToUse:           true,
	}, groupSnapshot.Snapshots[0])
	assert.Equal(t, "volume-id-2", groupSnapshot.Snapshots[1].VolumeID)
	assert.EqualValues(t, 20*GiB, groupSnapshot.Snapshots[1].SnapshotSize)
	assert.Equal(t, 1, groupService.createCalls)
	assert.Equal(t, "group-snapshot-name", groupService.createRequest.Name)
	assert.Equal(t, "resource-group-id", groupService.createRequest.ResourceGroup.ID)
	assert.True(t, groupService.createRequest.DeleteSnapshotsOnDelete)
	require.Len(t, groupService.createRequest.Snapshots, 2)
	assert.Equal(t, "volume-id-1", groupService.createRequest.Snapshots[0].SourceVolume.ID)
	assert.Equal(t, "volume-id-2", groupService.createRequest.Snapshots[1].SourceVolume.ID)
	assert.Zero(t, uc.SnapshotServiceCallCount())
}

// Run shared response checks through Create, Get, and GetByName.
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
			{Code: SnapshotConsistencyGroupNotFound},
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

// A failed VPC create returns the original backend error in the provider message.
func TestCreateGroupSnapshotWrapsBackendError(t *testing.T) {
	logger, teardown := GetTestLogger(t)
	defer teardown()

	vpcs, uc, _, err := GetTestOpenSession(t, logger)
	require.NoError(t, err)
	limitRetries(t, 1)

	groupService := &fakeSnapshotConsistencyGroupManager{createErr: &models.Error{Errors: []models.ErrorItem{
		{Code: models.ErrorCode("snapshots_source_volume_not_found"), Status: "404 Not Found"},
	}}}
	uc.SnapshotConsistencyGroupServiceReturns(groupService)

	result, err := vpcs.CreateGroupSnapshot([]string{"volume-id-1"}, provider.GroupSnapshotParameters{
		Name: "group-snapshot-name", ResourceGroup: "resource-group-id",
	})

	assert.Nil(t, result)
	requireProviderError(t, err, "FailedToCreateGroupSnapshot", "", "snapshots_source_volume_not_found")
	assert.Equal(t, 1, groupService.createCalls)
	assert.Zero(t, uc.SnapshotServiceCallCount())
}
