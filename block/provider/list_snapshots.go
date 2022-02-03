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
	"fmt"
	"strings"
	"time"

	"github.com/IBM/ibmcloud-volume-interface/lib/metrics"
	"github.com/IBM/ibmcloud-volume-interface/lib/provider"
	userError "github.com/IBM/ibmcloud-volume-vpc/common/messages"
	"github.com/IBM/ibmcloud-volume-vpc/common/vpcclient/models"
	"go.uber.org/zap"
)

// ListSnapshots list all snapshots
func (vpcs *VPCSession) ListSnapshots(limit int, start string, tags map[string]string) (*provider.SnapshotList, error) {
	vpcs.Logger.Info("Entry ListeSnapshots")
	defer vpcs.Logger.Info("Exit ListSnapshots")
	defer metrics.UpdateDurationFromStart(vpcs.Logger, "ListSnapshots", time.Now())

	if limit < 0 {
		return nil, userError.GetUserError("InvalidListSnapshotLimit", nil, limit)
	}

	if limit > maxLimit {
		vpcs.Logger.Warn(fmt.Sprintf("listSnapshots requested max entries of %v, supports values <= %v so defaulting value back to %v", limit, maxLimit, maxLimit))
		limit = maxLimit
	}

	filters := &models.LisSnapshotFilters{
		ResourceGroupID: tags["resource_group.id"],
		Name:            tags["name"],
		SourceVolumeID:  tags["source_volume.id"],
	}

	vpcs.Logger.Info("Getting snapshot list from VPC provider...", zap.Reflect("start", start), zap.Reflect("filters", filters))

	var snapshots *models.SnapshotList
	var err error
	err = retry(vpcs.Logger, func() error {
		snapshots, err = vpcs.Apiclient.SnapshotService().ListSnapshots(limit, start, filters, vpcs.Logger)
		return err
	})

	if err != nil {
		return nil, userError.GetUserError("ListSnapshotsFailed", err)
	}

	vpcs.Logger.Info("Successfully retrieved snapshot list from VPC backend", zap.Reflect("SnapshotList", snapshots))

	var respSnapshotList = &provider.SnapshotList{}
	if snapshots != nil {
		if snapshots.Next != nil {
			var next string
			// "Next":{"href":"https://eu-gb.iaas.cloud.ibm.com/v1/snapshots?start=3e898aa7-ac71-4323-952d-a8d741c65a68\u0026limit=1\u0026zone.name=eu-gb-1"}
			if strings.Contains(snapshots.Next.Href, "start=") {
				next = strings.Split(strings.Split(snapshots.Next.Href, "start=")[1], "\u0026")[0]
			} else {
				vpcs.Logger.Warn("snapshots.Next.Href is not in expected format", zap.Reflect("snapshots.Next.Href", snapshots.Next.Href))
			}
			respSnapshotList.Next = next
		}

		snapshotslist := snapshots.Snapshots
		if len(snapshotslist) > 0 {
			for _, snapItem := range snapshotslist {
				respSnapshot := &provider.Snapshot{
					VolumeID:             snapItem.SourceVolume.ID,
					SnapshotID:           snapItem.ID,
					SnapshotCreationTime: *snapItem.CreatedAt,
					SnapshotSize:         GiBToBytes(snapItem.Size),
					VPC:                  &provider.VPC{Href: snapItem.Href},
				}
				if snapItem.LifecycleState == snapshotReadyState {
					respSnapshot.ReadyToUse = true
				} else {
					respSnapshot.ReadyToUse = false
				}
				respSnapshotList.Snapshots = append(respSnapshotList.Snapshots, respSnapshot)
			}
		}
	}
	return respSnapshotList, err
}
