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
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"testing"
	"time"

	providerError "github.com/IBM/ibmcloud-volume-interface/lib/utils"
	"github.com/IBM/ibmcloud-volume-vpc/common/vpcclient/models"
	"github.com/IBM/ibmcloud-volume-vpc/common/vpcclient/riaas/test"
	"github.com/IBM/ibmcloud-volume-vpc/common/vpcclient/vpcvolume"
	serviceFakes "github.com/IBM/ibmcloud-volume-vpc/common/vpcclient/vpcvolume/fakes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// The group API returns references, not the full snapshot details.
func newGroupResponse(lifecycleState string) *models.SnapshotConsistencyGroupResponse {
	createdAt := time.Date(2026, time.January, 2, 3, 4, 5, 0, time.UTC)
	return &models.SnapshotConsistencyGroupResponse{
		ID:             "group-snapshot-id",
		CRN:            "group-snapshot-crn",
		Href:           "group-snapshot-href",
		LifecycleState: lifecycleState,
		CreatedAt:      &createdAt,
		Snapshots: []*models.Snapshot{
			{ID: "snapshot-id-1", CRN: "snapshot-crn-1"},
			{ID: "snapshot-id-2", CRN: "snapshot-crn-2"},
		},
	}
}

func newMemberList(secondMemberState string) *models.SnapshotList {
	createdAt := time.Date(2026, time.January, 2, 3, 4, 6, 0, time.UTC)
	return &models.SnapshotList{
		Snapshots: []*models.Snapshot{
			{
				ID:              "snapshot-id-1",
				CRN:             "snapshot-crn-1",
				Href:            "snapshot-href-1",
				MinimumCapacity: 10,
				CreatedAt:       &createdAt,
				LifecycleState:  snapshotReadyState,
				SourceVolume:    &models.SourceVolume{ID: "volume-id-1"},
			},
			{
				ID:              "snapshot-id-2",
				CRN:             "snapshot-crn-2",
				MinimumCapacity: 20,
				CreatedAt:       &createdAt,
				LifecycleState:  secondMemberState,
				SourceVolume:    &models.SourceVolume{ID: "volume-id-2"},
			},
		},
	}
}

func TestGetGroupSnapshot(t *testing.T) {
	notFoundErr := &models.Error{Errors: []models.ErrorItem{
		{Code: SnapshotConsistencyGroupNotFound},
	}}
	internalErr := &models.Error{Errors: []models.ErrorItem{
		{Code: "internal_error", Status: "500 Internal Server Error"},
	}}

	testCases := []struct {
		name       string
		groupID    string
		group      *models.SnapshotConsistencyGroupResponse
		members    *models.SnapshotList
		groupErr   error
		maxRetries int // 0 keeps the defaults; a positive value sets total attempts, including the first call.

		wantErrCode     string
		wantErrText     string
		wantBackendText string
		wantReady       bool
		wantVolumeIDs   []string
		wantGroupCalls  int
	}{
		{
			name:           "stable group with ready members",
			groupID:        "group-snapshot-id",
			group:          newGroupResponse(snapshotReadyState),
			members:        newMemberList(snapshotReadyState),
			wantReady:      true,
			wantVolumeIDs:  []string{"volume-id-1", "volume-id-2"},
			wantGroupCalls: 1,
		},
		{
			name:           "pending member makes the group not ready",
			groupID:        "group-snapshot-id",
			group:          newGroupResponse(snapshotReadyState),
			members:        newMemberList("pending"),
			wantReady:      false,
			wantVolumeIDs:  []string{"volume-id-1", "volume-id-2"},
			wantGroupCalls: 1,
		},
		{
			name:           "member without a lifecycle state makes the group not ready",
			groupID:        "group-snapshot-id",
			group:          newGroupResponse(snapshotReadyState),
			members:        newMemberList(""),
			wantReady:      false,
			wantVolumeIDs:  []string{"volume-id-1", "volume-id-2"},
			wantGroupCalls: 1,
		},
		{
			name:           "pending group is not ready even when members are ready",
			groupID:        "group-snapshot-id",
			group:          newGroupResponse("pending"),
			members:        newMemberList(snapshotReadyState),
			wantReady:      false,
			wantVolumeIDs:  []string{"volume-id-1", "volume-id-2"},
			wantGroupCalls: 1,
		},
		{
			name:            "backend error is wrapped as GroupSnapshotRetrievalFailed",
			groupID:         "group-snapshot-id",
			groupErr:        internalErr,
			maxRetries:      1,
			wantErrCode:     "GroupSnapshotRetrievalFailed",
			wantErrText:     "group-snapshot-id",
			wantBackendText: "internal_error",
			wantGroupCalls:  1,
		},
		{
			name:            "not-found error is not retried",
			groupID:         "group-snapshot-id",
			groupErr:        notFoundErr,
			maxRetries:      3,
			wantErrCode:     "GroupSnapshotRetrievalFailed",
			wantErrText:     "group-snapshot-id",
			wantBackendText: SnapshotConsistencyGroupNotFound,
			wantGroupCalls:  1,
		},
		{
			name:           "empty backend response is invalid",
			groupID:        "group-snapshot-id",
			wantErrCode:    "InvalidGroupSnapshotResponse",
			wantErrText:    "the response is empty",
			wantGroupCalls: 1,
		},
		{
			name:           "group without member snapshots",
			groupID:        "group-snapshot-id",
			group:          &models.SnapshotConsistencyGroupResponse{ID: "group-snapshot-id"},
			wantErrCode:    "GroupSnapshotMemberDetailsUnavailable",
			wantErrText:    "group-snapshot-id",
			wantGroupCalls: 1,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			logger, teardown := GetTestLogger(t)
			defer teardown()

			vpcs, uc, _, err := GetTestOpenSession(t, logger)
			require.NoError(t, err)

			// Set this after opening the session, which loads the retry settings from config.
			if tc.maxRetries > 0 {
				limitRetries(t, tc.maxRetries)
			}

			groupService := &fakeSnapshotConsistencyGroupManager{getResult: tc.group, getErr: tc.groupErr}
			uc.SnapshotConsistencyGroupServiceReturns(groupService)
			snapshotService := &serviceFakes.SnapshotManager{}
			snapshotService.ListSnapshotsReturns(tc.members, nil)
			uc.SnapshotServiceReturns(snapshotService)

			result, err := vpcs.GetGroupSnapshot(tc.groupID)

			assert.Equal(t, tc.wantGroupCalls, groupService.getCalls)
			if tc.members == nil {
				assert.Zero(t, snapshotService.ListSnapshotsCallCount())
			} else {
				assert.Equal(t, 1, snapshotService.ListSnapshotsCallCount())
			}
			assert.Equal(t, tc.groupID, groupService.getGroupID)

			switch {
			case tc.wantErrCode != "":
				assert.Nil(t, result)
				requireProviderError(t, err, tc.wantErrCode, tc.wantErrText, tc.wantBackendText)
			default:
				require.NoError(t, err)
				require.NotNil(t, result)
				assert.Equal(t, tc.group.ID, result.GroupSnapshotID)
				assert.Equal(t, tc.group.CRN, result.GroupSnapshotCRN)
				assert.Equal(t, tc.wantReady, result.ReadyToUse)
				var gotVolumeIDs []string
				require.Len(t, result.Snapshots, len(tc.group.Snapshots))
				for i, member := range result.Snapshots {
					gotVolumeIDs = append(gotVolumeIDs, member.VolumeID)
					assert.Equal(t, tc.members.Snapshots[i].LifecycleState == snapshotReadyState, member.ReadyToUse)
				}
				assert.Equal(t, tc.wantVolumeIDs, gotVolumeIDs)
			}
		})
	}
}

func TestGetGroupSnapshotByName(t *testing.T) {
	notFoundErr := &models.Error{Errors: []models.ErrorItem{
		{Code: SnapshotConsistencyGroupNotFound},
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
		maxRetries    int // 0 keeps the defaults; a positive value sets total attempts, including the first call.

		wantErrCode       string // Expected provider error code; empty means no error is expected.
		wantBackendText   string
		wantNilResult     bool
		wantVolumeIDs     []string
		wantByNameCalls   int
		wantQueriedByName string
	}{
		{
			name:              "name and resource group are forwarded to the backend",
			groupName:         "my-group",
			resourceGroup:     "my-resource-group",
			group:             newGroupResponse(snapshotReadyState),
			wantVolumeIDs:     []string{"volume-id-1", "volume-id-2"},
			wantByNameCalls:   1,
			wantQueriedByName: "my-group",
		},
		{
			name:              "no matching group returns nil without an error",
			groupName:         "nonexistent-group",
			resourceGroup:     "my-resource-group",
			wantNilResult:     true,
			wantByNameCalls:   1,
			wantQueriedByName: "nonexistent-group",
		},
		{
			name:            "empty name is rejected without calling the backend",
			groupName:       "",
			resourceGroup:   "my-resource-group",
			wantErrCode:     "InvalidGroupSnapshotName",
			wantByNameCalls: 0,
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
			wantQueriedByName: "group-name",
		},
		{
			name:              "not-found error is not retried",
			groupName:         "group-name",
			resourceGroup:     "my-resource-group",
			groupErr:          notFoundErr,
			maxRetries:        3,
			wantErrCode:       "GroupSnapshotNameLookupFailed",
			wantBackendText:   SnapshotConsistencyGroupNotFound,
			wantByNameCalls:   1,
			wantQueriedByName: "group-name",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			logger, teardown := GetTestLogger(t)
			defer teardown()

			vpcs, uc, _, err := GetTestOpenSession(t, logger)
			require.NoError(t, err)

			// Set this after opening the session, which loads the retry settings from config.
			if tc.maxRetries > 0 {
				limitRetries(t, tc.maxRetries)
			}

			groupService := &fakeSnapshotConsistencyGroupManager{getByNameResult: tc.group, getByNameErr: tc.groupErr}
			uc.SnapshotConsistencyGroupServiceReturns(groupService)
			snapshotService := &serviceFakes.SnapshotManager{}
			snapshotService.ListSnapshotsReturns(newMemberList(snapshotReadyState), nil)
			uc.SnapshotServiceReturns(snapshotService)

			result, err := vpcs.GetGroupSnapshotByName(tc.groupName, tc.resourceGroup)

			assert.Equal(t, tc.wantByNameCalls, groupService.getByNameCalls)
			if tc.group == nil {
				assert.Zero(t, snapshotService.ListSnapshotsCallCount())
			} else {
				assert.Equal(t, 1, snapshotService.ListSnapshotsCallCount())
			}
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

func TestGroupSnapshotMemberPages(t *testing.T) {
	group := newGroupResponse(snapshotReadyState)
	members := newMemberList(snapshotReadyState).Snapshots
	pageLink := func(query string) *models.HReference {
		return &models.HReference{Href: "https://vpc.example/v1/snapshots?" + query}
	}
	firstPage := &models.SnapshotList{Snapshots: members[1:], Next: pageLink("start=page-2")}
	lastPage := &models.SnapshotList{Snapshots: members[:1]}
	forbidden := &models.Error{Errors: []models.ErrorItem{{Code: "snapshots_not_authorized", Status: "403 Forbidden"}}}
	testCases := []struct {
		name            string
		pages           []*models.SnapshotList
		tokens          []string
		lastErr         error
		wantCode        string
		wantDetail      string
		wantDescription string
	}{
		{name: "one page in reverse order", pages: []*models.SnapshotList{{Snapshots: []*models.Snapshot{members[1], {ID: "unrelated-snapshot"}, members[0]}}}, tokens: []string{""}},
		{name: "two pages in reverse order", pages: []*models.SnapshotList{firstPage, lastPage}, tokens: []string{"", "page-2"}},
		{name: "empty middle page still has a next page", pages: []*models.SnapshotList{firstPage, {Next: pageLink("start=page-3")}, lastPage}, tokens: []string{"", "page-2", "page-3"}},
		{name: "second page permission failure", pages: []*models.SnapshotList{firstPage, nil}, tokens: []string{"", "page-2"}, lastErr: forbidden, wantCode: "GroupSnapshotMemberLookupFailed", wantDetail: "403 Forbidden"},
		{name: "second page connection failure", pages: []*models.SnapshotList{firstPage, nil}, tokens: []string{"", "page-2"}, lastErr: errors.New("connection reset"), wantCode: "GroupSnapshotMemberLookupFailed", wantDetail: "connection reset"},
		{
			name:            "nil second page",
			pages:           []*models.SnapshotList{firstPage, nil},
			tokens:          []string{"", "page-2"},
			wantCode:        "InvalidGroupSnapshotResponse",
			wantDescription: "empty member snapshot page",
		},
		{name: "next page points to itself", pages: []*models.SnapshotList{firstPage, {Next: pageLink("start=page-2")}}, tokens: []string{"", "page-2"}, wantCode: "GroupSnapshotMemberLookupFailed", wantDetail: "repeated a snapshot page token"},
		{name: "next page points to an earlier token", pages: []*models.SnapshotList{firstPage, {Next: pageLink("start=page-3")}, {Next: pageLink("start=page-2")}}, tokens: []string{"", "page-2", "page-3"}, wantCode: "GroupSnapshotMemberLookupFailed", wantDetail: "repeated a snapshot page token"},
		{name: "next link is not a URL", pages: []*models.SnapshotList{{Next: &models.HReference{Href: "https://[bad"}}}, tokens: []string{""}, wantCode: "GroupSnapshotMemberLookupFailed", wantDetail: "invalid snapshot next-page link"},
		{name: "next token has invalid escaping", pages: []*models.SnapshotList{{Next: pageLink("start=%ZZ")}}, tokens: []string{""}, wantCode: "GroupSnapshotMemberLookupFailed", wantDetail: "invalid snapshot next-page query"},
		{name: "next link has no token", pages: []*models.SnapshotList{{Next: pageLink("limit=100")}}, tokens: []string{""}, wantCode: "GroupSnapshotMemberLookupFailed", wantDetail: "one non-empty start token"},
		{name: "next token is empty", pages: []*models.SnapshotList{{Next: pageLink("start=")}}, tokens: []string{""}, wantCode: "GroupSnapshotMemberLookupFailed", wantDetail: "one non-empty start token"},
		{name: "next link has two tokens", pages: []*models.SnapshotList{{Next: pageLink("start=a&start=b")}}, tokens: []string{""}, wantCode: "GroupSnapshotMemberLookupFailed", wantDetail: "one non-empty start token"},
		{name: "same snapshot on two pages", pages: []*models.SnapshotList{firstPage, {Snapshots: members}}, tokens: []string{"", "page-2"}, wantCode: "GroupSnapshotMemberLookupFailed", wantDetail: "more than once"},
		{name: "last page did not include all members", pages: []*models.SnapshotList{lastPage}, tokens: []string{""}, wantCode: "GroupSnapshotMemberDetailsUnavailable", wantDetail: "snapshot-id-2"},
	}

	for _, operation := range groupSnapshotOperations {
		t.Run(operation.name, func(t *testing.T) {
			for _, tc := range testCases {
				t.Run(tc.name, func(t *testing.T) {
					logger, teardown := GetTestLogger(t)
					defer teardown()
					session, client, _, err := GetTestOpenSession(t, logger)
					require.NoError(t, err)
					client.SnapshotConsistencyGroupServiceReturns(&fakeSnapshotConsistencyGroupManager{
						createResult: group, getResult: group, getByNameResult: group,
					})
					snapshotService := &serviceFakes.SnapshotManager{}
					for i, page := range tc.pages {
						var pageErr error
						if i == len(tc.pages)-1 {
							pageErr = tc.lastErr
						}
						snapshotService.ListSnapshotsReturnsOnCall(i, page, pageErr)
					}
					client.SnapshotServiceReturns(snapshotService)

					result, err := operation.call(session)
					require.Equal(t, len(tc.tokens), snapshotService.ListSnapshotsCallCount())
					for i, token := range tc.tokens {
						limit, start, filters, _ := snapshotService.ListSnapshotsArgsForCall(i)
						assert.Equal(t, maxLimit, limit)
						assert.Equal(t, token, start)
						require.NotNil(t, filters)
						assert.Equal(t, group.ID, filters.SnapshotConsistencyGroupID)
					}
					assert.Zero(t, snapshotService.GetSnapshotCallCount())
					if tc.wantCode != "" {
						require.Nil(t, result)
						requireProviderError(t, err, tc.wantCode, group.ID, tc.wantDetail)
						if tc.wantDescription != "" {
							assert.ErrorContains(t, err, tc.wantDescription)
						}
						return
					}

					require.NoError(t, err)
					require.NotNil(t, result)
					require.Len(t, result.Snapshots, len(members))
					assert.True(t, result.ReadyToUse)
					for i, snapshot := range result.Snapshots {
						assert.Equal(t, members[i].ID, snapshot.SnapshotID)
						assert.Equal(t, members[i].SourceVolume.ID, snapshot.VolumeID)
						assert.Equal(t, *members[i].CreatedAt, snapshot.SnapshotCreationTime)
						assert.Equal(t, group.ID, snapshot.GroupSnapshotID)
					}
				})
			}
		})
	}
}

func TestGroupSnapshotRejectsIncompleteMembers(t *testing.T) {
	testCases := []struct {
		name   string
		change func(*models.SnapshotConsistencyGroupResponse, *models.SnapshotList)
	}{
		{"nil group reference", func(g *models.SnapshotConsistencyGroupResponse, _ *models.SnapshotList) { g.Snapshots[1] = nil }},
		{"group reference without ID", func(g *models.SnapshotConsistencyGroupResponse, _ *models.SnapshotList) { g.Snapshots[1].ID = "" }},
		{"duplicate group reference", func(g *models.SnapshotConsistencyGroupResponse, _ *models.SnapshotList) {
			g.Snapshots[1] = g.Snapshots[0]
		}},
		{"missing member", func(_ *models.SnapshotConsistencyGroupResponse, s *models.SnapshotList) { s.Snapshots[1] = nil }},
		{"missing source volume", func(_ *models.SnapshotConsistencyGroupResponse, s *models.SnapshotList) {
			s.Snapshots[1].SourceVolume = nil
		}},
		{"empty source volume ID", func(_ *models.SnapshotConsistencyGroupResponse, s *models.SnapshotList) {
			s.Snapshots[1].SourceVolume.ID = ""
		}},
		{"missing CRN", func(_ *models.SnapshotConsistencyGroupResponse, s *models.SnapshotList) { s.Snapshots[1].CRN = "" }},
		{"missing creation time", func(_ *models.SnapshotConsistencyGroupResponse, s *models.SnapshotList) {
			s.Snapshots[1].CreatedAt = nil
		}},
		{"CRN differs from group reference", func(_ *models.SnapshotConsistencyGroupResponse, s *models.SnapshotList) {
			s.Snapshots[1].CRN = "wrong-crn"
		}},
		{"member reports a different group", func(_ *models.SnapshotConsistencyGroupResponse, s *models.SnapshotList) {
			s.Snapshots[1].SnapshotConsistencyGroup = &models.SnapshotConsistencyGroupReference{ID: "other-group"}
		}},
	}
	for _, operation := range groupSnapshotOperations {
		t.Run(operation.name, func(t *testing.T) {
			for _, tc := range testCases {
				t.Run(tc.name, func(t *testing.T) {
					logger, teardown := GetTestLogger(t)
					defer teardown()
					session, client, _, err := GetTestOpenSession(t, logger)
					require.NoError(t, err)
					group, members := newGroupResponse(snapshotReadyState), newMemberList(snapshotReadyState)
					tc.change(group, members)
					client.SnapshotConsistencyGroupServiceReturns(&fakeSnapshotConsistencyGroupManager{
						createResult: group, getResult: group, getByNameResult: group,
					})
					snapshotService := &serviceFakes.SnapshotManager{}
					snapshotService.ListSnapshotsReturns(members, nil)
					client.SnapshotServiceReturns(snapshotService)

					result, err := operation.call(session)
					require.Nil(t, result)
					requireProviderError(t, err, "GroupSnapshotMemberDetailsUnavailable", group.ID, "")
				})
			}
		})
	}
}

// Use the real client to check that the cursor is decoded and sent back only once.
func TestGroupSnapshotMemberCursorRoundTrip(t *testing.T) {
	logger, teardown := GetTestLogger(t)
	defer teardown()
	session, client, _, err := GetTestOpenSession(t, logger)
	require.NoError(t, err)

	group := newGroupResponse(snapshotReadyState)
	members := newMemberList(snapshotReadyState).Snapshots
	queries := make(chan url.Values, 3)
	mux, httpClient, stop := test.SetupServer(t)
	defer stop()
	mux.HandleFunc(vpcvolume.Version+"/snapshots", func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query()
		queries <- query
		page := &models.SnapshotList{Snapshots: members[1:]}
		if query.Get("start") == "" {
			page.Snapshots = members[:1]
			page.Next = &models.HReference{Href: "https://vpc.example/v1/snapshots?start=a%2Bb%3D%26c&limit=100"}
		}
		assert.NoError(t, json.NewEncoder(w).Encode(page))
	})
	client.SnapshotServiceReturns(vpcvolume.NewSnapshotManager(httpClient))
	client.SnapshotConsistencyGroupServiceReturns(&fakeSnapshotConsistencyGroupManager{getResult: group})

	result, err := session.GetGroupSnapshot(group.ID)
	require.NoError(t, err)
	require.Len(t, result.Snapshots, 2)
	require.Len(t, queries, 2)
	first, second := <-queries, <-queries
	for _, query := range []url.Values{first, second} {
		assert.Equal(t, group.ID, query.Get("snapshot_consistency_group.id"))
		assert.Equal(t, "100", query.Get("limit"))
	}
	assert.False(t, first.Has("start"))
	assert.Equal(t, "a+b=&c", second.Get("start"))
}

// Malformed group responses are errors, not missing groups.
func TestGroupSnapshotInvalidResponse(t *testing.T) {
	testCases := []struct {
		name       string
		group      *models.SnapshotConsistencyGroupResponse
		wantReason string
	}{
		{name: "empty response", wantReason: "the response is empty"},
		{
			name:       "missing group ID",
			group:      &models.SnapshotConsistencyGroupResponse{LifecycleState: snapshotReadyState},
			wantReason: "the group ID is missing",
		},
	}
	for _, tc := range testCases {
		for _, operation := range groupSnapshotOperations {
			t.Run(tc.name+"/"+operation.name, func(t *testing.T) {
				logger, teardown := GetTestLogger(t)
				defer teardown()
				session, client, _, err := GetTestOpenSession(t, logger)
				require.NoError(t, err)

				groupService := &fakeSnapshotConsistencyGroupManager{
					createResult: tc.group, getResult: tc.group, getByNameResult: tc.group,
				}
				client.SnapshotConsistencyGroupServiceReturns(groupService)

				result, err := operation.call(session)

				assert.Nil(t, result)
				assert.Equal(t, 1, groupService.createCalls+groupService.getCalls+groupService.getByNameCalls)
				assert.Zero(t, client.SnapshotServiceCallCount())
				// A name lookup uses nil to mean no matching group.
				if tc.group == nil && operation.name == "get by name" {
					require.NoError(t, err)
					return
				}

				requireProviderError(t, err, "InvalidGroupSnapshotResponse", tc.wantReason, "")
				var responseError providerError.Message
				require.ErrorAs(t, err, &responseError)
				assert.Equal(t, 500, responseError.RC)
				assert.Equal(t, providerError.RetrivalFailed, responseError.Type)
				assert.NotEqual(t, providerError.EntityNotFound, responseError.Type)
			})
		}
	}
}
