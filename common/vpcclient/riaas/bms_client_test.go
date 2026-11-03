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
	"github.com/stretchr/testify/assert"
)

func TestBMSClientNewRequest(t *testing.T) {
	cfg := Config{BaseURL: "http://gc", APIVersion: "2019-06-05", APIGeneration: 2}
	base := newBaseClient(cfg)
	operation := &client.Operation{Name: "op", Method: "GET", PathPattern: "v1/test"}

	// BMS sets version and maturity, and keeps the generation of the base client
	bmsURL, err := url.Parse(newClient(base, profileBMS, cfg).NewRequest(operation).URL())
	assert.NoError(t, err)
	assert.Equal(t, time.Now().UTC().Format(time.DateOnly), bmsURL.Query().Get("version"))
	assert.Equal(t, "development", bmsURL.Query().Get("maturity"))
	assert.Equal(t, "2", bmsURL.Query().Get("generation"))
}

func TestBMSClientSharesAuthToken(t *testing.T) {
	base := &fakes.SessionClient{}
	bms := newClient(base, profileBMS, Config{})

	bms.WithAuthToken("token")
	if assert.Equal(t, 1, base.WithAuthTokenCallCount()) {
		assert.Equal(t, "token", base.WithAuthTokenArgsForCall(0))
	}
}
