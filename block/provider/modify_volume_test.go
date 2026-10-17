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
	"errors"
	"testing"

	"github.com/IBM/ibmcloud-volume-interface/lib/provider"
	userError "github.com/IBM/ibmcloud-volume-vpc/common/messages"
	"github.com/IBM/ibmcloud-volume-vpc/common/vpcclient/models"
	volumeServiceFakes "github.com/IBM/ibmcloud-volume-vpc/common/vpcclient/vpcvolume/fakes"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestModifyVolume(t *testing.T) {
	userError.MessagesEn = userError.InitMessages()
	logger, teardown := GetTestLogger(t)
	defer teardown()

	testCases := []struct {
		testCaseName string
		request      provider.ModifyVolumeRequest
		// volume returned by ModifyVolume (client layer)
		modifyVolumeResp *models.Volume
		modifyVolumeErr  error
		// volume returned by GetVolume (for WaitForValidVolumeState)
		getVolumeResp *models.Volume
		getVolumeErr  error
		// expected
		expectErr          bool
		expectedReasonCode string
		expectedIops       int64
		expectedBandwidth  int32
	}{
		{
			testCaseName: "missing VolumeID returns error",
			request:      provider.ModifyVolumeRequest{},
			expectErr:    true,
		},
		{
			testCaseName: "no iops or bandwidth returns empty response",
			request:      provider.ModifyVolumeRequest{VolumeID: "vol-id"},
		},
		{
			testCaseName:     "successful IOPS modify",
			request:          provider.ModifyVolumeRequest{VolumeID: "vol-id", Iops: 5000},
			modifyVolumeResp: &models.Volume{ID: "vol-id", Status: models.StatusType("available"), Iops: 5000},
			getVolumeResp:    &models.Volume{ID: "vol-id", Status: models.StatusType("available"), Iops: 5000},
			expectedIops:     5000,
		},
		{
			testCaseName:      "successful bandwidth modify",
			request:           provider.ModifyVolumeRequest{VolumeID: "vol-id", Bandwidth: 1000},
			modifyVolumeResp:  &models.Volume{ID: "vol-id", Status: models.StatusType("available"), Bandwidth: 1000},
			getVolumeResp:     &models.Volume{ID: "vol-id", Status: models.StatusType("available"), Bandwidth: 1000},
			expectedBandwidth: 1000,
		},
		{
			testCaseName:       "ModifyVolume client error returns FailedToModifyVolume",
			request:            provider.ModifyVolumeRequest{VolumeID: "vol-id", Iops: 5000},
			modifyVolumeErr:    errors.New("backend failure"),
			expectErr:          true,
			expectedReasonCode: "FailedToModifyVolume",
		},
		{
			testCaseName:       "ModifyVolume returns empty volume ID returns FailedToModifyVolume",
			request:            provider.ModifyVolumeRequest{VolumeID: "vol-id", Iops: 5000},
			modifyVolumeResp:   &models.Volume{},
			expectErr:          true,
			expectedReasonCode: "FailedToModifyVolume",
		},
		{
			testCaseName:       "volume stuck in non-available state returns VolumeNotInValidState",
			request:            provider.ModifyVolumeRequest{VolumeID: "vol-id", Iops: 5000},
			modifyVolumeResp:   &models.Volume{ID: "vol-id", Status: models.StatusType("pending"), Iops: 5000},
			getVolumeResp:      &models.Volume{ID: "vol-id", Status: models.StatusType("pending")},
			expectErr:          true,
			expectedReasonCode: "VolumeNotInValidState",
		},
	}

	for _, testcase := range testCases {
		t.Run(testcase.testCaseName, func(t *testing.T) {
			logger.Info("Started", zap.String("testCase", testcase.testCaseName))
			vpcs, uc, sc, err := GetTestOpenSession(t, logger)
			assert.NotNil(t, vpcs)
			assert.NotNil(t, uc)
			assert.NotNil(t, sc)
			assert.Nil(t, err)

			// Limit retries so error-path cases don't sleep between attempts
			origMaxRetryAttempt := maxRetryAttempt
			maxRetryAttempt = 1
			t.Cleanup(func() { maxRetryAttempt = origMaxRetryAttempt })

			volumeService := &volumeServiceFakes.VolumeService{}
			uc.VolumeServiceReturns(volumeService)
			volumeService.ModifyVolumeReturns(testcase.modifyVolumeResp, testcase.modifyVolumeErr)
			volumeService.GetVolumeReturns(testcase.getVolumeResp, testcase.getVolumeErr)

			resp, err := vpcs.ModifyVolume(testcase.request)
			if testcase.expectErr {
				assert.NotNil(t, err)
				assert.Nil(t, resp)
				if testcase.expectedReasonCode != "" && err != nil {
					assert.Contains(t, err.Error(), testcase.expectedReasonCode)
				}
			} else {
				assert.Nil(t, err)
				if assert.NotNil(t, resp) {
					assert.Equal(t, testcase.expectedIops, resp.Iops)
					assert.Equal(t, testcase.expectedBandwidth, resp.Bandwidth)
				}
			}
		})
	}
}
