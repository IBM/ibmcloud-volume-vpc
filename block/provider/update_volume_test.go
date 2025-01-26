/**
 * Copyright 2025 IBM Corp.
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
	util "github.com/IBM/ibmcloud-volume-interface/lib/utils"
	"github.com/IBM/ibmcloud-volume-interface/lib/utils/reasoncode"
	"github.com/IBM/ibmcloud-volume-vpc/common/vpcclient/models"
	volumeServiceFakes "github.com/IBM/ibmcloud-volume-vpc/common/vpcclient/vpcvolume/fakes"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestUpdateVolume(t *testing.T) {
	logger, teardown := GetTestLogger(t)
	defer teardown()

	var (
		volumeService *volumeServiceFakes.VolumeService
	)

	testCases := []struct {
		testCaseName       string
		volumeID           string
		baseVolume         *models.Volume
		newSize            int64
		tags               []string
		expectedErr        string
		expectedSize       int64
		expectedReasonCode string
	}{
		{
			testCaseName: "OK",
			volumeID:     "16f293bf-test-4bff-816f-e199c0c65db5",
			baseVolume: &models.Volume{
				ID:       "16f293bf-test-4bff-816f-e199c0c65db5",
				Status:   models.StatusType("available"),
				Capacity: int64(10),
				Iops:     int64(1000),
				UserTags: []string{"tag3:val3"},
				Zone:     &models.Zone{Name: "test-zone"},
			},
			tags: []string{"tag1:val1", "tag2:val2"},
		},
		{
			testCaseName: "Tags are equal",
			volumeID:     "16f293bf-test-4bff-816f-e199c0c65db5",
			baseVolume: &models.Volume{
				ID:       "16f293bf-test-4bff-816f-e199c0c65db5",
				Status:   models.StatusType("available"),
				Capacity: int64(10),
				Iops:     int64(1000),
				UserTags: []string{"tag1:val1", "tag2:val2"},
				Zone:     &models.Zone{Name: "test-zone"},
			},
			tags: []string{"tag1:val1", "tag2:val2"},
		},
		{
			testCaseName: "New tags added",
			volumeID:     "16f293bf-test-4bff-816f-e199c0c65db5",
			baseVolume: &models.Volume{
				ID:       "16f293bf-test-4bff-816f-e199c0c65db5",
				Status:   models.StatusType("available"),
				Capacity: int64(10),
				Iops:     int64(1000),
				UserTags: []string{"tag3:val3"},
				Zone:     &models.Zone{Name: "test-zone"},
			},
			tags: []string{"tag1:val1", "tag2:val2"},
		},
		{
			testCaseName: "Volume is not available for update",
			volumeID:     "16f293bf-test-4bff-816f-e199c0c65db5",
			baseVolume: &models.Volume{
				ID:       "16f293bf-test-4bff-816f-e199c0c65db5",
				Status:   models.StatusType("updating"),
				Capacity: int64(10),
				Iops:     int64(1000),
				Zone:     &models.Zone{Name: "test-zone"},
			},
			tags:               []string{"tag1:val1", "tag2:val2"},
			expectedErr:        "{Code:ErrorUnclassified, Type:VolumeNotInValidState, Description:Volume did not get valid (available) status within timeout period., BackendError:, RC:500}",
			expectedReasonCode: "ErrorUnclassified",
		},

		{
			testCaseName:       "volume not found",
			volumeID:           "16f293bf-test-4bff-816f-e199c0c65db5",
			baseVolume:         nil,
			expectedErr:        "{Code:ErrorUnclassified, Type:InvalidRequest, Description:'Wrong volume ID' volume ID is not valid. Please check https://cloud.ibm.com/docs/infrastructure/vpc?topic=vpc-rias-error-messages#volume_id_invalid, BackendError:, RC:400}",
			expectedReasonCode: "ErrorUnclassified",
		},
	}

	for _, testcase := range testCases {
		t.Run(testcase.testCaseName, func(t *testing.T) {
			vpcs, uc, sc, err := GetTestOpenSession(t, logger)
			assert.NotNil(t, vpcs)
			assert.NotNil(t, uc)
			assert.NotNil(t, sc)
			assert.Nil(t, err)

			volumeService = &volumeServiceFakes.VolumeService{}
			assert.NotNil(t, volumeService)
			uc.VolumeServiceReturns(volumeService)

			if testcase.expectedErr != "" {
				volumeService.GetVolumeReturns(testcase.baseVolume, errors.New(testcase.expectedReasonCode))
				volumeService.UpdateVolumeReturns(errors.New(testcase.expectedReasonCode))
			} else {
				volumeService.GetVolumeReturns(testcase.baseVolume, nil)
				volumeService.UpdateVolumeReturns(nil)
			}
			requestExp := provider.Volume{VolumeID: testcase.volumeID,
				VPCVolume: provider.VPCVolume{Tags: testcase.tags}}

			err = vpcs.UpdateVolume(requestExp)

			if testcase.expectedErr != "" {
				assert.NotNil(t, err)
				logger.Info("Error details", zap.Reflect("Error details", err.Error()))
				assert.Equal(t, reasoncode.ReasonCode(testcase.expectedReasonCode), util.ErrorReasonCode(err))
			} else {
				assert.Nil(t, err)
			}
		})
	}
}
