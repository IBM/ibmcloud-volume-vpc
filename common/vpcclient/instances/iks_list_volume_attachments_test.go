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
	"net/http"
	"testing"

	"github.com/IBM/ibmcloud-volume-vpc/common/vpcclient/instances"
	"github.com/IBM/ibmcloud-volume-vpc/common/vpcclient/models"
	"github.com/IBM/ibmcloud-volume-vpc/common/vpcclient/riaas/test"
	"github.com/stretchr/testify/assert"
)

func TestIKSListVolumeAttachment(t *testing.T) {
	// Setup new style zap logger
	logger, _ := GetTestContextLogger()
	defer logger.Sync()

	instanceID := "testinstance"
	clusterID := "testcluster"
	// IKS tests
	mux, client, teardown := test.SetupServer(t)

	content := "{\"volume_attachments\":[{\"id\":\"volumeattachmentid\",\"volume\":{\"name\":\"volume-name\",\"id\":\"volume-id\"},\"device\":{\"id\":\"xvda\"},\"name\":\"attachment-volume-id\",\"status\":\"attached\",\"type\":\"boot\"}]}"

	test.SetupMuxResponse(t, mux, "/v2/storage/vpc/getAttachmentsList", http.MethodGet, nil, http.StatusOK, content, nil)
	volumeAttachService := instances.NewIKSVolumeAttachmentManager(client)

	template := &models.VolumeAttachment{
		ID:         "volumeattachmentid",
		Name:       "attachment-volume-id",
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

	volumeAttachmentsList, err := volumeAttachService.ListVolumeAttachments(template, logger)

	assert.NoError(t, err)
	assert.NotNil(t, volumeAttachmentsList)
}
