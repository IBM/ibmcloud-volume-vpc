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

package instances_test

import (
	"github.com/IBM/ibmcloud-storage-volume-lib/volume-providers/vpc/vpcclient/instances"
	"github.com/IBM/ibmcloud-storage-volume-lib/volume-providers/vpc/vpcclient/models"
	"github.com/IBM/ibmcloud-storage-volume-lib/volume-providers/vpc/vpcclient/riaas/test"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"net/http"
	"testing"
)

func TestIKSAttachVolume(t *testing.T) {
	// Setup new style zap logger
	logger, _ := GetTestContextLogger()
	defer logger.Sync()

	instanceID := "testinstance"
	clusterID := "testcluster"
	// IKS tests
	mux, client, teardown := test.SetupServer(t)

	content := "{\"id\":\"volume attachment id\", \"name\":\"volume attachment\", \"device\": {\"id\":\"xvdc\"}, \"volume\": {\"id\":\"volume-id\",\"name\":\"volume-name\",\"capacity\":10,\"iops\":3000,\"status\":\"pending\"}}"

	test.SetupMuxResponse(t, mux, "/v2/storage/vpc/createAttachment", http.MethodPost, nil, http.StatusOK, content, nil)
	volumeAttachService := instances.NewIKSVolumeAttachmentManager(client)

	template := &models.VolumeAttachment{
		Name:       "volume attachment",
		ClusterID:  &clusterID,
		InstanceID: &instanceID,
		Volume: &models.Volume{
			ID:       "volume-id",
			Name:     "volume-name",
			Capacity: 10,
			ResourceGroup: &models.ResourceGroup{
				ID: "rg1",
			},
			Zone: &models.Zone{Name: "test-1"},
		},
	}
	defer teardown()

	volumeAttachment, err := volumeAttachService.AttachVolume(template, logger)
	logger.Info("Volume attachment details", zap.Reflect("volumeAttachment", volumeAttachment))

	assert.NoError(t, err)
	assert.NotNil(t, volumeAttachment)
}
