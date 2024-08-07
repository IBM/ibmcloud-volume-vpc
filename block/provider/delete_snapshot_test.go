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

	"github.com/IBM/ibmcloud-volume-interface/lib/provider"
	util "github.com/IBM/ibmcloud-volume-interface/lib/utils"
	"github.com/IBM/ibmcloud-volume-interface/lib/utils/reasoncode"
	userError "github.com/IBM/ibmcloud-volume-vpc/common/messages"
	"github.com/IBM/ibmcloud-volume-vpc/common/vpcclient/models"
	serviceFakes "github.com/IBM/ibmcloud-volume-vpc/common/vpcclient/vpcvolume/fakes"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestDeleteSnapshot(t *testing.T) {
	//var err error
	logger, teardown := GetTestLogger(t)
	defer teardown()

	var (
		snapshotService *serviceFakes.SnapshotManager
	)

	testCases := []struct {
		testCaseName     string
		providerSnapshot *provider.Snapshot
		setup            func()

		skipErrTest        bool
		expectedErr        string
		expectedReasonCode string
		errorCode          *models.Error

		verify func(t *testing.T, err error)
	}{
		{
			testCaseName: "Not supported yet",
			providerSnapshot: &provider.Snapshot{
				VolumeID:   "16f293bf-test-4bff-816f-e199c0c65db5",
				SnapshotID: "16f293bf-test-4bff-816f-e199c0c65db6",
			},
			verify: func(t *testing.T, err error) {
				assert.Nil(t, err)
			},
		}, {
			testCaseName: "False positive: No Snapshot being sent",
			providerSnapshot: &provider.Snapshot{
				SnapshotID: "",
			},
			expectedErr:        "{Code:ErrorUnclassified, Type:InvalidRequest, Description:'Not a valid snapshot ID",
			expectedReasonCode: "ErrorUnclassified",
			verify: func(t *testing.T, err error) {
				assert.NotNil(t, err)
			},
		}, {
			testCaseName: "Incorrect snapshot ID",
			providerSnapshot: &provider.Snapshot{
				VolumeID:   "16f293bf-test-4bff-816f-e199c0c65db5",
				SnapshotID: "16f293bf-test-4bff-816f-e199c0c65db6",
			},
			expectedErr:        "{Code:SnapshotIDNotFound, Type:RetrivalFailed, Description:'Not a valid snapshot ID",
			expectedReasonCode: "SnapshotIDNotFound",
			errorCode: &models.Error{
				Errors: []models.ErrorItem{{
					Code: models.ErrorCode(SnapshotNotFound),
				}},
			},
			verify: func(t *testing.T, err error) {
				assert.NotNil(t, err)
			},
		},
	}

	for _, testcase := range testCases {
		userError.MessagesEn = userError.InitMessages()
		t.Run(testcase.testCaseName, func(t *testing.T) {
			vpcs, uc, sc, err := GetTestOpenSession(t, logger)
			assert.NotNil(t, vpcs)
			assert.NotNil(t, uc)
			assert.NotNil(t, sc)
			assert.Nil(t, err)

			snapshotService = &serviceFakes.SnapshotManager{}
			assert.NotNil(t, snapshotService)
			uc.SnapshotServiceReturns(snapshotService)

			if testcase.errorCode != nil {
				snapshotService.DeleteSnapshotReturns(testcase.errorCode)
			} else if testcase.expectedErr != "" {
				snapshotService.DeleteSnapshotReturns(errors.New(testcase.expectedReasonCode))
			} else {
				snapshotService.DeleteSnapshotReturns(nil)
			}
			err = vpcs.DeleteSnapshot(testcase.providerSnapshot)

			if testcase.expectedErr != "" && testcase.errorCode == nil {
				assert.NotNil(t, err)
				logger.Info("Error details", zap.Reflect("Error details", err.Error()))
				assert.Equal(t, reasoncode.ReasonCode(testcase.expectedReasonCode), util.ErrorReasonCode(err))
			} else if testcase.errorCode != nil {
				assert.NotNil(t, err)
				logger.Info("Error details", zap.Reflect("Error details", err.Error()))
				assert.Equal(t, err.(util.Message).Code, testcase.expectedReasonCode)
			}

			if testcase.verify != nil {
				testcase.verify(t, err)
			}
		})
	}
}

func TestDeleteSnapshotTwo(t *testing.T) {
	//var err error
	logger, teardown := GetTestLogger(t)
	defer teardown()

	var (
		snapshotService *serviceFakes.SnapshotManager
	)
	providerSnapshot := &provider.Snapshot{
		VolumeID:   "16f293bf-test-4bff-816f-e199c0c65db5",
		SnapshotID: "16f293bf-test-4bff-816f-e199c0c65db6",
	}

	vpcs, uc, sc, err := GetTestOpenSession(t, logger)
	assert.NotNil(t, vpcs)
	assert.NotNil(t, uc)
	assert.NotNil(t, sc)
	assert.Nil(t, err)
	snapshotService = &serviceFakes.SnapshotManager{}
	assert.NotNil(t, snapshotService)
	uc.SnapshotServiceReturns(snapshotService)

	snapshotService.DeleteSnapshotReturns(errors.New("not_found"))

	err = vpcs.DeleteSnapshot(providerSnapshot)
	assert.NotNil(t, err)

	snapshotService.DeleteSnapshotReturns(errors.New("failedToDeleteSnapshot"))

	err = vpcs.DeleteSnapshot(providerSnapshot)
	assert.NotNil(t, err)

	snapshotService.DeleteSnapshotReturns(errors.New("failedToDeleteSnapshot"))

	err = vpcs.DeleteSnapshot(providerSnapshot)
	assert.NotNil(t, err)
}
