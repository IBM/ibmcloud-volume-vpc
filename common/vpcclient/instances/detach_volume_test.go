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

// Package instances_test ...
package instances_test

import (
	"net/http"
	"testing"

	"github.com/IBM/ibmcloud-volume-vpc/common/vpcclient/instances"
	"github.com/IBM/ibmcloud-volume-vpc/common/vpcclient/models"
	"github.com/IBM/ibmcloud-volume-vpc/common/vpcclient/riaas/test"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestDetachVolume(t *testing.T) {
	// Setup new style zap logger
	logger, _ := GetTestContextLogger()
	defer logger.Sync()

	instanceID := "testinstance"

	testCases := []struct {
		name string

		// Response
		status  int
		content string

		// Expected return
		expectErr string
		verify    func(*testing.T, *http.Response, error)
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
			name:   "Verify that the volume detachment is done correctly",
			status: http.StatusOK,
			verify: func(t *testing.T, httpResponse *http.Response, err error) {
				if assert.Nil(t, err) {
					assert.Nil(t, httpResponse)
				}
			},
		},
	}

	for _, testcase := range testCases {
		t.Run(testcase.name, func(t *testing.T) {
			template := &models.VolumeAttachment{
				ID:         "volumeattachmentid",
				Name:       "volumeattachment",
				InstanceID: &instanceID,
				Volume: &models.Volume{
					ID:       "volume-id",
					Name:     "volume-name",
					Capacity: 10,
					ResourceGroup: &models.ResourceGroup{
						ID: "rg1",
					},
					Zone: &models.Zone{Name: "test-1"},
				},
			}

			mux, client, teardown := test.SetupServer(t)
			test.SetupMuxResponse(t, mux, "/v1/instances/testinstance/volume_attachments/volumeattachmentid", http.MethodDelete, nil, testcase.status, testcase.content, nil)

			defer teardown()

			logger.Info("Test case being executed", zap.Reflect("testcase", testcase.name))

			volumeAttachService := instances.New(client)

			response, err := volumeAttachService.DetachVolume(template, logger)

			if testcase.expectErr != "" && assert.Error(t, err) {
				assert.Equal(t, testcase.expectErr, err.Error())
				assert.NotNil(t, response)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, response)
			}
			defer response.Body.Close()
		})
	}
}
