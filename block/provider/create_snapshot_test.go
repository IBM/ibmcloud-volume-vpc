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

// Package provider ...
package provider

import (
	"errors"
	"testing"
	"time"

	"github.com/IBM/ibmcloud-volume-interface/lib/provider"
	util "github.com/IBM/ibmcloud-volume-interface/lib/utils"
	"github.com/IBM/ibmcloud-volume-interface/lib/utils/reasoncode"
	"github.com/IBM/ibmcloud-volume-vpc/common/vpcclient/models"
	serviceFakes "github.com/IBM/ibmcloud-volume-vpc/common/vpcclient/vpcvolume/fakes"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestCreateSnapshot(t *testing.T) {
	//var err error
	logger, teardown := GetTestLogger(t)
	defer teardown()

	var (
		snapshotService *serviceFakes.SnapshotManager
		volumeService   *serviceFakes.VolumeService
	)
	timeNow := time.Now()

	testCases := []struct {
		baseVolume                 *models.Volume
		testCaseName               string
		sourceVolumeID             string
		baseSnapshot               *models.Snapshot
		providerSnapshotParameters *provider.SnapshotParameters
		providerSnapshot           *provider.Snapshot
		setup                      func()

		skipErrTest        bool
		expectedErr        string
		expectedReasonCode string

		verify func(t *testing.T, snapshotResponse *provider.Snapshot, err error)
	}{
		{
			testCaseName:   "Snapshot name not provided",
			sourceVolumeID: "16f293bf-test-4bff-816f-e199c0c65db5",
			providerSnapshotParameters: &provider.SnapshotParameters{
				Name: nil,
			},
			providerSnapshot: nil,
			baseSnapshot: &models.Snapshot{
				ID:             "16f293bf-test-4bff-816f-e199c0c65db5",
				Name:           "test snapshot name",
				LifecycleState: "stable",
			},
			expectedErr:        "{Code:ErrorUnclassified, Type:InvalidRequest, Description:'Name is required to complete the operation.",
			expectedReasonCode: "ErrorUnclassified",
			verify: func(t *testing.T, snapshotResponse *provider.Snapshot, err error) {
				assert.Nil(t, snapshotResponse)
				assert.NotNil(t, err)
			},
		}, {
			testCaseName:   "Source VolumeID for snapshot not provided",
			sourceVolumeID: "",
			providerSnapshotParameters: &provider.SnapshotParameters{
				Name: String("test snapshot"),
			},
			providerSnapshot: nil,
			baseSnapshot: &models.Snapshot{
				ID:             "16f293bf-test-4bff-816f-e199c0c65db5",
				Name:           "test snapshot name",
				LifecycleState: "stable",
			},
			expectedErr:        "{Code:ErrorUnclassified, Type:InvalidRequest, Description:'SourceVolumeID is required to complete the operation.",
			expectedReasonCode: "ErrorUnclassified",
			verify: func(t *testing.T, snapshotResponse *provider.Snapshot, err error) {
				assert.Nil(t, snapshotResponse)
				assert.NotNil(t, err)
			},
		}, {
			testCaseName:   "Source Volume provided for snapshot not present",
			sourceVolumeID: "16f293bf-test-4bff-816f-e199c0c65db5",
			providerSnapshotParameters: &provider.SnapshotParameters{
				Name: String("test snapshot"),
			},
			providerSnapshot: nil,
			baseSnapshot: &models.Snapshot{
				ID:             "16f293bf-test-4bff-816f-e199c0c65db5",
				Name:           "test snapshot name",
				LifecycleState: "stable",
			},
			expectedErr:        "{Code:ErrorUnclassified, Type:RetrivalFailed, Description:'A volume with the specified volume ID '16f293bf-test-4bff-816f-e199c0c65db5' could not be found.",
			expectedReasonCode: "ErrorUnclassified",
			verify: func(t *testing.T, snapshotResponse *provider.Snapshot, err error) {
				assert.Nil(t, snapshotResponse)
				assert.NotNil(t, err)
			},
		}, {
			testCaseName:   "Snapshot creation failed",
			sourceVolumeID: "16f293bf-test-4bff-816f-e199c0c65db5",
			providerSnapshotParameters: &provider.SnapshotParameters{
				Name: String("test snapshot"),
			},
			providerSnapshot: nil,
			baseSnapshot: &models.Snapshot{
				ID:             "16f293bf-test-4bff-816f-e199c0c65db5",
				Name:           "test snapshot name",
				LifecycleState: "stable",
			},
			expectedReasonCode: "SnapshotSpaceOrderFailed",
			verify: func(t *testing.T, snapshotResponse *provider.Snapshot, err error) {
				assert.Nil(t, snapshotResponse)
				assert.NotNil(t, err)
			},
		}, {
			testCaseName:               "Snapshot name is nil",
			sourceVolumeID:             "16f293bf-test-4bff-816f-e199c0c65db5",
			providerSnapshotParameters: &provider.SnapshotParameters{},
			verify: func(t *testing.T, snapshotResponse *provider.Snapshot, err error) {
				assert.Nil(t, snapshotResponse)
				assert.NotNil(t, err)
			},
		}, {
			testCaseName:   "Snapshot name is empty",
			sourceVolumeID: "16f293bf-test-4bff-816f-e199c0c65db5",
			baseSnapshot: &models.Snapshot{
				ID:             "16f293bf-test-4bff-816f-e199c0c65db5",
				Name:           "test snapshot name",
				LifecycleState: "stable",
			},
			providerSnapshotParameters: &provider.SnapshotParameters{
				Name: String(""),
			},
			verify: func(t *testing.T, snapshotResponse *provider.Snapshot, err error) {
				assert.Nil(t, snapshotResponse)
				assert.NotNil(t, err)
			},
		}, {
			testCaseName: "Snapshot in pending state",
			baseSnapshot: &models.Snapshot{
				ID:             "16f293bf-test-4bff-816f-e199c0c65db5",
				Name:           "test-snapshot-name",
				LifecycleState: "pending",
				SourceVolume:   &models.SourceVolume{ID: "16f293bf-test-4bff-816f-e199c0c65db6"},
				CreatedAt:      &timeNow,
				Size:           100,
			},
			sourceVolumeID: "16f293bf-test-4bff-816f-e199c0c65db5",
			providerSnapshotParameters: &provider.SnapshotParameters{
				Name: String("test snapshot name"),
			},
			providerSnapshot: &provider.Snapshot{
				VolumeID:             "16f293bf-test-4bff-816f-e199c0c65db6",
				SnapshotID:           "16f293bf-test-4bff-816f-e199c0c65db5",
				SnapshotSize:         100,
				SnapshotCreationTime: timeNow,
				ReadyToUse:           false,
				VPC:                  &provider.VPC{Href: "href"},
			},
			baseVolume: &models.Volume{
				ID: "16f293bf-test-4bff-816f-e199c0c65db6",
			},
			verify: func(t *testing.T, snapshotResponse *provider.Snapshot, err error) {
				assert.NotNil(t, snapshotResponse)
				assert.Nil(t, err)
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

			snapshotService = &serviceFakes.SnapshotManager{}
			assert.NotNil(t, snapshotService)
			uc.SnapshotServiceReturns(snapshotService)

			volumeService = &serviceFakes.VolumeService{}
			assert.NotNil(t, volumeService)
			uc.VolumeServiceReturns(volumeService)

			if testcase.expectedErr != "" {
				snapshotService.CreateSnapshotReturns(testcase.baseSnapshot, errors.New(testcase.expectedReasonCode))
				volumeService.GetVolumeReturns(testcase.baseVolume, errors.New(testcase.expectedReasonCode))
			} else {
				snapshotService.CreateSnapshotReturns(testcase.baseSnapshot, nil)
				volumeService.GetVolumeReturns(testcase.baseVolume, nil)
			}
			snapshot, err := vpcs.CreateSnapshot(testcase.sourceVolumeID, *testcase.providerSnapshotParameters)
			logger.Info("snapshot details", zap.Reflect("snapshot", snapshot))

			if testcase.expectedErr != "" {
				assert.NotNil(t, err)
				logger.Info("Error details", zap.Reflect("Error details", err.Error()))
				assert.Equal(t, reasoncode.ReasonCode(testcase.expectedReasonCode), util.ErrorReasonCode(err))
			}

			if testcase.verify != nil {
				testcase.verify(t, snapshot, err)
			}
		})
	}
}
