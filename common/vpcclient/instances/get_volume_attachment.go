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

package instances

import (
	"github.com/IBM/ibmcloud-storage-volume-lib/lib/metrics"
	"github.com/IBM/ibmcloud-storage-volume-lib/lib/utils"
	"github.com/IBM/ibmcloud-storage-volume-lib/volume-providers/vpc/vpcclient/client"
	"github.com/IBM/ibmcloud-storage-volume-lib/volume-providers/vpc/vpcclient/models"
	"go.uber.org/zap"
	"time"
)

// GetVolumeAttachment retrives the volume attach status with given volume attachment details
func (vs *VolumeAttachService) GetVolumeAttachment(volumeAttachmentTemplate *models.VolumeAttachment, ctxLogger *zap.Logger) (*models.VolumeAttachment, error) {

	methodName := "VolumeAttachService.GetVolumeAttachment"
	defer util.TimeTracker(methodName, time.Now())
	defer metrics.UpdateDurationFromStart(ctxLogger, methodName, time.Now())

	operation := &client.Operation{
		Name:        "GetVolumeAttachment",
		Method:      "GET",
		PathPattern: vs.pathPrefix + instanceIDattachmentIDPath,
	}

	apiErr := vs.receiverError
	var volumeAttachment models.VolumeAttachment
	operationRequest := vs.client.NewRequest(operation)

	ctxLogger.Info("Equivalent curl command details", zap.Reflect("URL", operationRequest.URL()), zap.Reflect("volumeAttachmentTemplate", volumeAttachmentTemplate), zap.Reflect("Operation", operation))
	ctxLogger.Info("Pathparameters", zap.Reflect(instanceIDParam, volumeAttachmentTemplate.InstanceID), zap.Reflect(attachmentIDParam, volumeAttachmentTemplate.ID))
	operationRequest = vs.populatePathPrefixParameters(operationRequest, volumeAttachmentTemplate)
	operationRequest = operationRequest.PathParameter(attachmentIDParam, volumeAttachmentTemplate.ID)

	_, err := operationRequest.JSONSuccess(&volumeAttachment).JSONError(apiErr).Invoke()
	if err != nil {
		ctxLogger.Error("Error occured while getting volume attachment", zap.Error(err))
		return nil, err
	}
	ctxLogger.Info("Successfuly retrieved the volume attachment", zap.Reflect("volumeAttachment", volumeAttachment))
	return &volumeAttachment, err
}
