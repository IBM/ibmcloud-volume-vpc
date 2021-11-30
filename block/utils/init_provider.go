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

// Package utils ...
package utils

import (
	"errors"
	"fmt"

	"github.com/IBM/ibm-csi-common/pkg/utils"
	"github.com/IBM/ibmcloud-volume-interface/config"
	"github.com/IBM/ibmcloud-volume-interface/lib/provider"
	util "github.com/IBM/ibmcloud-volume-interface/lib/utils"
	"github.com/IBM/ibmcloud-volume-interface/provider/local"
	vpc_provider "github.com/IBM/ibmcloud-volume-vpc/block/provider"
	vpcconfig "github.com/IBM/ibmcloud-volume-vpc/block/vpcconfig"
	"github.com/IBM/ibmcloud-volume-vpc/common/registry"
	iks_vpc_provider "github.com/IBM/ibmcloud-volume-vpc/iks/provider"
	"go.uber.org/zap"
	"golang.org/x/net/context"
)

// InitProviders initialization for all providers as per configurations
func InitProviders(conf *vpcconfig.VPCBlockConfig, logger *zap.Logger) (registry.Providers, error) {
	var haveProviders bool
	providerRegistry := &registry.ProviderRegistry{}

	// VPC provider registration
	if conf.VPCConfig != nil && conf.VPCConfig.Enabled {
		logger.Info("Configuring VPC Block Provider")
		prov, err := vpc_provider.NewProvider(conf, logger)
		if err != nil {
			logger.Info("VPC block provider error!")
			return nil, err
		}
		providerRegistry.Register(conf.VPCConfig.VPCBlockProviderName, prov)
		haveProviders = true
	}

	// IKS provider registration
	if conf.IKSConfig != nil && conf.IKSConfig.Enabled {
		logger.Info("Configuring IKS-VPC Block Provider")
		prov, err := iks_vpc_provider.NewProvider(conf, logger)
		if err != nil {
			logger.Info("VPC block provider error!")
			return nil, err
		}
		providerRegistry.Register(conf.IKSConfig.IKSBlockProviderName, prov)
		haveProviders = true
	}

	if haveProviders {
		logger.Info("Provider registration done!!!")
		return providerRegistry, nil
	}

	return nil, errors.New("no providers registered")
}

// OpenProviderSession ...
func OpenProviderSession(providerConfig *config.Config, providers registry.Providers, providerID string, ctxLogger *zap.Logger) (session provider.Session, fatal bool, err error) {
	return OpenProviderSessionWithContext(context.TODO(), providerConfig, providers, providerID, ctxLogger)
}

// OpenProviderSessionWithContext ...
func OpenProviderSessionWithContext(ctx context.Context, providerConfig *config.Config, providers registry.Providers, providerID string, ctxLogger *zap.Logger) (provider.Session, bool, error) {
	prov, err := providers.Get(providerID)
	if err != nil {
		ctxLogger.Error("Not able to get the said provider, might be its not registered", local.ZapError(err))
		return nil, true, err
	}

	vpcBlockConfig := &vpcconfig.VPCBlockConfig{
		VPCConfig:    providerConfig.VPC,
		IKSConfig:    providerConfig.IKS,
		APIConfig:    providerConfig.API,
		ServerConfig: providerConfig.Server,
	}

	maxRetryAttempt := 3
	for retryCount := 0; retryCount < maxRetryAttempt; retryCount++ {
		ccf, err := prov.ContextCredentialsFactory(nil)
		if err != nil {
			ctxLogger.Error("Unable to fetch credentials", local.ZapError(err))
			return nil, true, err
		}
		ctxLogger.Info("Calling provider/utils/init_provider.go GenerateContextCredentials")
		contextCredentials, err := GenerateContextCredentials(vpcBlockConfig, providerID, ccf, ctxLogger)
		if err != nil {
			ctxLogger.Error("Unable to generate credentials", local.ZapError(err))
			return nil, true, err
		}
		ctxLogger.Info(fmt.Sprintf("Attempt %d - open provider session", retryCount+1))
		session, err := prov.OpenSession(ctx, contextCredentials, ctxLogger)
		if err == nil {
			return session, false, nil
		}
		if providerError, ok := err.(provider.Error); ok && string(providerError.Code()) == provider.APIKeyNotFound {
			apiKeyImp, err := utils.NewAPIKeyImpl(ctxLogger)
			if err != nil {
				ctxLogger.Fatal("Unable to create API key getter", zap.Reflect("Error", err))
				return nil, true, err
			}
			ctxLogger.Info("Created NewAPIKeyImpl...")
			err = apiKeyImp.UpdateIAMKeys(providerConfig)
			if err != nil {
				ctxLogger.Fatal("Unable to get API key", local.ZapError(err))
				return nil, true, err
			}
			vpcBlockConfig.VPCConfig.APIKey = providerConfig.VPC.G2APIKey
			vpcBlockConfig.VPCConfig.G2APIKey = providerConfig.VPC.G2APIKey
			err = prov.UpdateAPIKey(vpcBlockConfig, ctxLogger)
			if err != nil {
				ctxLogger.Error("Failed to update API key in the provider", local.ZapError(err))
				return nil, true, err
			}
			continue
		}
		return nil, true, err
	}

	ctxLogger.Error("Failed to open provider session", local.ZapError(errors.New("API key not found")))
	return nil, true, errors.New("API key not found, retry after few minutes")

}

// GenerateContextCredentials ...
func GenerateContextCredentials(conf *vpcconfig.VPCBlockConfig, providerID string, contextCredentialsFactory local.ContextCredentialsFactory, ctxLogger *zap.Logger) (provider.ContextCredentials, error) {
	ctxLogger.Info("Generating generateContextCredentials for ", zap.String("Provider ID", providerID))

	// Select appropriate authentication strategy
	switch {
	case (conf.VPCConfig != nil && providerID == conf.VPCConfig.VPCBlockProviderName):
		ctxLogger.Info("Calling provider/init_provider.go ForIAMAccessToken")
		return contextCredentialsFactory.ForIAMAccessToken(conf.VPCConfig.APIKey, ctxLogger)

	case (conf.IKSConfig != nil && providerID == conf.IKSConfig.IKSBlockProviderName):
		return provider.ContextCredentials{}, nil // Get credentials  in OpenSession method

	default:
		return provider.ContextCredentials{}, util.NewError("ErrorInsufficientAuthentication",
			"Insufficient authentication credentials")
	}
}
