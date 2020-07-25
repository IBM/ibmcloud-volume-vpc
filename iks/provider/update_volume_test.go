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

package provider

import (
	//"errors"
	"github.com/IBM/ibmcloud-storage-volume-lib/lib/provider"
	volumeServiceFakes "github.com/IBM/ibmcloud-storage-volume-lib/volume-providers/vpc/vpcclient/vpcvolume/fakes"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"testing"
)

func TestUpdateVolume(t *testing.T) {
	//var err error
	logger, teardown := GetTestLogger(t)
	defer teardown()

	var (
		volumeService *volumeServiceFakes.VolumeService
	)

	testCases := []struct {
		testCaseName   string
		providerVolume provider.Volume
		profileName    string

		setup func(providerVolume *provider.Volume)

		skipErrTest        bool
		expectedErr        string
		expectedReasonCode string

		verify func(t *testing.T, err error)
	}{
		{
			testCaseName: "Volume Update Success",
			providerVolume: provider.Volume{
				VolumeID:   "16f293bf-test-4bff-816f-e199c0c65db5",
				Name:       String("test volume name"),
				Capacity:   nil,
				Provider:   provider.VolumeProvider("vpc-classic"),
				VolumeType: provider.VolumeType("block"),
			},
			verify: func(t *testing.T, err error) {
				assert.Nil(t, err)
			},
		}, {
			testCaseName: "VolumeID Empty",
			providerVolume: provider.Volume{
				Name:     String("test volume name"),
				Capacity: Int(0),
			},
			expectedErr:        "{Code:ErrorRequiredFieldMissing, Type:InvalidRequest, Description:[VolumeID] is required to complete the operation., BackendError:, RC:400}",
			expectedReasonCode: "ErrorRequiredFieldMissing",
			verify: func(t *testing.T, err error) {
				assert.NotNil(t, err)
			},
		}, {
			testCaseName: "Volume Provider Empty",
			providerVolume: provider.Volume{
				VolumeID: "16f293bf-test-4bff-816f-e199c0c65db5",
				Name:     String("test volume name"),
				Capacity: Int(0),
			},
			expectedErr:        "{Code:ErrorRequiredFieldMissing, Type:InvalidRequest, Description:[Provider] is required to complete the operation., BackendError:, RC:400}",
			expectedReasonCode: "ErrorRequiredFieldMissing",
			verify: func(t *testing.T, err error) {
				assert.NotNil(t, err)
			},
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
			/*
				if testcase.expectedErr != "" {
					volumeService.UpdateVolumeReturns(errors.New(testcase.expectedReasonCode))
				} else {
					volumeService.UpdateVolumeReturns(nil)
				}*/
			err = vpcs.UpdateVolume(testcase.providerVolume)

			if testcase.expectedErr != "" {
				assert.NotNil(t, err)
				logger.Info("Error details", zap.Reflect("Error details", err.Error()))
				assert.Equal(t, testcase.expectedErr, err.Error())
			}

			if testcase.verify != nil {
				testcase.verify(t, err)
			}

		})
	}
}

// String returns a pointer to the string value provided
func String(v string) *string {
	return &v
}

// Int returns a pointer to the int value provided
func Int(v int) *int {
	return &v
}
