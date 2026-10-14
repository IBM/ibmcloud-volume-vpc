/**
 * Copyright 2026 IBM Corp.
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

package provider

import (
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/IBM/ibmcloud-volume-vpc/common/vpcclient/models"
	"github.com/IBM/ibmcloud-volume-vpc/common/vpcclient/riaas/test"
	"github.com/IBM/ibmcloud-volume-vpc/common/vpcclient/vpcvolume"
	serviceFakes "github.com/IBM/ibmcloud-volume-vpc/common/vpcclient/vpcvolume/fakes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

// newGroupResponse returns a VPC consistency group that references two member snapshots.
func newGroupResponse(lifecycleState string) *models.SnapshotConsistencyGroupResponse {
	createdAt := time.Now()
	return &models.SnapshotConsistencyGroupResponse{
		ID:             "group-snapshot-id",
		CRN:            "group-snapshot-crn",
		LifecycleState: lifecycleState,
		CreatedAt:      &createdAt,
		Snapshots: []models.SnapshotReference{
			{ID: "snapshot-id-1", CRN: "snapshot-crn-1"},
			{ID: "snapshot-id-2", CRN: "snapshot-crn-2"},
		},
	}
}

// newMemberList returns full details for both members, with the second member in memberState.
func newMemberList(secondMemberState string) *models.SnapshotList {
	return &models.SnapshotList{Snapshots: []*models.Snapshot{
		{ID: "snapshot-id-1", CRN: "snapshot-crn-1", LifecycleState: snapshotReadyState, SourceVolume: &models.SourceVolume{ID: "volume-id-1"}},
		{ID: "snapshot-id-2", CRN: "snapshot-crn-2", LifecycleState: secondMemberState, SourceVolume: &models.SourceVolume{ID: "volume-id-2"}},
	}}
}

func TestGetGroupSnapshot(t *testing.T) {
	notFoundErr := &models.Error{Errors: []models.ErrorItem{
		{Code: models.ErrorCode("snapshot_consistency_groups_not_found")},
	}}
	internalErr := &models.Error{Errors: []models.ErrorItem{
		{Code: "internal_error", Status: "500 Internal Server Error"},
	}}

	testCases := []struct {
		name       string
		groupID    string
		group      *models.SnapshotConsistencyGroupResponse
		groupErr   error
		members    *models.SnapshotList
		maxRetries int // 0 uses the defaults; >0 keeps error cases fast

		wantErrCode     string // provider reason code; empty when no provider error is expected
		wantErrText     string // plain error text for errors that are not provider messages
		wantBackendText string
		wantReady       bool
		wantVolumeIDs   []string
		wantGroupCalls  int
		wantListCalls   int
	}{
		{
			name:           "stable group with ready members",
			groupID:        "group-snapshot-id",
			group:          newGroupResponse(snapshotReadyState),
			members:        newMemberList(snapshotReadyState),
			wantReady:      true,
			wantVolumeIDs:  []string{"volume-id-1", "volume-id-2"},
			wantGroupCalls: 1,
			wantListCalls:  1,
		},
		{
			name:           "pending member makes the group not ready",
			groupID:        "group-snapshot-id",
			group:          newGroupResponse(snapshotReadyState),
			members:        newMemberList("pending"),
			wantReady:      false,
			wantVolumeIDs:  []string{"volume-id-1", "volume-id-2"},
			wantGroupCalls: 1,
			wantListCalls:  1,
		},
		{
			name:           "pending group is not ready even when members are ready",
			groupID:        "group-snapshot-id",
			group:          newGroupResponse("pending"),
			members:        newMemberList(snapshotReadyState),
			wantReady:      false,
			wantVolumeIDs:  []string{"volume-id-1", "volume-id-2"},
			wantGroupCalls: 1,
			wantListCalls:  1,
		},
		{
			name:            "backend error is wrapped as GroupSnapshotRetrievalFailed",
			groupID:         "group-snapshot-id",
			groupErr:        internalErr,
			maxRetries:      1,
			wantErrCode:     "GroupSnapshotRetrievalFailed",
			wantBackendText: "internal_error",
			wantGroupCalls:  1,
			wantListCalls:   0,
		},
		{
			name:            "not-found error is not retried",
			groupID:         "group-snapshot-id",
			groupErr:        notFoundErr,
			maxRetries:      3,
			wantErrCode:     "GroupSnapshotRetrievalFailed",
			wantBackendText: "snapshot_consistency_groups_not_found",
			wantGroupCalls:  1,
			wantListCalls:   0,
		},
		{
			name:           "empty backend response fails before listing members",
			groupID:        "group-snapshot-id",
			wantErrText:    "empty snapshot consistency group response",
			wantGroupCalls: 1,
			wantListCalls:  0,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			logger, teardown := GetTestLogger(t)
			defer teardown()

			vpcs, uc, _, err := GetTestOpenSession(t, logger)
			require.NoError(t, err)

			// Must run after the session opens: opening it resets the retry settings from config.
			if tc.maxRetries > 0 {
				limitRetries(t, tc.maxRetries)
			}

			groupService := &fakeSnapshotConsistencyGroupManager{getResult: tc.group, getErr: tc.groupErr}
			snapshotService := &serviceFakes.SnapshotManager{}
			snapshotService.ListSnapshotsReturns(tc.members, nil)
			uc.SnapshotConsistencyGroupServiceReturns(groupService)
			uc.SnapshotServiceReturns(snapshotService)

			result, err := vpcs.GetGroupSnapshot(tc.groupID)

			assert.Equal(t, tc.wantGroupCalls, groupService.getCalls)
			assert.Equal(t, tc.wantListCalls, snapshotService.ListSnapshotsCallCount())
			assert.Equal(t, tc.groupID, groupService.getGroupID)

			switch {
			case tc.wantErrCode != "":
				assert.Nil(t, result)
				requireProviderError(t, err, tc.wantErrCode, tc.groupID, tc.wantBackendText)
			case tc.wantErrText != "":
				assert.Nil(t, result)
				require.Error(t, err)
				assert.Contains(t, err.Error(), tc.wantErrText)
			default:
				require.NoError(t, err)
				require.NotNil(t, result)
				assert.Equal(t, tc.group.ID, result.GroupSnapshotID)
				assert.Equal(t, tc.group.CRN, result.GroupSnapshotCRN)
				assert.Equal(t, tc.wantReady, result.ReadyToUse)
				var gotVolumeIDs []string
				for _, member := range result.Snapshots {
					gotVolumeIDs = append(gotVolumeIDs, member.VolumeID)
				}
				assert.Equal(t, tc.wantVolumeIDs, gotVolumeIDs)
				_, _, filters, _ := snapshotService.ListSnapshotsArgsForCall(0)
				assert.Equal(t, tc.group.ID, filters.SnapshotConsistencyGroupID)
			}
		})
	}
}

func TestGetGroupSnapshotByName(t *testing.T) {
	notFoundErr := &models.Error{Errors: []models.ErrorItem{
		{Code: models.ErrorCode("snapshot_consistency_groups_not_found")},
	}}
	forbiddenErr := &models.Error{Errors: []models.ErrorItem{
		{Code: "snapshots_not_authorized", Status: "403 Forbidden"},
	}}

	testCases := []struct {
		name          string
		groupName     string
		resourceGroup string
		group         *models.SnapshotConsistencyGroupResponse
		groupErr      error
		members       *models.SnapshotList
		maxRetries    int // 0 uses the defaults; >0 keeps error cases fast

		wantErrCode       string // provider reason code; empty when no error is expected
		wantBackendText   string
		wantNilResult     bool
		wantVolumeIDs     []string
		wantByNameCalls   int
		wantListCalls     int
		wantQueriedByName string
	}{
		{
			name:              "name and resource group are forwarded to the backend",
			groupName:         "my-group",
			resourceGroup:     "my-resource-group",
			group:             newGroupResponse(snapshotReadyState),
			members:           newMemberList(snapshotReadyState),
			wantVolumeIDs:     []string{"volume-id-1", "volume-id-2"},
			wantByNameCalls:   1,
			wantListCalls:     1,
			wantQueriedByName: "my-group",
		},
		{
			name:              "no matching group returns nil without an error",
			groupName:         "nonexistent-group",
			resourceGroup:     "my-resource-group",
			wantNilResult:     true,
			wantByNameCalls:   1,
			wantListCalls:     0,
			wantQueriedByName: "nonexistent-group",
		},
		{
			name:            "empty name is rejected without calling the backend",
			groupName:       "",
			resourceGroup:   "my-resource-group",
			wantErrCode:     "InvalidGroupSnapshotName",
			wantByNameCalls: 0,
			wantListCalls:   0,
		},
		{
			name:              "backend error is wrapped as GroupSnapshotNameLookupFailed",
			groupName:         "group-name",
			resourceGroup:     "my-resource-group",
			groupErr:          forbiddenErr,
			maxRetries:        1,
			wantErrCode:       "GroupSnapshotNameLookupFailed",
			wantBackendText:   "snapshots_not_authorized",
			wantByNameCalls:   1,
			wantListCalls:     0,
			wantQueriedByName: "group-name",
		},
		{
			name:              "not-found error is not retried",
			groupName:         "group-name",
			resourceGroup:     "my-resource-group",
			groupErr:          notFoundErr,
			maxRetries:        3,
			wantErrCode:       "GroupSnapshotNameLookupFailed",
			wantBackendText:   "snapshot_consistency_groups_not_found",
			wantByNameCalls:   1,
			wantListCalls:     0,
			wantQueriedByName: "group-name",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			logger, teardown := GetTestLogger(t)
			defer teardown()

			vpcs, uc, _, err := GetTestOpenSession(t, logger)
			require.NoError(t, err)

			// Must run after the session opens: opening it resets the retry settings from config.
			if tc.maxRetries > 0 {
				limitRetries(t, tc.maxRetries)
			}

			groupService := &fakeSnapshotConsistencyGroupManager{getByNameResult: tc.group, getByNameErr: tc.groupErr}
			snapshotService := &serviceFakes.SnapshotManager{}
			snapshotService.ListSnapshotsReturns(tc.members, nil)
			uc.SnapshotConsistencyGroupServiceReturns(groupService)
			uc.SnapshotServiceReturns(snapshotService)

			result, err := vpcs.GetGroupSnapshotByName(tc.groupName, tc.resourceGroup)

			assert.Equal(t, tc.wantByNameCalls, groupService.getByNameCalls)
			assert.Equal(t, tc.wantListCalls, snapshotService.ListSnapshotsCallCount())
			if tc.wantByNameCalls > 0 {
				assert.Equal(t, tc.wantQueriedByName, groupService.getByNameName)
				assert.Equal(t, tc.resourceGroup, groupService.getByNameResourceGroupID)
			}

			switch {
			case tc.wantErrCode != "":
				assert.Nil(t, result)
				requireProviderError(t, err, tc.wantErrCode, "", tc.wantBackendText)
			case tc.wantNilResult:
				require.NoError(t, err)
				assert.Nil(t, result)
			default:
				require.NoError(t, err)
				require.NotNil(t, result)
				assert.Equal(t, tc.group.ID, result.GroupSnapshotID)
				var gotVolumeIDs []string
				for _, member := range result.Snapshots {
					gotVolumeIDs = append(gotVolumeIDs, member.VolumeID)
				}
				assert.Equal(t, tc.wantVolumeIDs, gotVolumeIDs)
			}
		})
	}
}

// pagedMember returns full details for one group member.
func pagedMember(id string, sourceVolumeID string, state string) *models.Snapshot {
	snapshot := &models.Snapshot{ID: id, CRN: "crn-" + id, LifecycleState: state}
	if sourceVolumeID != "" {
		snapshot.SourceVolume = &models.SourceVolume{ID: sourceVolumeID}
	}
	return snapshot
}

// pageLink builds a "next" link as the VPC API returns it.
func pageLink(start string) *models.HReference {
	return &models.HReference{Href: "https://eu-gb.iaas.cloud.ibm.com/v1/snapshots?limit=100&start=" + start}
}

// TestGetGroupSnapshotMemberPagination checks that every page of member snapshots is read,
// and that a problem on any page returns an error instead of a partial group.
func TestGetGroupSnapshotMemberPagination(t *testing.T) {
	// The group references members A and B, so a result with only A must never be accepted.
	newGroup := func() *models.SnapshotConsistencyGroupResponse {
		return &models.SnapshotConsistencyGroupResponse{
			ID: "group-snapshot-id", LifecycleState: snapshotReadyState,
			Snapshots: []models.SnapshotReference{
				{ID: "snapshot-a", CRN: "crn-snapshot-a"},
				{ID: "snapshot-b", CRN: "crn-snapshot-b"},
			},
		}
	}
	memberA := pagedMember("snapshot-a", "volume-a", snapshotReadyState)
	memberB := pagedMember("snapshot-b", "volume-b", snapshotReadyState)
	// Both errors are retryable by retry(): internal_error is not skip-listed, and a plain error is not a *models.Error.
	// (snapshots_service_unavailable is skip-listed, so it would not prove that no retry happens.)
	backendErr := &models.Error{Trace: "page-2-trace", Errors: []models.ErrorItem{
		{Code: models.ErrorCode("internal_error"), Status: "500 Internal Server Error"},
	}}
	transportErr := errors.New("page 2 connection reset")

	type page struct {
		list *models.SnapshotList
		err  error
	}
	testCases := []struct {
		name  string
		pages map[string]page // keyed by the start token the driver asks for; "" is the first page

		wantErrCode     string // empty means success
		wantBackendText string
		wantErrText     string
		wantStarts      []string // start tokens requested, in order
		wantVolumeIDs   []string
		wantReady       bool
	}{
		{
			name:          "single page without Next",
			pages:         map[string]page{"": {list: &models.SnapshotList{Snapshots: []*models.Snapshot{memberA, memberB}}}},
			wantStarts:    []string{""},
			wantVolumeIDs: []string{"volume-a", "volume-b"},
			wantReady:     true,
		},
		{
			name: "two pages contain all members",
			pages: map[string]page{
				"":       {list: &models.SnapshotList{Snapshots: []*models.Snapshot{memberA}, Next: pageLink("page-2")}},
				"page-2": {list: &models.SnapshotList{Snapshots: []*models.Snapshot{memberB}}},
			},
			wantStarts:    []string{"", "page-2"},
			wantVolumeIDs: []string{"volume-a", "volume-b"},
			wantReady:     true,
		},
		{
			name: "three pages with an empty middle page",
			pages: map[string]page{
				"":       {list: &models.SnapshotList{Snapshots: []*models.Snapshot{memberA}, Next: pageLink("page-2")}},
				"page-2": {list: &models.SnapshotList{Next: pageLink("page-3")}},
				"page-3": {list: &models.SnapshotList{Snapshots: []*models.Snapshot{memberB}}},
			},
			wantStarts:    []string{"", "page-2", "page-3"},
			wantVolumeIDs: []string{"volume-a", "volume-b"},
			wantReady:     true,
		},
		{
			name: "page 2 failure returns the backend error and no group",
			pages: map[string]page{
				"":       {list: &models.SnapshotList{Snapshots: []*models.Snapshot{memberA}, Next: pageLink("page-2")}},
				"page-2": {err: backendErr},
			},
			wantErrCode:     "GroupSnapshotMemberLookupFailed",
			wantBackendText: "internal_error",
			wantStarts:      []string{"", "page-2"},
		},
		{
			name: "page 2 transport failure returns the error and no group",
			pages: map[string]page{
				"":       {list: &models.SnapshotList{Snapshots: []*models.Snapshot{memberA}, Next: pageLink("page-2")}},
				"page-2": {err: transportErr},
			},
			wantErrCode:     "GroupSnapshotMemberLookupFailed",
			wantBackendText: "connection reset",
			wantStarts:      []string{"", "page-2"},
		},
		{
			name: "nil page 2 is an error",
			pages: map[string]page{
				"":       {list: &models.SnapshotList{Snapshots: []*models.Snapshot{memberA}, Next: pageLink("page-2")}},
				"page-2": {},
			},
			wantErrCode:     "GroupSnapshotMemberLookupFailed",
			wantBackendText: "empty snapshot page",
			wantStarts:      []string{"", "page-2"},
		},
		{
			name:            "nil first page is an error",
			pages:           map[string]page{"": {}},
			wantErrCode:     "GroupSnapshotMemberLookupFailed",
			wantBackendText: "empty snapshot page",
			wantStarts:      []string{""},
		},
		{
			name: "repeated token: a page points to itself",
			pages: map[string]page{
				"":       {list: &models.SnapshotList{Snapshots: []*models.Snapshot{memberA}, Next: pageLink("page-2")}},
				"page-2": {list: &models.SnapshotList{Snapshots: []*models.Snapshot{memberB}, Next: pageLink("page-2")}},
			},
			wantErrCode:     "GroupSnapshotMemberLookupFailed",
			wantBackendText: "repeated page token",
			wantStarts:      []string{"", "page-2"},
		},
		{
			name: "token cycle A to B to A",
			pages: map[string]page{
				"":       {list: &models.SnapshotList{Snapshots: []*models.Snapshot{memberA}, Next: pageLink("page-a")}},
				"page-a": {list: &models.SnapshotList{Next: pageLink("page-b")}},
				"page-b": {list: &models.SnapshotList{Next: pageLink("page-a")}},
			},
			wantErrCode:     "GroupSnapshotMemberLookupFailed",
			wantBackendText: "repeated page token",
			wantStarts:      []string{"", "page-a", "page-b"},
		},
		{
			name: "empty next URL",
			pages: map[string]page{
				"": {list: &models.SnapshotList{Snapshots: []*models.Snapshot{memberA}, Next: &models.HReference{}}},
			},
			wantErrCode:     "GroupSnapshotMemberLookupFailed",
			wantBackendText: "no start token",
			wantStarts:      []string{""},
		},
		{
			name: "malformed next URL",
			pages: map[string]page{
				"": {list: &models.SnapshotList{Snapshots: []*models.Snapshot{memberA}, Next: &models.HReference{Href: "https://host/v1/snapshots?start=%zz"}}},
			},
			wantErrCode:     "GroupSnapshotMemberLookupFailed",
			wantBackendText: "invalid",
			wantStarts:      []string{""},
		},
		{
			name: "next URL without a start parameter",
			pages: map[string]page{
				"": {list: &models.SnapshotList{Snapshots: []*models.Snapshot{memberA}, Next: &models.HReference{Href: "https://host/v1/snapshots?limit=100"}}},
			},
			wantErrCode:     "GroupSnapshotMemberLookupFailed",
			wantBackendText: "no start token",
			wantStarts:      []string{""},
		},
		{
			name: "next URL with an empty start value",
			pages: map[string]page{
				"": {list: &models.SnapshotList{Snapshots: []*models.Snapshot{memberA}, Next: &models.HReference{Href: "https://host/v1/snapshots?start=&limit=100"}}},
			},
			wantErrCode:     "GroupSnapshotMemberLookupFailed",
			wantBackendText: "no start token",
			wantStarts:      []string{""},
		},
		{
			name: "all pages read but a member has no source volume",
			pages: map[string]page{
				"":       {list: &models.SnapshotList{Snapshots: []*models.Snapshot{memberA}, Next: pageLink("page-2")}},
				"page-2": {list: &models.SnapshotList{Snapshots: []*models.Snapshot{pagedMember("snapshot-b", "", snapshotReadyState)}}},
			},
			wantErrCode: "GroupSnapshotMemberDetailsUnavailable",
			wantStarts:  []string{"", "page-2"},
		},
		{
			name: "all pages read but a referenced member never appears",
			pages: map[string]page{
				"": {list: &models.SnapshotList{Snapshots: []*models.Snapshot{memberA}}},
			},
			wantErrCode: "GroupSnapshotMemberDetailsUnavailable",
			wantStarts:  []string{""},
		},
		{
			name: "complete members that are not ready are still a valid response",
			pages: map[string]page{
				"":       {list: &models.SnapshotList{Snapshots: []*models.Snapshot{memberA}, Next: pageLink("page-2")}},
				"page-2": {list: &models.SnapshotList{Snapshots: []*models.Snapshot{pagedMember("snapshot-b", "volume-b", "pending")}}},
			},
			wantStarts:    []string{"", "page-2"},
			wantVolumeIDs: []string{"volume-a", "volume-b"},
			wantReady:     false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			logger, teardown := GetTestLogger(t)
			defer teardown()

			session, client, _, err := GetTestOpenSession(t, logger)
			require.NoError(t, err)
			// Several attempts are allowed on purpose: if member listing ever starts using the
			// retry loop, a failed page is requested again and the checks below fail.
			limitRetries(t, 3)

			group := newGroup()
			client.SnapshotConsistencyGroupServiceReturns(&fakeSnapshotConsistencyGroupManager{getResult: group})
			snapshotService := &serviceFakes.SnapshotManager{}
			calls := 0
			snapshotService.ListSnapshotsCalls(func(limit int, start string, filters *models.LisSnapshotFilters, _ *zap.Logger) (*models.SnapshotList, error) {
				// Ends runaway pagination with a clear failure instead of hanging the test run.
				calls++
				if calls > len(tc.pages)+1 {
					return nil, errors.New("runaway pagination: more requests than pages")
				}
				resp, known := tc.pages[start]
				require.True(t, known, "unexpected start token %q", start)
				return resp.list, resp.err
			})
			client.SnapshotServiceReturns(snapshotService)

			result, err := session.GetGroupSnapshot(group.ID)

			// Every request must use the same limit and group filter; only the start token changes.
			var gotStarts []string
			for i := 0; i < snapshotService.ListSnapshotsCallCount(); i++ {
				limit, start, filters, _ := snapshotService.ListSnapshotsArgsForCall(i)
				assert.Equal(t, maxLimit, limit)
				assert.Equal(t, group.ID, filters.SnapshotConsistencyGroupID)
				gotStarts = append(gotStarts, start)
			}
			assert.Equal(t, tc.wantStarts, gotStarts)
			// No page may be requested twice, so a failed page is never retried inside the loop.
			requested := map[string]int{}
			for _, start := range gotStarts {
				requested[start]++
				assert.Equal(t, 1, requested[start], "start token %q was requested more than once", start)
			}

			if tc.wantErrCode != "" {
				assert.Nil(t, result, "a failed listing must never return a group")
				requireProviderError(t, err, tc.wantErrCode, group.ID, tc.wantBackendText)
				return
			}
			require.NoError(t, err)
			require.NotNil(t, result)
			var gotVolumeIDs []string
			for _, member := range result.Snapshots {
				gotVolumeIDs = append(gotVolumeIDs, member.VolumeID)
			}
			assert.Equal(t, tc.wantVolumeIDs, gotVolumeIDs)
			assert.Equal(t, tc.wantReady, result.ReadyToUse)
		})
	}
}

// The token is decoded from the next link and encoded once in the next HTTP request.
// This uses the real SnapshotService over HTTP, so the helper, the loop and the client work together.
func TestGetGroupSnapshotEncodesNextPageTokenOnce(t *testing.T) {
	logger, teardown := GetTestLogger(t)
	defer teardown()

	session, client, _, err := GetTestOpenSession(t, logger)
	require.NoError(t, err)

	var rawQueries []string
	mux, httpClient, closeServer := test.SetupServer(t)
	defer closeServer()
	mux.HandleFunc(vpcvolume.Version+"/snapshots", func(w http.ResponseWriter, r *http.Request) {
		rawQueries = append(rawQueries, r.URL.RawQuery)
		if r.URL.Query().Get("start") == "" {
			// The token is URL-encoded inside the link, as the VPC API returns it.
			_, _ = w.Write([]byte(`{"snapshots":[{"id":"snapshot-a","crn":"crn-snapshot-a","lifecycle_state":"stable","source_volume":{"id":"volume-a"}}],` +
				`"next":{"href":"https://host/v1/snapshots?start=ab%2Bc%3D%3D&limit=100"}}`))
			return
		}
		_, _ = w.Write([]byte(`{"snapshots":[{"id":"snapshot-b","crn":"crn-snapshot-b","lifecycle_state":"stable","source_volume":{"id":"volume-b"}}]}`))
	})
	client.SnapshotServiceReturns(vpcvolume.NewSnapshotManager(httpClient))
	client.SnapshotConsistencyGroupServiceReturns(&fakeSnapshotConsistencyGroupManager{getResult: &models.SnapshotConsistencyGroupResponse{
		ID: "group-snapshot-id", LifecycleState: snapshotReadyState,
		Snapshots: []models.SnapshotReference{{ID: "snapshot-a", CRN: "crn-snapshot-a"}, {ID: "snapshot-b", CRN: "crn-snapshot-b"}},
	}})

	result, err := session.GetGroupSnapshot("group-snapshot-id")

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Len(t, result.Snapshots, 2)
	require.Len(t, rawQueries, 2)
	assert.Contains(t, rawQueries[0], "snapshot_consistency_group.id=group-snapshot-id")
	assert.Contains(t, rawQueries[0], "limit=100")
	assert.NotContains(t, rawQueries[0], "start=")
	// Same filter and limit on page 2, and the token is encoded once.
	assert.Contains(t, rawQueries[1], "snapshot_consistency_group.id=group-snapshot-id")
	assert.Contains(t, rawQueries[1], "limit=100")
	assert.Contains(t, rawQueries[1], "start=ab%2Bc%3D%3D")
	assert.NotContains(t, rawQueries[1], "%252B", "token must not be double-encoded")
}

// Create, get by ID and get by name share getGroupSnapshotWithMembers, so each of them
// must read every member page, not only the first one.
func TestGroupSnapshotMemberPaginationForEveryOperation(t *testing.T) {
	for _, operation := range groupSnapshotOperations {
		t.Run(operation.name, func(t *testing.T) {
			logger, teardown := GetTestLogger(t)
			defer teardown()
			session, client, _, err := GetTestOpenSession(t, logger)
			require.NoError(t, err)

			group := &models.SnapshotConsistencyGroupResponse{
				ID: "group-snapshot-id", LifecycleState: snapshotReadyState,
				Snapshots: []models.SnapshotReference{
					{ID: "snapshot-a", CRN: "crn-snapshot-a"},
					{ID: "snapshot-b", CRN: "crn-snapshot-b"},
				},
			}
			client.SnapshotConsistencyGroupServiceReturns(&fakeSnapshotConsistencyGroupManager{
				createResult: group, getResult: group, getByNameResult: group,
			})
			pages := map[string]*models.SnapshotList{
				"":       {Snapshots: []*models.Snapshot{pagedMember("snapshot-a", "volume-a", snapshotReadyState)}, Next: pageLink("page-2")},
				"page-2": {Snapshots: []*models.Snapshot{pagedMember("snapshot-b", "volume-b", snapshotReadyState)}},
			}
			snapshotService := &serviceFakes.SnapshotManager{}
			snapshotService.ListSnapshotsCalls(func(_ int, start string, _ *models.LisSnapshotFilters, _ *zap.Logger) (*models.SnapshotList, error) {
				page, known := pages[start]
				require.True(t, known, "unexpected start token %q", start)
				return page, nil
			})
			client.SnapshotServiceReturns(snapshotService)

			result, err := operation.call(session)

			require.NoError(t, err)
			require.NotNil(t, result)
			var gotVolumeIDs, gotStarts []string
			for _, member := range result.Snapshots {
				gotVolumeIDs = append(gotVolumeIDs, member.VolumeID)
			}
			assert.Equal(t, []string{"volume-a", "volume-b"}, gotVolumeIDs)
			for i := 0; i < snapshotService.ListSnapshotsCallCount(); i++ {
				limit, start, filters, _ := snapshotService.ListSnapshotsArgsForCall(i)
				assert.Equal(t, maxLimit, limit)
				assert.Equal(t, group.ID, filters.SnapshotConsistencyGroupID)
				gotStarts = append(gotStarts, start)
			}
			assert.Equal(t, []string{"", "page-2"}, gotStarts)
		})
	}
}

// An empty group ID would make the client drop the group filter and list every snapshot
// in the account, so the provider must reject it before any snapshot is listed.
func TestGroupSnapshotWithoutIDIsRejectedBeforeListing(t *testing.T) {
	for _, operation := range groupSnapshotOperations {
		t.Run(operation.name, func(t *testing.T) {
			logger, teardown := GetTestLogger(t)
			defer teardown()
			session, client, _, err := GetTestOpenSession(t, logger)
			require.NoError(t, err)

			group := &models.SnapshotConsistencyGroupResponse{LifecycleState: snapshotReadyState}
			client.SnapshotConsistencyGroupServiceReturns(&fakeSnapshotConsistencyGroupManager{
				createResult: group, getResult: group, getByNameResult: group,
			})
			snapshotService := &serviceFakes.SnapshotManager{}
			client.SnapshotServiceReturns(snapshotService)

			result, err := operation.call(session)

			require.Error(t, err)
			assert.Contains(t, err.Error(), "without an ID")
			assert.Nil(t, result)
			assert.Equal(t, 0, snapshotService.ListSnapshotsCallCount(), "no snapshot may be listed without a group ID")
		})
	}
}
