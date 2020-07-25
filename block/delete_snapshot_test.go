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
	"errors"
	"github.com/IBM/ibmcloud-storage-volume-lib/lib/provider"
	"github.com/IBM/ibmcloud-storage-volume-lib/lib/utils"
	"github.com/IBM/ibmcloud-storage-volume-lib/lib/utils/reasoncode"
	"github.com/IBM/ibmcloud-storage-volume-lib/volume-providers/vpc/vpcclient/models"
	serviceFakes "github.com/IBM/ibmcloud-storage-volume-lib/volume-providers/vpc/vpcclient/vpcvolume/fakes"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"testing"
)

func TestDeleteSnapshot(t *testing.T) {
	//var err error
	logger, teardown := GetTestLogger(t)
	defer teardown()

	var (
		snapshotService *serviceFakes.SnapshotService
		volumeService   *serviceFakes.VolumeService
	)

	testCases := []struct {
		testCaseName     string
		baseVolume       *models.Volume
		baseSnapshot     *models.Snapshot
		providerVolume   *provider.Volume
		providerSnapshot *provider.Snapshot

		tags  map[string]string
		setup func()

		skipErrTest        bool
		expectedErr        string
		expectedReasonCode string

		verify func(t *testing.T, err error)
	}{
		{
			testCaseName: "Not supported yet",
			providerSnapshot: &provider.Snapshot{
				SnapshotID: "s6f293bf-test-4bff-816f-e199c0c65db5",
				Volume: provider.Volume{
					VolumeID: "16f293bf-test-4bff-816f-e199c0c65db5",
					Name:     String("Test volume"),
					Capacity: Int(10),
					Iops:     String("1000"),
					VPCVolume: provider.VPCVolume{
						Profile: &provider.Profile{Name: "general-purpose"},
					},
				},
			},
			verify: func(t *testing.T, err error) {
				assert.Nil(t, err)
			},
		}, {
			testCaseName: "Not a valid snapshot",
			providerSnapshot: &provider.Snapshot{
				SnapshotID: "s6f293bf-test-4bff-816f-e199c0c65db5",
				Volume: provider.Volume{
					VolumeID: "16f293bf-test-4bff-816f-e199c0c65db5",
					Name:     String("Test volume"),
					Capacity: Int(10),
					Iops:     String("1000"),
					VPCVolume: provider.VPCVolume{
						Profile: &provider.Profile{Name: "general-purpose"},
					},
				},
			},
			expectedErr:        "{Code:ErrorUnclassified, Type:InvalidRequest, Description:'Not a valid snapshot ID",
			expectedReasonCode: "ErrorUnclassified",
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

			snapshotService = &serviceFakes.SnapshotService{}
			assert.NotNil(t, snapshotService)
			uc.SnapshotServiceReturns(snapshotService)

			volumeService = &serviceFakes.VolumeService{}
			assert.NotNil(t, volumeService)
			uc.VolumeServiceReturns(volumeService)

			if testcase.expectedErr != "" {
				snapshotService.DeleteSnapshotReturns(errors.New(testcase.expectedReasonCode))
				snapshotService.GetSnapshotReturns(testcase.baseSnapshot, errors.New(testcase.expectedReasonCode))
			} else {
				snapshotService.DeleteSnapshotReturns(nil)
				snapshotService.GetSnapshotReturns(testcase.baseSnapshot, nil)
			}
			err = vpcs.DeleteSnapshot(testcase.providerSnapshot)

			if testcase.expectedErr != "" {
				assert.NotNil(t, err)
				logger.Info("Error details", zap.Reflect("Error details", err.Error()))
				assert.Equal(t, reasoncode.ReasonCode(testcase.expectedReasonCode), util.ErrorReasonCode(err))
			}

			if testcase.verify != nil {
				testcase.verify(t, err)
			}

		})
	}
}
