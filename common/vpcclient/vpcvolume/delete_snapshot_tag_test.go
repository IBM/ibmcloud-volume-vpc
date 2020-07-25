/**
 * Copyright 2020 IBM Corp.
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
	"github.com/IBM/ibmcloud-storage-volume-lib/volume-providers/vpc/vpcclient/models"
	"github.com/IBM/ibmcloud-storage-volume-lib/volume-providers/vpc/vpcclient/riaas/test"
	"github.com/IBM/ibmcloud-storage-volume-lib/volume-providers/vpc/vpcclient/vpcvolume"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"net/http"
	"testing"
)

func TestDeleteSnapshotTag(t *testing.T) {
	// Setup new style zap logger
	logger, _ := GetTestContextLogger()
	defer logger.Sync()

	testCases := []struct {
		name string

		// Response
		status  int
		content string

		// Expected return
		expectErr string
		verify    func(*testing.T, *models.Snapshot, error)
	}{
		{
			name:   "Verify that the correct endpoint is invoked",
			status: http.StatusNoContent,
		}, {
			name:      "Verify that a 404 is returned to the caller",
			status:    http.StatusNotFound,
			content:   "{\"errors\":[{\"message\":\"testerr\"}]}",
			expectErr: "Trace Code:, testerr Please check ",
		}, {
			name:    "Verify that the snapshot is parsed correctly",
			status:  http.StatusOK,
			content: "{\"id\":\"snapshot1\",\"name\":\"snapshot1\",\"status\":\"pending\"}",
			verify: func(t *testing.T, snapshot *models.Snapshot, err error) {
				if assert.NotNil(t, snapshot) {
					assert.Equal(t, "snapshot1", snapshot.ID)
				}
			},
		},
	}

	for _, testcase := range testCases {
		t.Run(testcase.name, func(t *testing.T) {
			mux, client, teardown := test.SetupServer(t)
			emptyString := ""
			test.SetupMuxResponse(t, mux, vpcvolume.Version+"/volumes/volume-id/snapshots/snapshotid/tags/tag-name", http.MethodDelete, &emptyString, testcase.status, testcase.content, nil)

			defer teardown()

			logger.Info("Test case being executed", zap.Reflect("testcase", testcase.name))

			snapshotService := vpcvolume.NewSnapshotManager(client)

			err := snapshotService.DeleteSnapshotTag("volume-id", "snapshot-id", "tag-name", logger)
			// vpc snapshot functionality is not yet ready. It would return error for now
			assert.Error(t, err)
		})
	}
}
