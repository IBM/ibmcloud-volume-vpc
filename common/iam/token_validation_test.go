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

package iam

import (
	"testing"

	"github.com/dgrijalva/jwt-go"

	"github.com/IBM/ibmcloud-volume-interface/config"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func Test_GetIAMAccountIDFromAccessToken(t *testing.T) {
	logger, _ := zap.NewDevelopment(zap.AddCaller())

	fakeAccountID := "12345"
	fakeSigningKey := []byte("aabbccdd")

	fakeToken, _ := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{"account": map[string]interface{}{"bss": fakeAccountID}}).SignedString(fakeSigningKey)

	testcases := []struct {
		name              string
		token             string
		expectedAccountID string
	}{{
		name:              "fake_token",
		token:             fakeToken,
		expectedAccountID: fakeAccountID,
	}}

	for _, testcase := range testcases {
		t.Run(testcase.name, func(t *testing.T) {

			httpSetup()

			config := config.BluemixConfig{
				IamURL:          server.URL,
				IamClientID:     "test",
				IamClientSecret: "secret",
			}

			tes, err := NewTokenExchangeService(&config)
			assert.NoError(t, err)

			accountID, err := tes.GetIAMAccountIDFromAccessToken(AccessToken{Token: testcase.token}, logger)
			assert.Equal(t, testcase.expectedAccountID, accountID)
			assert.NoError(t, err)

		})
	}
}
