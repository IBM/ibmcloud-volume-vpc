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

// AttachVolume attached volume to instances with givne volume attachment details
func (vs *VolumeAttachService) AttachVolume(volumeAttachmentTemplate *models.VolumeAttachment, ctxLogger *zap.Logger) (*models.VolumeAttachment, error) {
	methodName := "VolumeAttachService.AttachVolume"
	defer util.TimeTracker(methodName, time.Now())
	defer metrics.UpdateDurationFromStart(ctxLogger, methodName, time.Now())

	operation := &client.Operation{
		Name:        "AttachVolume",
		Method:      "POST",
		PathPattern: vs.pathPrefix + instanceIDvolumeAttachmentPath,
	}

	var volumeAttachment models.VolumeAttachment
	apiErr := vs.receiverError

	operationRequest := vs.client.NewRequest(operation)

	ctxLogger.Info("Equivalent curl command and payload details", zap.Reflect("URL", operationRequest.URL()), zap.Reflect("Payload", volumeAttachmentTemplate), zap.Reflect("Operation", operation), zap.Reflect("PathParameters", volumeAttachmentTemplate.InstanceID))
	_, err := vs.populatePathPrefixParameters(operationRequest, volumeAttachmentTemplate).JSONBody(volumeAttachmentTemplate).JSONSuccess(&volumeAttachment).JSONError(apiErr).Invoke()
	if err != nil {
		return nil, err
	}

	ctxLogger.Info("Successfuly attached the volume")
	return &volumeAttachment, nil
}
