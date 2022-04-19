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
	"github.com/stretchr/testify/assert"

	"testing"

	"github.com/IBM/ibmcloud-volume-interface/lib/provider"
	"github.com/IBM/ibmcloud-volume-interface/provider/auth"
)

func TestTokenGenerator(t *testing.T) {
	logger, teardown := GetTestLogger(t)
	defer teardown()

	tg := tokenGenerator{}
	assert.NotNil(t, tg)

	cf := provider.ContextCredentials{
		AuthType:     provider.IAMAccessToken,
		Credential:   TestProviderAccessToken,
		IAMAccountID: TestIKSAccountID,
	}
	signedToken, err := tg.getServiceToken(cf, *logger)
	assert.Nil(t, signedToken)
	assert.NotNil(t, err)

	tg.tokenKID = "sample_key"
	signedToken, err = tg.getServiceToken(cf, *logger)
	assert.Nil(t, signedToken)
	assert.NotNil(t, err)

	cf = provider.ContextCredentials{
		AuthType:     provider.IAMAccessToken,
		Credential:   TestProviderAccessToken,
		IAMAccountID: TestIKSAccountID,
		UserID:       TestIKSAccountID,
	}

	tg.tokenKID = "no_sample_key"
	signedToken, err = tg.getServiceToken(cf, *logger)
	assert.Nil(t, signedToken)
	assert.NotNil(t, err)

	signedToken, err = tg.getServiceToken(cf, *logger)
	assert.Nil(t, signedToken)
	assert.NotNil(t, err)

	tg.tokenKID = "no_sample_key"
	cf = provider.ContextCredentials{
		AuthType:     auth.IMSToken,
		Credential:   TestProviderAccessToken,
		IAMAccountID: TestIKSAccountID,
		UserID:       TestIKSAccountID,
	}
	signedToken, err = tg.getServiceToken(cf, *logger)
	assert.Nil(t, signedToken)
	assert.NotNil(t, err)

	tg.tokenKID = "sample_key_invalid"
	cf = provider.ContextCredentials{
		AuthType:     auth.IMSToken,
		Credential:   TestProviderAccessToken,
		IAMAccountID: TestIKSAccountID,
		UserID:       TestIKSAccountID,
	}
	signedToken, err = tg.getServiceToken(cf, *logger)
	assert.Nil(t, signedToken)
	assert.NotNil(t, err)
}
