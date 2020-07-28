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

package auth

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/IBM/ibmcloud-volume-interface/config"

	"github.com/IBM/ibmcloud-volume-interface/lib/provider"
	"go.uber.org/zap"
)

var (
	logger *zap.Logger
)

func init() {
	logger, _ = zap.NewDevelopment()
}

func TestForIaaSAPIKey(t *testing.T) {
	account := "account1"
	username := "user1"
	apiKey := "abcdefg"
	endpointURL := "http://myEndpointUrl"

	ccf := &ContextCredentialsFactory{
		softlayerConfig: &config.SoftlayerConfig{
			SoftlayerEndpointURL: endpointURL,
		},
	}

	contextCredentials, err := ccf.ForIaaSAPIKey(account, username, apiKey, logger)

	assert.NoError(t, err)

	assert.Equal(t, provider.ContextCredentials{
		AuthType:     provider.IaaSAPIKey,
		IAMAccountID: account,
		UserID:       username,
		Credential:   apiKey,
	}, contextCredentials)

}
