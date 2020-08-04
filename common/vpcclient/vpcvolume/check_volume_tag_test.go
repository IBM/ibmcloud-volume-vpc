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

// Package vpcvolume_test ...
package vpcvolume_test

import (
	"net/http"
	"testing"

	"github.com/IBM/ibmcloud-volume-vpc/common/vpcclient/riaas/test"
	"github.com/IBM/ibmcloud-volume-vpc/common/vpcclient/vpcvolume"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestCheckVolumeTag(t *testing.T) {
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
		verify    func(*testing.T, error)
	}{
		{
			name:   "Verify that the correct endpoint is invoked",
			url:    "/v1/volumes/volume-id/tags/tag-name",
			status: http.StatusNoContent,
		}, {
			name:      "Verify that a 404 is returned to the caller",
			url:       "/v1/volumes/volume-id/tags/tag-name",
			status:    http.StatusNotFound,
			content:   "{\"errors\":[{\"message\":\"testerr\"}]}",
			expectErr: "Trace Code:, testerr Please check ",
		}, {
			name:    "Verify that the volume is parsed correctly and has correct tag name",
			url:     "/v1/volumes/volume-id/tags/tag-name",
			status:  http.StatusOK,
			content: "{\"id\":\"volume-id\",\"name\":\"volume-name\",\"capacity\":10,\"iops\":3000,\"status\":\"pending\",\"zone\":{\"name\":\"test-1\",\"href\":\"https://us-south.iaas.cloud.ibm.com/v1/regions/us-south/zones/test-1\"},\"crn\":\"crn:v1:bluemix:public:is:test-1:a/rg1::volume:vol1\", \"tags\":[\"tag-name\"]}",
		}, {
			name:      "False positive: What if the volume ID is not matched",
			url:       "/v1/volumes/wrong-volume-id/tags/tag-name",
			status:    http.StatusNotFound,
			content:   "{\"id\":\"wrong-vol\",\"name\":\"wrong-vol\",\"capacity\":10,\"iops\":3000,\"status\":\"pending\",\"zone\":{\"name\":\"test-1\",\"href\":\"https://us-south.iaas.cloud.ibm.com/v1/regions/us-south/zones/test-1\"},\"crn\":\"crn:v1:bluemix:public:is:test-1:a/rg1::volume:wrong-vol\", \"tags\":[\"Wrong Tag\"]}",
			expectErr: "json: cannot unmarshal number into Go value of type models.Error",
		}, {
			name:    "False positive: What if the tag name is not matched",
			url:     "/v1/volumes/volume-id/tags/tag-name",
			status:  http.StatusOK,
			content: "{\"id\":\"volume-id\",\"name\":\"volume-name\",\"capacity\":10,\"iops\":3000,\"status\":\"pending\",\"zone\":{\"name\":\"test-1\",\"href\":\"https://us-south.iaas.cloud.ibm.com/v1/regions/us-south/zones/test-1\"},\"crn\":\"crn:v1:bluemix:public:is:test-1:a/rg1::volume:vol1\", \"tags\":[\"Test Tag\"]}",
		},
	}

	for _, testcase := range testCases {
		t.Run(testcase.name, func(t *testing.T) {
			mux, client, teardown := test.SetupServer(t)
			test.SetupMuxResponse(t, mux, testcase.url, http.MethodGet, nil, testcase.status, testcase.content, nil)

			defer teardown()

			volumeService := vpcvolume.New(client)

			logger.Info("Test case being executed", zap.Reflect("testcase", testcase.name))

			err := volumeService.CheckVolumeTag("volume-id", "tag-name", logger)

			if testcase.expectErr != "" && assert.Error(t, err) {
				assert.Equal(t, testcase.expectErr, err.Error())
			} else {
				assert.NoError(t, err)
			}

			if testcase.verify != nil {
				testcase.verify(t, err)
			}
		})
	}
}
