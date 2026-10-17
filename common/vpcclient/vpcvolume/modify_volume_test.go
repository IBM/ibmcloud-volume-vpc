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

// Package vpcvolume ...
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

func TestModifyVolume(t *testing.T) {
	logger, _ := GetTestContextLogger()
	defer logger.Sync()

	testCases := []struct {
		name     string
		template *models.Volume
		// Response
		status  int
		content string
		// Expected
		expectErr string
		verify    func(*testing.T, *models.Volume, error)
	}{
		{
			name:     "Successfully modify volume IOPS",
			template: &models.Volume{Iops: 5000},
			status:   http.StatusOK,
			content:  `{"id":"volume-id","name":"volume-name","capacity":100,"iops":5000,"status":"updating","zone":{"name":"test-1","href":"https://us-south.iaas.cloud.ibm.com/v1/regions/us-south/zones/test-1"},"crn":"crn:v1:bluemix:public:is:test-1:a/rg1::volume:vol1"}`,
			verify: func(t *testing.T, volume *models.Volume, err error) {
				if assert.NotNil(t, volume) {
					assert.Equal(t, "volume-id", volume.ID)
					assert.Equal(t, int64(5000), volume.Iops)
				}
			},
		},
		{
			name:     "Successfully modify volume bandwidth",
			template: &models.Volume{Bandwidth: 1000},
			status:   http.StatusOK,
			content:  `{"id":"volume-id","name":"volume-name","capacity":100,"iops":3000,"bandwidth":1000,"status":"updating","zone":{"name":"test-1","href":"https://us-south.iaas.cloud.ibm.com/v1/regions/us-south/zones/test-1"},"crn":"crn:v1:bluemix:public:is:test-1:a/rg1::volume:vol1"}`,
			verify: func(t *testing.T, volume *models.Volume, err error) {
				if assert.NotNil(t, volume) {
					assert.Equal(t, "volume-id", volume.ID)
					assert.Equal(t, int32(1000), volume.Bandwidth)
				}
			},
		},
		{
			name:      "Verify that a 404 is returned to the caller",
			template:  &models.Volume{Iops: 5000},
			status:    http.StatusNotFound,
			content:   `{"errors":[{"message":"testerr","Code":"volume_not_found"}],"trace":"2af63776-4df7-4970-b52d-4e25676ec0e4"}`,
			expectErr: "Trace Code:2af63776-4df7-4970-b52d-4e25676ec0e4, Code:volume_not_found, Description:testerr, RC:404 Not Found",
		},
		{
			name:      "Verify that a 400 bad request returns an error",
			template:  &models.Volume{Iops: -1},
			status:    http.StatusBadRequest,
			content:   `{"errors":[{"message":"invalid iops value","Code":"invalid_field_value"}],"trace":"abc-trace-123"}`,
			expectErr: "Trace Code:abc-trace-123, Code:invalid_field_value, Description:invalid iops value, RC:400 Bad Request",
		},
		{
			name:     "Verify that the correct endpoint is invoked",
			template: &models.Volume{Iops: 3000},
			status:   http.StatusNoContent,
		},
	}

	for _, testcase := range testCases {
		t.Run(testcase.name, func(t *testing.T) {
			mux, client, teardown := test.SetupServer(t)
			test.SetupMuxResponse(t, mux, vpcvolume.Version+"/volumes/volume-id", http.MethodPatch, nil, testcase.status, testcase.content, nil)

			defer teardown()

			logger.Info("Test case being executed", zap.Reflect("testcase", testcase.name))

			volumeService := vpcvolume.New(client)

			volume, err := volumeService.ModifyVolume("volume-id", testcase.template, logger)
			logger.Info("Volume details", zap.Reflect("volume", volume))

			if testcase.expectErr != "" && assert.Error(t, err) {
				assert.Equal(t, testcase.expectErr, err.Error())
				assert.Nil(t, volume)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, volume)
			}

			if testcase.verify != nil {
				testcase.verify(t, volume, err)
			}
		})
	}
}
