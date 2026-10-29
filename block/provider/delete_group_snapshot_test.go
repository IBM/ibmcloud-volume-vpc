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

	"github.com/IBM/ibmcloud-volume-vpc/common/vpcclient/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
				{Code: models.ErrorCode("snapshot_consistency_groups_not_found")},
			}},
			maxRetries: 3,
			want: want{
				errCode:      "FailedToDeleteGroupSnapshot",
				descContain:  "group-snapshot-id",
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

			groupService := &fakeSnapshotConsistencyGroupManager{deleteErr: tc.deleteErr}
			uc.SnapshotConsistencyGroupServiceReturns(groupService)

			err = vpcs.DeleteGroupSnapshot(tc.groupID)

			assert.Equal(t, tc.want.backendCalls, groupService.deleteCalls)
			if tc.want.errCode == "" {
				require.NoError(t, err)
				assert.Equal(t, tc.groupID, groupService.deleteGroupID)
				return
			}
			requireProviderError(t, err, tc.want.errCode, tc.want.descContain, tc.want.backContain)
		})
	}
}
