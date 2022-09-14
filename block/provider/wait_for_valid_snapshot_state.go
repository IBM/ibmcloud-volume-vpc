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
	"time"

	"github.com/IBM/ibmcloud-volume-interface/lib/metrics"
	userError "github.com/IBM/ibmcloud-volume-vpc/common/messages"
	"github.com/IBM/ibmcloud-volume-vpc/common/vpcclient/models"
	"go.uber.org/zap"
)

const (
	validSnapshotStatus = "stable"
)

// WaitForValidSnapshotState checks the snapshot for valid status
func WaitForValidSnapshotState(vpcs *VPCSession, snapshotID string) (err error) {
	vpcs.Logger.Debug("Entry of WaitForValidSnapshotState method...")
	defer vpcs.Logger.Debug("Exit from WaitForValidSnapshotState method...")
	defer metrics.UpdateDurationFromStart(vpcs.Logger, "WaitForValidSnapshotState", time.Now())

	vpcs.Logger.Info("Getting snapshot details from VPC provider...", zap.Reflect("SnapshotID", snapshotID))

	var snapshot *models.Snapshot
	err = retry(vpcs.Logger, func() error {
		snapshot, err = vpcs.Apiclient.SnapshotService().GetSnapshot(snapshotID, vpcs.Logger)
		if err != nil {
			return err
		}
		vpcs.Logger.Info("Getting snapshot details from VPC provider...", zap.Reflect("snapshot", snapshot))
		if snapshot != nil && snapshot.LifecycleState == validSnapshotStatus {
			vpcs.Logger.Info("Snapshot got valid (available) state", zap.Reflect("SnapshotDetails", snapshot))
			return nil
		}
		return userError.GetUserError("SnapshotNotInValidState", err, snapshotID)
	})

	if err != nil {
		vpcs.Logger.Info("Snapshot could not get valid (available) state", zap.Reflect("SnapshotDetails", snapshot))
		return userError.GetUserError("SnapshotNotInValidState", err, snapshotID)
	}

	return nil
}
