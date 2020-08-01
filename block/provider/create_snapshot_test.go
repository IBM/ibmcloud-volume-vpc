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
		snapshotService *serviceFakes.SnapshotService
		volumeService   *serviceFakes.VolumeService
	)

	testCases := []struct {
		testCaseName     string
		baseSnapshot     *models.Snapshot
		providerSnapshot *provider.Snapshot
		providerVolume   *provider.Volume
		baseVolume       *models.Volume
		tags             map[string]string
		setup            func()

		skipErrTest        bool
		expectedErr        string
		expectedReasonCode string

		verify func(t *testing.T, snapshotResponse *provider.Snapshot, err error)
	}{
		{
			testCaseName: "Not supported yet",
			providerVolume: &provider.Volume{
				VolumeID: "16f293bf-test-4bff-816f-e199c0c65db5",
			},
			baseVolume: &models.Volume{
				ID:       "16f293bf-test-4bff-816f-e199c0c65db5",
				Name:     "test volume name",
				Status:   models.StatusType("OK"),
				Capacity: int64(10),
				Iops:     int64(1000),
			},
			baseSnapshot: &models.Snapshot{
				ID:     "16f293bf-test-4bff-816f-e199c0c65db5",
				Name:   "test snapshot name",
				Status: models.StatusType("OK"),
			},
			verify: func(t *testing.T, snapshotResponse *provider.Snapshot, err error) {
				assert.Nil(t, snapshotResponse)
				assert.NotNil(t, err)
			},
		}, {
			testCaseName: "Not supported yet",
			providerVolume: &provider.Volume{
				VolumeID: "16f293bf-test-4bff-816f-e199c0c65db5",
			},
			baseVolume: &models.Volume{
				ID:       "16f293bf-test-4bff-816f-e199c0c65db5",
				Name:     "test volume name",
				Status:   models.StatusType("OK"),
				Capacity: int64(10),
				Iops:     int64(1000),
			},
			expectedErr:        "{Code:StorageFindFailedWithSnapshotId, Type:InvalidRequest, Description:'Not a valid snapshot ID",
			expectedReasonCode: "ErrorUnclassified",
			verify: func(t *testing.T, snapshotResponse *provider.Snapshot, err error) {
				assert.Nil(t, snapshotResponse)
				assert.NotNil(t, err)
			},
		}, {
			testCaseName:       "Volume is nil",
			expectedErr:        "{Code:StorageFindFailedWithSnapshotId, Type:InvalidRequest, Description:'Not a valid snapshot ID",
			expectedReasonCode: "ErrorUnclassified",
			verify: func(t *testing.T, snapshotResponse *provider.Snapshot, err error) {
				assert.Nil(t, snapshotResponse)
				assert.NotNil(t, err)
			},
		}, {
			testCaseName: "Volume is not valid",
			providerVolume: &provider.Volume{
				VolumeID: "16f293bf-test-4bff-816f-e199c0c65db5",
			},
			verify: func(t *testing.T, snapshotResponse *provider.Snapshot, err error) {
				assert.Nil(t, snapshotResponse)
				assert.NotNil(t, err)
			},
		}, {
			testCaseName: "Snapshot creation failed",
			skipErrTest:  true,
			providerVolume: &provider.Volume{
				VolumeID: "16f293bf-test-4bff-816f-e199c0c65db5",
			},
			baseVolume: &models.Volume{
				ID:       "16f293bf-test-4bff-816f-e199c0c65db5",
				Name:     "test volume name",
				Status:   models.StatusType("OK"),
				Capacity: int64(10),
				Iops:     int64(1000),
			},
			expectedErr:        "{Code:StorageFindFailedWithSnapshotId, Type:InvalidRequest, Description:'Not a valid snapshot ID",
			expectedReasonCode: "ErrorUnclassified",
			verify: func(t *testing.T, snapshotResponse *provider.Snapshot, err error) {
				assert.Nil(t, snapshotResponse)
				assert.NotNil(t, err)
			},
		}, {
			testCaseName: "Create snapshot",
			providerVolume: &provider.Volume{
				VolumeID: "16f293bf-test-4bff-816f-e199c0c65db5",
			},
			baseVolume: &models.Volume{
				ID:       "16f293bf-test-4bff-816f-e199c0c65db5",
				Name:     "test volume name",
				Status:   models.StatusType("OK"),
				Capacity: int64(10),
				Iops:     int64(1000),
			},
			providerSnapshot: &provider.Snapshot{
				Volume: provider.Volume{
					VolumeID: "16f293bf-test-4bff-816f-e199c0c65db5",
				},
				SnapshotID: "s6f293bf-test-4bff-816f-e199c0c65db5",
			},
			baseSnapshot: &models.Snapshot{
				ID:     "16f293bf-test-4bff-816f-e199c0c65db5",
				Name:   "test-snapshot-name",
				Status: models.StatusType("OK"),
			},
			verify: func(t *testing.T, snapshotResponse *provider.Snapshot, err error) {
				assert.Nil(t, snapshotResponse)
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

			if testcase.skipErrTest == true {
				snapshotService.CreateSnapshotReturns(testcase.baseSnapshot, nil)
				volumeService.GetVolumeReturns(testcase.baseVolume, errors.New("errorUnclassified"))
			} else {
				if testcase.expectedErr != "" {
					snapshotService.CreateSnapshotReturns(testcase.baseSnapshot, errors.New(testcase.expectedReasonCode))
					volumeService.GetVolumeReturns(testcase.baseVolume, errors.New(testcase.expectedReasonCode))
				} else {
					snapshotService.CreateSnapshotReturns(testcase.baseSnapshot, nil)
					volumeService.GetVolumeReturns(testcase.baseVolume, nil)
				}
			}

			snapshot, err := vpcs.CreateSnapshot(testcase.providerVolume, testcase.tags)
			logger.Info("Snapshot details", zap.Reflect("snapshot", snapshot))

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

func TestCreateSnapshotTwo(t *testing.T) {
	//var err error
	logger, teardown := GetTestLogger(t)
	defer teardown()

	timeNow := time.Now()

	var (
		snapshotService *serviceFakes.SnapshotService
		volumeService   *serviceFakes.VolumeService
		baseSnapshot    *models.Snapshot
		providerVolume  *provider.Volume
		baseVolume      *models.Volume
		tags            map[string]string
	)

	providerVolume = &provider.Volume{
		VolumeID: "16f293bf-test-4bff-816f-e199c0c65db5",
	}
	baseVolume = &models.Volume{
		ID:       "16f293bf-test-4bff-816f-e199c0c65db5",
		Name:     "test volume name",
		Status:   models.StatusType("OK"),
		Capacity: int64(10),
		Iops:     int64(1000),
		Zone:     &models.Zone{Name: "test-zone"},
	}
	baseSnapshot = &models.Snapshot{
		ID:        "16f293bf-test-4bff-816f-e199c0c65db5",
		Name:      "test-snapshot-name",
		Status:    models.StatusType("OK"),
		CreatedAt: &timeNow,
	}
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

	snapshotService.CreateSnapshotReturns(baseSnapshot, errors.New("errorUnclassified"))
	volumeService.GetVolumeReturns(baseVolume, nil)

	snapshot, err := vpcs.CreateSnapshot(providerVolume, tags)
	logger.Info("Snapshot details", zap.Reflect("snapshot", snapshot))

	assert.Nil(t, snapshot)
	assert.NotNil(t, err)

	snapshotService.CreateSnapshotReturns(baseSnapshot, nil)
	volumeService.GetVolumeReturns(baseVolume, nil)

	snapshot, err = vpcs.CreateSnapshot(providerVolume, tags)
	logger.Info("Snapshot details", zap.Reflect("snapshot", snapshot))

	assert.NotNil(t, snapshot)
	assert.Nil(t, err)
}
