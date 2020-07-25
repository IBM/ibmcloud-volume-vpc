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
	"strings"
	"testing"
)

func TestCreateSnapshot(t *testing.T) {
	// Setup new style zap logger
	logger, _ := GetTestContextLogger()
	defer logger.Sync()

	testCases := []struct {
		name string

		// backend url
		url string

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
			url:    "/volumes/volume-id/snapshots",
		}, {
			name:      "Verify that a 404 is returned to the caller",
			status:    http.StatusNotFound,
			url:       vpcvolume.Version + "/volumes/volume-id/snapshots",
			content:   "{\"errors\":[{\"message\":\"testerr\"}]}",
			expectErr: "Trace Code:, testerr Please check ",
		}, {
			name:    "Verify that the snapshot is parsed correctly",
			status:  http.StatusOK,
			url:     vpcvolume.Version + "/volumes/volume-id/snapshots",
			content: "{\"id\":\"snapshot1\",\"status\":\"pending\"}",
			verify: func(t *testing.T, snapshot *models.Snapshot, err error) {
				assert.NotNil(t, snapshot)
			},
		}, {
			name:    "Verify that the snapshot is parsed correctly",
			status:  http.StatusOK,
			content: "{\"id\":\"snapshot1\",\"status\":\"pending\"}",
			url:     vpcvolume.Version + "/volumes/volume-id/snapshots",
			verify: func(t *testing.T, snapshot *models.Snapshot, err error) {
				assert.NotNil(t, snapshot)
			},
		},
	}

	for _, testcase := range testCases {
		t.Run(testcase.name, func(t *testing.T) {
			template := &models.Snapshot{
				Name: "snapshot-name",
				ID:   "snapshot-id",
			}
			mux, client, teardown := test.SetupServer(t)
			requestBody := `{
        			"id":"snapshot-id",
  			        "name":"snapshot-name"
      			}`
			requestBody = strings.Join(strings.Fields(requestBody), "") + "\n"
			test.SetupMuxResponse(t, mux, testcase.url, http.MethodPost, &requestBody, testcase.status, testcase.content, nil)

			defer teardown()

			logger.Info("Test case being executed", zap.Reflect("testcase", testcase.name))

			snapshotService := vpcvolume.NewSnapshotManager(client)

			snapshot, err := snapshotService.CreateSnapshot("volume-id", template, logger)
			logger.Info("Snapshot", zap.Reflect("snapshot", snapshot))

			// vpc snapshot functionality is not yet ready. It would return error for now
			if testcase.verify != nil {
				testcase.verify(t, snapshot, err)
			}
		})
	}
}
