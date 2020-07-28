/*******************************************************************************
 * IBM Confidential
 * OCO Source Materials
 * IBM Cloud Container Service, 5737-D43
 * (C) Copyright IBM Corp. 2018 All Rights Reserved.
 * The source code for this program is not  published or otherwise divested of
 * its trade secrets, irrespective of what has been deposited with
 * the U.S. Copyright Office.
 ******************************************************************************/

package provider

import (
	"errors"
	"github.com/IBM/ibmcloud-volume-interface/lib/provider"
	"github.com/IBM/ibmcloud-volume-interface/lib/utils"
	"github.com/IBM/ibmcloud-volume-interface/lib/utils/reasoncode"
	volumeAttachServiceFakes "github.com/IBM/ibmcloud-volume-vpc/common/vpcclient/instances/fakes"
	"github.com/IBM/ibmcloud-volume-vpc/common/vpcclient/models"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"testing"
)

func TestGetVolumeAttachment(t *testing.T) {
	//var err error
	logger, teardown := GetTestLogger(t)
	defer teardown()

	var (
		volumeAttachService *volumeAttachServiceFakes.VolumeAttachService
	)

	testCases := []struct {
		testCaseName                      string
		providerVolumeAttachmentRequest   provider.VolumeAttachmentRequest
		baseVolumeAttachmentRequest       *models.VolumeAttachment
		providerVolumeAttachmentResponse  provider.VolumeAttachmentResponse
		baseVolumeAttachmentsListResponse *models.VolumeAttachmentList

		setup func(providerVolume *provider.Volume)

		skipErrTest        bool
		expectedErr        string
		expectedReasonCode string

		verify func(t *testing.T, volumeAttachmentResponse *provider.VolumeAttachmentResponse, err error)
	}{
		{
			testCaseName: "Instance ID is nil",
			providerVolumeAttachmentRequest: provider.VolumeAttachmentRequest{
				VolumeID: "volume-id1",
			},
		}, {
			testCaseName: "Volume ID is nil",
			providerVolumeAttachmentRequest: provider.VolumeAttachmentRequest{
				InstanceID: "instance-id1",
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

			volumeAttachService = &volumeAttachServiceFakes.VolumeAttachService{}
			assert.NotNil(t, volumeAttachService)
			uc.VolumeAttachServiceReturns(volumeAttachService)

			if testcase.expectedErr != "" {
				volumeAttachService.ListVolumeAttachmentsReturns(testcase.baseVolumeAttachmentsListResponse, errors.New(testcase.expectedReasonCode))
			} else {
				volumeAttachService.ListVolumeAttachmentsReturns(testcase.baseVolumeAttachmentsListResponse, nil)
			}
			volumeAttachment, err := vpcs.GetVolumeAttachment(testcase.providerVolumeAttachmentRequest)
			logger.Info("Volume attachment details", zap.Reflect("VolumeAttachmentResponse", volumeAttachment))

			if testcase.expectedErr != "" {
				assert.NotNil(t, err)
				logger.Info("Error details", zap.Reflect("Error details", err.Error()))
				assert.Equal(t, reasoncode.ReasonCode(testcase.expectedReasonCode), util.ErrorReasonCode(err))
			}

			if testcase.verify != nil {
				testcase.verify(t, volumeAttachment, err)
			}

		})
	}
}
