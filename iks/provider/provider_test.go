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
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/IBM/ibmcloud-volume-interface/config"
	"github.com/IBM/ibmcloud-volume-interface/lib/provider"
	util "github.com/IBM/ibmcloud-volume-interface/lib/utils"
	"github.com/IBM/ibmcloud-volume-interface/provider/auth"
	"github.com/IBM/ibmcloud-volume-interface/provider/local"
	vpcprovider "github.com/IBM/ibmcloud-volume-vpc/block/provider"
	vpcconfig "github.com/IBM/ibmcloud-volume-vpc/block/vpcconfig"
	"github.com/IBM/ibmcloud-volume-vpc/common/vpcclient/riaas/fakes"
	volumeServiceFakes "github.com/IBM/ibmcloud-volume-vpc/common/vpcclient/vpcvolume/fakes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const (
	TestProviderAccountID   = "test-provider-account"
	TestProviderAccessToken = "test-provider-access-token"
	TestIKSAccountID        = "test-iks-account"
	TestZone                = "test-zone"
	IamURL                  = "test-iam-url"
	IamClientID             = "test-iam_client_id"
	IamClientSecret         = "test-iam_client_secret"
	IamAPIKey               = "test-iam_api_key"
	RefreshToken            = "test-refresh_token"
	TestEndpointURL         = "http://some_endpoint"
	TestAPIVersion          = "2019-07-02"
	PrivateContainerAPIURL  = "private.test-iam-url"
)

var _ local.ContextCredentialsFactory = &auth.ContextCredentialsFactory{}

func GetTestLogger(t *testing.T) (logger *zap.Logger, teardown func()) {
	atom := zap.NewAtomicLevel()
	atom.SetLevel(zap.DebugLevel)

	encoderCfg := zap.NewProductionEncoderConfig()
	encoderCfg.TimeKey = "timestamp"
	encoderCfg.EncodeTime = zapcore.ISO8601TimeEncoder

	buf := &bytes.Buffer{}

	logger = zap.New(
		zapcore.NewCore(
			zapcore.NewJSONEncoder(encoderCfg),
			zapcore.AddSync(buf),
			atom,
		),
		zap.AddCaller(),
	)

	teardown = func() {
		err := logger.Sync()
		assert.Nil(t, err)

		if t.Failed() {
			t.Log(buf)
		}
	}

	return
}

func TestNewProvider(t *testing.T) {
	var err error
	conf := &vpcconfig.VPCBlockConfig{
		ServerConfig: &config.ServerConfig{
			DebugTrace: true,
		},
		VPCConfig: &config.VPCProviderConfig{
			Enabled:         true,
			EndpointURL:     TestEndpointURL,
			VPCTimeout:      "30s",
			IamClientID:     IamClientID,
			IamClientSecret: IamClientSecret,
		},
	}
	logger, teardown := GetTestLogger(t)
	defer teardown()

	prov, err := NewProvider(conf, logger)
	assert.Nil(t, err)
	assert.NotNil(t, prov)

	conf = &vpcconfig.VPCBlockConfig{
		ServerConfig: &config.ServerConfig{
			DebugTrace: true,
		},
		VPCConfig: &config.VPCProviderConfig{
			Enabled:         true,
			EndpointURL:     TestEndpointURL,
			VPCTimeout:      "",
			IamClientID:     IamClientID,
			IamClientSecret: IamClientSecret,
		},
	}

	prov, err = NewProvider(conf, logger)
	assert.NotNil(t, prov)
	assert.Nil(t, err)

	// private endpoint related test
	conf = &vpcconfig.VPCBlockConfig{
		ServerConfig: &config.ServerConfig{
			DebugTrace: true,
		},
		VPCConfig: &config.VPCProviderConfig{
			Enabled:         true,
			EndpointURL:     TestEndpointURL,
			VPCTimeout:      "",
			IamClientID:     IamClientID,
			IamClientSecret: IamClientSecret,
		},
	}

	prov, err = NewProvider(conf, logger)
	assert.NotNil(t, prov)
	assert.Nil(t, err)

	zone := "Test Zone"
	contextCF, _ := prov.ContextCredentialsFactory(&zone)
	assert.NotNil(t, contextCF)
}

func GetTestProvider(t *testing.T, logger *zap.Logger) (local.Provider, error) {
	var cp *fakes.RegionalAPIClientProvider
	var uc, sc *fakes.RegionalAPI

	// SetRetryParameters sets the retry logic parameters
	//SetRetryParameters(2, 5)

	logger.Info("Getting New test Provider")
	conf := &vpcconfig.VPCBlockConfig{
		ServerConfig: &config.ServerConfig{
			DebugTrace: true,
		},
		VPCConfig: &config.VPCProviderConfig{
			Enabled:         true,
			EndpointURL:     TestEndpointURL,
			VPCTimeout:      "30s",
			MaxRetryAttempt: 5,
			MaxRetryGap:     10,
			APIVersion:      TestAPIVersion,
			IamClientID:     IamClientID,
			IamClientSecret: IamClientSecret,
		},
	}

	p, err := NewProvider(conf, logger)
	assert.NotNil(t, p)
	assert.Nil(t, err)

	timeout, _ := time.ParseDuration(conf.VPCConfig.VPCTimeout)

	// Inject a fake RIAAS API client
	cp = &fakes.RegionalAPIClientProvider{}
	uc = &fakes.RegionalAPI{}
	cp.NewReturnsOnCall(0, uc, nil)
	sc = &fakes.RegionalAPI{}
	cp.NewReturnsOnCall(1, sc, nil)

	volumeService := &volumeServiceFakes.VolumeService{}
	uc.VolumeServiceReturns(volumeService)

	httpClient, err := config.GeneralCAHttpClientWithTimeout(timeout)
	if err != nil {
		logger.Error("Failed to prepare HTTP client", util.ZapError(err))
		return nil, err
	}
	assert.NotNil(t, httpClient)

	assert.NotNil(t, p)

	return p, nil
}

func TestGetTestProvider(t *testing.T) {
	//var err error
	logger, teardown := GetTestLogger(t)
	defer teardown()

	prov, err := GetTestProvider(t, logger)
	assert.NotNil(t, prov)
	assert.Nil(t, err)

	zone := "Test Zone"
	contextCF, _ := prov.ContextCredentialsFactory(&zone)
	assert.NotNil(t, contextCF)
}

func TestOpenSession(t *testing.T) {
	//var err error
	logger, teardown := GetTestLogger(t)
	defer teardown()

	vpcp, err := GetTestProvider(t, logger)
	assert.Nil(t, err)
	// sessn, err := vpcp.OpenSession(context.Background(), provider.ContextCredentials{
	// 	AuthType:     provider.IAMAccessToken,
	// 	Credential:   TestProviderAccessToken,
	// 	IAMAccountID: TestIKSAccountID,
	// }, logger)

	// require.NoError(t, err)
	// assert.NotNil(t, sessn)

	sessn, err := vpcp.OpenSession(context.Background(), provider.ContextCredentials{
		AuthType:     provider.IAMAccessToken,
		IAMAccountID: TestIKSAccountID,
	}, logger)

	require.Error(t, err)
	assert.Nil(t, sessn)

	sessn, err = vpcp.OpenSession(context.Background(), provider.ContextCredentials{
		AuthType:     "WrongType",
		IAMAccountID: TestIKSAccountID,
	}, logger)

	require.Error(t, err)
	assert.Nil(t, sessn)
}

func TestUpdateAPIKey(t *testing.T) {
	logger, teardown := GetTestLogger(t)
	defer teardown()

	vpcp, err := GetTestProvider(t, logger)
	assert.Nil(t, err)

	err = vpcp.UpdateAPIKey(nil, logger)
	assert.NotNil(t, err)

	config := &vpcconfig.VPCBlockConfig{
		VPCConfig: &config.VPCProviderConfig{
			G2APIKey: "invalid",
			APIKey:   "invalid",
		},
	}

	err = vpcp.UpdateAPIKey(config, logger)
	assert.Nil(t, err)
}

func GetTestOpenSession(t *testing.T, logger *zap.Logger) (sessn *IksVpcSession, uc, sc *fakes.RegionalAPI, err error) {
	vpcp, err := GetTestProvider(t, logger)
	iksVpcProvider, _ := vpcp.(*IksVpcBlockProvider)

	m := http.NewServeMux()
	s := httptest.NewServer(m)
	assert.NotNil(t, s)

	//iksVpcProvider.VPCBlockProvider.httpClient = http.DefaultClient

	// Inject a fake RIAAS API client
	cp := &fakes.RegionalAPIClientProvider{}
	uc = &fakes.RegionalAPI{}
	cp.NewReturnsOnCall(0, uc, nil)
	sc = &fakes.RegionalAPI{}
	cp.NewReturnsOnCall(1, sc, nil)
	iksVpcProvider.VPCBlockProvider.ClientProvider = cp

	vpcSession := &vpcprovider.VPCSession{
		VPCAccountID: TestIKSAccountID,
		//Config:       iksVpcProvider.config,
		ContextCredentials: provider.ContextCredentials{
			AuthType:     provider.IAMAccessToken,
			Credential:   TestProviderAccessToken,
			IAMAccountID: TestIKSAccountID,
		},
		VolumeType: "vpc-block",
		Provider:   "vpc-classic",
		Apiclient:  uc,
		Logger:     logger,
		APIRetry:   vpcprovider.NewFlexyRetryDefault(),
	}
	sessn = &IksVpcSession{
		VPCSession: *vpcSession,
		IksSession: vpcSession,
	}

	return
}

func TestGetTestOpenSession(t *testing.T) {
	//var err error
	logger, teardown := GetTestLogger(t)
	defer teardown()

	vpcs, uc, sc, err := GetTestOpenSession(t, logger)
	assert.NotNil(t, vpcs)
	assert.NotNil(t, uc)
	assert.NotNil(t, sc)
	assert.Nil(t, err)

	providerDisplayName := vpcs.GetProviderDisplayName()
	assert.Equal(t, providerDisplayName, provider.VolumeProvider("IKS-VPC-Block"))
	vpcs.Close()

	providerName := vpcs.ProviderName()
	assert.Equal(t, providerName, provider.VolumeProvider("IKS-VPC-Block"))

	volumeType := vpcs.Type()
	assert.Equal(t, volumeType, provider.VolumeType("VPC-Block"))

	volume, _ := vpcs.GetVolume("test volume")
	assert.Nil(t, volume)
}
