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

// Package main ...
package main

import (
	"flag"
	"fmt"

	"github.com/IBM/ibmcloud-volume-interface/lib/provider"
	"go.uber.org/zap"
)

var (
	volumeID  = flag.String("volume_id", "", "Volume ID")
	volumeCRN = flag.String("volume_crn", "", "Volume CRN")
	clusterID = flag.String("cluster", "", "Cluster ID")
)

var volumeReq provider.Volume

// VolumeManager ...
type VolumeManager struct {
	Session   provider.Session
	Logger    *zap.Logger
	RequestID string
}

// NewVolumeManager ...
func NewVolumeManager(session provider.Session, logger *zap.Logger, requestID string) *VolumeManager {
	return &VolumeManager{
		Session:   session,
		Logger:    logger,
		RequestID: requestID,
	}
}

// UpdateVolume ...
func (vam *VolumeManager) UpdateVolume() {
	vam.setupVolumeRequest()
	err := vam.Session.UpdateVolume(volumeReq)
	if err != nil {
		err1 := updateRequestID(err, vam.RequestID)
		vam.Logger.Error("Failed to update the volume", zap.Error(err), zap.Error(err1))
		return
	}
	fmt.Println("Volume update", err)
}

func (vam *VolumeManager) setupVolumeRequest() {
	/*	fmt.Printf("Enter the volume id: ")
		_, _ = fmt.Scanf("%s", &volumeID)
		fmt.Printf("Enter the provider: ")
		_, _ = fmt.Scanf("%s", &instanceID)
		fmt.Printf("Enter the cluster id: ")
		_, _ = fmt.Scanf("%s", &clusterID)*/
	capacity := 30
	//iops := "10"
	volumeReq = provider.Volume{
		VolumeID: *volumeID,
		Capacity: &capacity,
		//Iops:     &iops,
		Provider:   "vpc-classic",
		VolumeType: "block",
	}
	volumeReq.Attributes = map[string]string{"clusterid": *clusterID, "reclaimpolicy": "Delete"}
	volumeReq.Tags = []string{"clusterid:" + *clusterID, "reclaimpolicy:Delete"}
	volumeReq.VPCVolume.Profile = &provider.Profile{Name: "general-purpose"}
	volumeReq.CRN = *volumeCRN
}
