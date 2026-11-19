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
	"context"
	"errors"
	"fmt"
	"testing"
	"testing/synctest"
	"time"

	"github.com/IBM/ibmcloud-volume-vpc/common/vpcclient/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestDeleteGroupSnapshot(t *testing.T) {
	type want struct {
		errCode      string
		descContain  string
		backContain  string
		backendCalls int
	}
	testCases := []struct {
		name       string
		groupID    string
		deleteErr  error
		maxRetries int // 0 means use defaults; >0 sets the total attempts, including the first call.
		want       want
	}{
		{
			name:    "success: group ID forwarded to backend",
			groupID: "group-snapshot-id",
			want:    want{backendCalls: 1},
		},
		{
			name:    "backend error is wrapped as FailedToDeleteGroupSnapshot",
			groupID: "group-snapshot-id",
			deleteErr: &models.Error{Errors: []models.ErrorItem{
				{Code: "internal_error", Status: "500 Internal Server Error"},
			}},
			maxRetries: 1,
			want: want{
				errCode:      "FailedToDeleteGroupSnapshot",
				descContain:  "group-snapshot-id",
				backContain:  "internal_error",
				backendCalls: 1,
			},
		},
		{
			name:    "not-found error is not retried",
			groupID: "group-snapshot-id",
			deleteErr: &models.Error{Errors: []models.ErrorItem{
				{Code: SnapshotConsistencyGroupNotFound},
			}},
			maxRetries: 3,
			want: want{
				errCode:      "FailedToDeleteGroupSnapshot",
				descContain:  "group-snapshot-id",
				backContain:  SnapshotConsistencyGroupNotFound,
				backendCalls: 1,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			logger, teardown := GetTestLogger(t)
			defer teardown()

			vpcs, uc, _, err := GetTestOpenSession(t, logger)
			require.NoError(t, err)

			// Must run after the session opens: opening it resets the retry settings from config.
			if tc.maxRetries > 0 {
				limitRetries(t, tc.maxRetries)
			}

			groupService := &fakeSnapshotConsistencyGroupManager{
				deleteErr: tc.deleteErr,
				getErr: &models.Error{Errors: []models.ErrorItem{
					{Code: SnapshotConsistencyGroupNotFound},
				}},
			}
			uc.SnapshotConsistencyGroupServiceReturns(groupService)

			err = vpcs.DeleteGroupSnapshot(tc.groupID)

			assert.Equal(t, tc.want.backendCalls, groupService.deleteCalls)
			assert.Zero(t, uc.SnapshotServiceCallCount())
			if tc.want.errCode == "" {
				require.NoError(t, err)
				assert.Equal(t, tc.groupID, groupService.deleteGroupID)
				assert.Equal(t, 1, groupService.getCalls)
				assert.Equal(t, tc.groupID, groupService.getGroupID)
				return
			}
			assert.Zero(t, groupService.getCalls)
			requireProviderError(t, err, tc.want.errCode, tc.want.descContain, tc.want.backContain)
		})
	}
}

// A successful DELETE must be followed by group not-found before we report success.
func TestDeleteGroupSnapshotWaitsForDeletion(t *testing.T) {
	type pollResult struct {
		group *models.SnapshotConsistencyGroupResponse
		err   error
	}
	backendError := func(code string) error {
		return &models.Error{Errors: []models.ErrorItem{{Code: models.ErrorCode(code)}}}
	}
	deletingGroup := &models.SnapshotConsistencyGroupResponse{
		ID: "group-snapshot-id", LifecycleState: "deleting",
	}
	notFound := backendError(SnapshotConsistencyGroupNotFound)
	testCases := []struct {
		name         string
		polls        []pollResult
		attempts     int
		wantGetCalls int
		wantError    string
		wantDelay    time.Duration
	}{
		{
			name:         "group is already gone when polling starts",
			polls:        []pollResult{{err: notFound}},
			attempts:     3,
			wantGetCalls: 1,
		},
		{
			name: "wait until deleting group is gone without repeating DELETE",
			polls: []pollResult{
				{group: deletingGroup},
				{group: deletingGroup},
				{err: notFound},
			},
			attempts:     3,
			wantGetCalls: 3,
			wantDelay:    30 * time.Second,
		},
		{
			name:         "wrapped group not-found also confirms deletion",
			polls:        []pollResult{{err: fmt.Errorf("group lookup failed: %w", notFound)}},
			attempts:     3,
			wantGetCalls: 1,
		},
		{
			name:         "group still exists after every check",
			polls:        []pollResult{{group: deletingGroup}},
			attempts:     3,
			wantGetCalls: 3,
			wantError:    "still exists after 3 retry attempts",
			wantDelay:    30 * time.Second,
		},
		{
			name: "stable group is not mistaken for a deleted group",
			polls: []pollResult{{group: &models.SnapshotConsistencyGroupResponse{
				ID: "group-snapshot-id", LifecycleState: snapshotReadyState,
			}}},
			attempts:     1,
			wantGetCalls: 1,
			wantError:    "still exists after 1 retry attempts",
		},
		{
			name:         "permission failure is not deletion success",
			polls:        []pollResult{{err: backendError("snapshots_not_authorized")}},
			attempts:     3,
			wantGetCalls: 1,
			wantError:    "snapshots_not_authorized",
		},
		{
			name:         "service unavailable is not deletion success",
			polls:        []pollResult{{err: backendError("snapshots_service_unavailable")}},
			attempts:     3,
			wantGetCalls: 1,
			wantError:    "snapshots_service_unavailable",
		},
		{
			name:         "rate limiting is not deletion success",
			polls:        []pollResult{{err: backendError("snapshots_too_many_requests")}},
			attempts:     3,
			wantGetCalls: 1,
			wantError:    "snapshots_too_many_requests",
		},
		{
			name:         "individual snapshot not-found does not confirm group deletion",
			polls:        []pollResult{{err: backendError("snapshots_not_found")}},
			attempts:     3,
			wantGetCalls: 1,
			wantError:    "snapshots_not_found",
		},
		{
			name:         "generic not-found does not confirm group deletion",
			polls:        []pollResult{{err: backendError("not_found")}},
			attempts:     3,
			wantGetCalls: 1,
			wantError:    "Code:not_found",
		},
		{
			name:         "transport error is returned without a type assertion panic",
			polls:        []pollResult{{err: errors.New("connection reset by peer")}},
			attempts:     3,
			wantGetCalls: 1,
			wantError:    "connection reset by peer",
		},
		{
			name:         "request timeout is not deletion success",
			polls:        []pollResult{{err: context.DeadlineExceeded}},
			attempts:     3,
			wantGetCalls: 1,
			wantError:    "context deadline exceeded",
		},
		{
			name:         "empty successful response does not confirm deletion",
			polls:        []pollResult{{}},
			attempts:     1,
			wantGetCalls: 1,
			wantError:    "still exists after 1 retry attempts",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			logger, teardown := GetTestLogger(t)
			defer teardown()
			vpcs, uc, _, err := GetTestOpenSession(t, logger)
			require.NoError(t, err)
			vpcs.APIRetry = NewFlexyRetry(tc.attempts, 60)

			groupService := &fakeSnapshotConsistencyGroupManager{}
			groupService.getStub = func(_ string, _ *zap.Logger) (*models.SnapshotConsistencyGroupResponse, error) {
				poll := tc.polls[min(groupService.getCalls-1, len(tc.polls)-1)]
				return poll.group, poll.err
			}
			uc.SnapshotConsistencyGroupServiceReturns(groupService)

			var delay time.Duration
			// Use fake time so the polling delays do not slow down the tests.
			synctest.Test(t, func(t *testing.T) {
				start := time.Now()
				err = vpcs.DeleteGroupSnapshot("group-snapshot-id")
				delay = time.Since(start)
			})

			assert.Equal(t, 1, groupService.deleteCalls)
			assert.Equal(t, "group-snapshot-id", groupService.deleteGroupID)
			assert.Equal(t, tc.wantGetCalls, groupService.getCalls)
			assert.Equal(t, "group-snapshot-id", groupService.getGroupID)
			assert.Equal(t, tc.wantDelay, delay)
			assert.Zero(t, uc.SnapshotServiceCallCount())
			if tc.wantError == "" {
				require.NoError(t, err)
				return
			}
			requireProviderError(t, err, "FailedToDeleteGroupSnapshot", "group-snapshot-id", tc.wantError)
		})
	}
}
