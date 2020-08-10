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

// Package auth ...
package auth

import (
	"github.com/IBM/ibmcloud-volume-interface/config"
	"github.com/IBM/ibmcloud-volume-interface/provider/auth"
	vpciam "github.com/IBM/ibmcloud-volume-vpc/common/iam"
)

// NewVPCContextCredentialsFactory ...
func NewVPCContextCredentialsFactory(bluemixConfig *config.BluemixConfig, vpcConfig *config.VPCProviderConfig) (*auth.ContextCredentialsFactory, error) {

	ccf, err := auth.NewContextCredentialsFactory(bluemixConfig, nil, vpcConfig)
	if bluemixConfig.PrivateAPIRoute != "" {
		ccf.TokenExchangeService, err = vpciam.NewTokenExchangeIKSService(bluemixConfig)
	}
	if err != nil {
		return nil, err
	}
	return ccf, nil

}
