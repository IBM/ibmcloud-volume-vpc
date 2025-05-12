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

	"github.com/IBM/ibmcloud-volume-vpc/common/vpcclient/models"
	"github.com/IBM/ibmcloud-volume-vpc/common/vpcclient/riaas/test"
	"github.com/IBM/ibmcloud-volume-vpc/common/vpcclient/vpcvolume"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestSetVolumeTag(t *testing.T) {
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
		verify    func(*testing.T, *models.Volume, error)
	}{
		{
			name:   "Verify that the correct endpoint is invoked",
			status: http.StatusNoContent,
			url:    vpcvolume.Version + "/volumes/volumeid/tags/tagname",
		}, {
			name:      "Verify that a 404 is returned to the caller",
			status:    http.StatusNotFound,
			url:       vpcvolume.Version + "/volumes/volumeid/tags/tagname",
			content:   "{\"errors\":[{\"message\":\"testerr\",\"Code\":\"volume_not_found\"}], \"trace\":\"2af63776-4df7-4970-b52d-4e25676ec0e4\"}",
			expectErr: "Trace Code:2af63776-4df7-4970-b52d-4e25676ec0e4, Code:volume_not_found, Description:testerr, RC:404 Not Found",
		}, {
			name:    "Verify that the volume is parsed correctly",
			status:  http.StatusOK,
			url:     vpcvolume.Version + "/volumes/volumeid/tags/tagname",
			content: "{\"id\":\"volumeid\",\"name\":\"volume-name\",\"capacity\":10,\"iops\":3000,\"status\":\"pending\",\"zone\":{\"name\":\"test-1\",\"href\":\"https://us-south.iaas.cloud.ibm.com/v1/regions/us-south/zones/test-1\"},\"crn\":\"crn:v1:bluemix:public:is:test-1:a/rg1::volume:vol1\"}",
		}, {
			name:    "False positive: What if the volume ID is not matched",
			status:  http.StatusOK,
			url:     vpcvolume.Version + "/volumes/volumeid/tags/tagname",
			content: "{\"id\":\"wrong-vol\",\"name\":\"wrong-vol\",\"capacity\":10,\"iops\":3000,\"status\":\"pending\",\"zone\":{\"name\":\"test-1\",\"href\":\"https://us-south.iaas.cloud.ibm.com/v1/regions/us-south/zones/test-1\"},\"crn\":\"crn:v1:bluemix:public:is:test-1:a/rg1::volume:wrong-vol\", \"tags\":[\"Wrong Tag\"]}",
		}, {
			name:    "False positive: What if the tag name is not matched",
			status:  http.StatusOK,
			url:     vpcvolume.Version + "/volumes/volumeid/tags/tagname",
			content: "{\"id\":\"volumeid\",\"name\":\"volume-name\",\"capacity\":10,\"iops\":3000,\"status\":\"pending\",\"zone\":{\"name\":\"test-1\",\"href\":\"https://us-south.iaas.cloud.ibm.com/v1/regions/us-south/zones/test-1\"},\"crn\":\"crn:v1:bluemix:public:is:test-1:a/rg1::volume:vol1\", \"tags\":[\"Test Tag\"]}",
		}, {
			name:    "False positive: What if the tag name is already set",
			status:  http.StatusOK,
			url:     vpcvolume.Version + "/volumes/volumeid/tags/tagname",
			content: "{\"id\":\"volumeid\",\"name\":\"volume-name\",\"capacity\":10,\"iops\":3000,\"status\":\"pending\",\"zone\":{\"name\":\"test-1\",\"href\":\"https://us-south.iaas.cloud.ibm.com/v1/regions/us-south/zones/test-1\"},\"crn\":\"crn:v1:bluemix:public:is:test-1:a/rg1::volume:vol1\", \"tags\":[\"tagname\"]}",
		},
	}

	for _, testcase := range testCases {
		t.Run(testcase.name, func(t *testing.T) {
			mux, client, teardown := test.SetupServer(t)
			test.SetupMuxResponse(t, mux, testcase.url, http.MethodPut, nil, testcase.status, testcase.content, nil)

			defer teardown()

			logger.Info("Test case being executed", zap.Reflect("testcase", testcase.name))

			volumeService := vpcvolume.New(client)

			err := volumeService.SetVolumeTag("volumeid", "tagname", logger)

			if testcase.expectErr != "" && assert.Error(t, err) {
				assert.Equal(t, testcase.expectErr, err.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
