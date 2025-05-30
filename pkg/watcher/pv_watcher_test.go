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

// Package watcher ...
package watcher

import (
	"bytes"
	"net/http"
	"os"
	"strings"
	"testing"

	"github.com/IBM/ibmcloud-volume-interface/config"
	cloudprovider "github.com/IBM/ibmcloud-volume-vpc/pkg/ibmcloudprovider"
	"github.com/golang/glog"
	"github.com/onsi/gomega/ghttp"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/kubernetes/scheme"
	v1core "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/tools/record"
)

func TestNew(t *testing.T) {
	// Creating test logger
	_, teardown := GetTestLogger(t)
	defer teardown()
}

func TestAddTags(t *testing.T) {
	var server *ghttp.Server
	conf := &config.Config{
		Bluemix: &config.BluemixConfig{
			IamAPIKey: "test",
		},
		VPC: &config.VPCProviderConfig{
			VPCBlockProviderName: "vpc-classic",
		},
	}
	logger, _ := GetTestLogger(t)
	fakeIBMCloudStorageProvider, _ := cloudprovider.NewFakeIBMCloudStorageProvider("configPath", logger)

	broadcaster := record.NewBroadcaster()
	broadcaster.StartLogging(glog.Infof)
	clientset := fake.NewSimpleClientset()
	eventInterface := clientset.CoreV1().Events("")
	broadcaster.StartRecordingToSink(&v1core.EventSinkImpl{Interface: eventInterface})

	pvw := &PVWatcher{
		provisionerName: "ibm-csi-driver",
		logger:          logger,
		config:          conf,
		cloudProvider:   fakeIBMCloudStorageProvider,
		recorder:        broadcaster.NewRecorder(scheme.Scheme, v1.EventSource{Component: "pod-name"}),
	}
	pv := &v1.PersistentVolume{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-pv",
		},
		Spec: v1.PersistentVolumeSpec{
			StorageClassName:              "test-storage-class",
			PersistentVolumeReclaimPolicy: v1.PersistentVolumeReclaimDelete,
			ClaimRef: &v1.ObjectReference{
				Namespace: "test-namespace",
				Name:      "test-pvc",
			},
			Capacity: v1.ResourceList(map[v1.ResourceName]resource.Quantity{
				v1.ResourceStorage: resource.MustParse("1Gi"),
			}),

			PersistentVolumeSource: v1.PersistentVolumeSource{
				CSI: &v1.CSIPersistentVolumeSource{
					Driver:       "vpc-csi-driver",
					VolumeHandle: "test-volumeid",

					VolumeAttributes: map[string]string{"tags": "mytag1:1,mytag2:2", ClusterIDLabel: "12345", "volumeCRN": "test-volcrn", "iops": "3000"},
				},
			},
		},
	}
	pvNoTags := pv.DeepCopy()
	pvNoTags.Spec.CSI.VolumeAttributes["tags"] = ""
	testCases := []struct {
		testCaseName string
		pv           *v1.PersistentVolume
		tags         string
	}{
		{
			testCaseName: "User tags- success",
			pv:           pv,
			tags:         "mytag1:1,mytag2:2",
		},
		{
			testCaseName: "No user tags- success",
			pv:           pvNoTags,
			tags:         "",
		},
	}
	for _, testcase := range testCases {
		//start test http server
		server = ghttp.NewServer()
		server.AppendHandlers(
			ghttp.CombineHandlers(
				ghttp.VerifyRequest(http.MethodGet, "/v3/tags"),
				ghttp.RespondWith(http.StatusOK, `
                           {
                            "items": {
                            }
                          }
                        `),
			),
		)
		_ = os.Setenv(IbmCloudGtAPIEndpoint, server.URL())
		t.Run(testcase.testCaseName, func(t *testing.T) {
			volCRN, tags := pvw.getTags(testcase.pv, logger)
			expectedTagNum := 7
			if len(testcase.tags) > 0 {
				expectedTagNum = 9
			}
			assert.Equal(t, expectedTagNum, len(tags))
			assert.Equal(t, "test-volcrn", volCRN)
			vol := pvw.getVolumeFromPV(pv, logger)
			assert.Equal(t, 1, *vol.Capacity)
			assert.Equal(t, "3000", *vol.Iops)
			assert.Equal(t, "test-volumeid", vol.VolumeID)
			assert.NotNil(t, vol.Attributes)
			assert.Equal(t, "12345", vol.Attributes[strings.ToLower(ClusterIDLabel)])

			pvw.updateVolume(testcase.pv, testcase.pv)
		})
	}
}

// GetTestLogger ...
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
		_ = logger.Sync()
		if t.Failed() {
			t.Log(buf)
		}
	}
	return
}
