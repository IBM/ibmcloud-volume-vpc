/**
 * Copyright 2025 IBM Corp.
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

// Package riaas ...
package riaas

import (
	"net/url"
	"testing"
	"time"

	"github.com/IBM/ibmcloud-volume-vpc/common/vpcclient/client"
	"github.com/IBM/ibmcloud-volume-vpc/common/vpcclient/client/fakes"
	"github.com/IBM/ibmcloud-volume-vpc/common/vpcclient/models"
	"github.com/stretchr/testify/assert"
)

// requestQuery returns the query values of a request created by c
func requestQuery(t *testing.T, c client.SessionClient) url.Values {
	u, err := url.Parse(c.NewRequest(&client.Operation{Name: "op", Method: "GET", PathPattern: "v1/test"}).URL())
	assert.NoError(t, err)
	return u.Query()
}

func TestNewClient(t *testing.T) {
	base := &fakes.SessionClient{}

	// BMS gets the BMS client, every other profile gets the VPC client
	assert.IsType(t, &bmsClient{}, newClient(base, profileBMS, Config{}))
	assert.IsType(t, &vpcClient{}, newClient(base, profileVPC, Config{}))
	assert.IsType(t, &vpcClient{}, newClient(base, profileIKS, Config{}))
}

func TestNewClientQueryValues(t *testing.T) {
	cfg := Config{BaseURL: "http://gc", APIVersion: "2019-06-05", APIGeneration: 2}
	base := newBaseClient(cfg)

	// the base client carries only the profile-neutral values
	baseQuery := requestQuery(t, base)
	assert.Empty(t, baseQuery.Get("version"))
	assert.Equal(t, "2", baseQuery.Get("generation"))

	// VPC and IKS use the configured API version
	for _, profile := range []apiProfile{profileVPC, profileIKS} {
		query := requestQuery(t, newClient(base, profile, cfg))
		assert.Equal(t, "2019-06-05", query.Get("version"), string(profile))
		assert.Equal(t, "2", query.Get("generation"), string(profile))
		assert.Empty(t, query.Get("maturity"), string(profile))
	}

	// BMS uses the current date and maturity=development, and never the configured version
	query := requestQuery(t, newClient(base, profileBMS, cfg))
	assert.Equal(t, time.Now().UTC().Format(time.DateOnly), query.Get("version"))
	assert.Equal(t, "development", query.Get("maturity"))
	assert.Equal(t, "2", query.Get("generation"))
}

func TestNewClientDefaultAPIVersion(t *testing.T) {
	cfg := Config{BaseURL: "http://gc"}
	query := requestQuery(t, newClient(newBaseClient(cfg), profileVPC, cfg))

	assert.Equal(t, models.APIVersion, query.Get("version"))
	assert.Equal(t, "1", query.Get("generation"))
}

func TestNewBaseClient(t *testing.T) {
	assert.NotNil(t, newBaseClient(Config{BaseURL: "http://gc"}))
}

func TestIKSVolumeAttachService(t *testing.T) {
	session := &Session{client: &fakes.SessionClient{}}
	assert.NotNil(t, session.IKSVolumeAttachService())
}

func TestBMSVolumeAttachService(t *testing.T) {
	session := &Session{client: &fakes.SessionClient{}}
	assert.NotNil(t, session.BMSVolumeAttachService())
}

// IKS ListVolumeAttachments does vs.client = vs.client.WithQueryValue(...), the
// profile must survive that chained call
func TestChainedWithQueryValueKeepsProfile(t *testing.T) {
	cfg := Config{BaseURL: "http://gc", APIVersion: "2019-06-05", APIGeneration: 2}

	for _, profile := range []apiProfile{profileVPC, profileIKS, profileBMS} {
		c := newClient(newBaseClient(cfg), profile, cfg).WithQueryValue("cluster", "c1")

		query := requestQuery(t, c)
		assert.Equal(t, "c1", query.Get("cluster"), string(profile))
		assert.NotEmpty(t, query.Get("version"), string(profile))
		assert.Equal(t, "2", query.Get("generation"), string(profile))
	}
}
