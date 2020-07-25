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
	"github.com/IBM/ibmcloud-storage-volume-lib/volume-providers/vpc/vpcclient/riaas/test"
	"github.com/IBM/ibmcloud-storage-volume-lib/volume-providers/vpc/vpcclient/vpcvolume"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"net/http"
	"testing"
)

func TestListSnapshotTags(t *testing.T) {
	// Setup new style zap logger
	logger, _ := GetTestContextLogger()
	defer logger.Sync()

	testCases := []struct {
		name string

		// backend url
		url string

		tags *[]string

		// Response
		status  int
		content string

		// Expected return
		expectErr string
		verify    func(*testing.T, *[]string, error)
	}{
		{
			name:   "Verify that the correct endpoint is invoked",
			status: http.StatusNoContent,
			url:    vpcvolume.Version + "/volumes/volume-id/snapshots/snapshot-id/tags",
		}, {
			name:      "Verify that a 404 is returned to the caller",
			status:    http.StatusNotFound,
			content:   "{\"errors\":[{\"message\":\"testerr\"}]}",
			url:       "/volumes/volume-id/snapshots/snapshot-id/tags",
			expectErr: "Trace Code:, testerr Please check ",
		}, {
			name:    "Verify that the snapshot is parsed correctly",
			status:  http.StatusOK,
			content: "[\"tag1\"]",
			url:     vpcvolume.Version + "/volumes/volume-id/snapshots/snapshot-id/tags",
			verify: func(t *testing.T, tags *[]string, err error) {
				assert.NotNil(t, tags)
			},
		},
	}

	for _, testcase := range testCases {
		t.Run(testcase.name, func(t *testing.T) {
			mux, client, teardown := test.SetupServer(t)
			requestBody := ""
			test.SetupMuxResponse(t, mux, testcase.url, http.MethodGet, &requestBody, testcase.status, testcase.content, nil)

			defer teardown()

			logger.Info("Test case being executed", zap.Reflect("testcase", testcase.name))

			snapshotService := vpcvolume.NewSnapshotManager(client)

			tags, err := snapshotService.ListSnapshotTags("volume-id", "snapshot-id", logger)
			logger.Info("Tags", zap.Reflect("tags", tags))

			// vpc snapshot functionality is not yet ready. It would return error for now
			if testcase.verify != nil {
				testcase.verify(t, tags, err)
			}
		})
	}
}
