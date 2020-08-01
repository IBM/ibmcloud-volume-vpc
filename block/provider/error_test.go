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
	"fmt"
	"testing"

	"github.com/IBM/ibmcloud-volume-interface/lib/utils/reasoncode"

	"github.com/stretchr/testify/assert"
)

func Test_Errors(t *testing.T) {
	testCases := []struct {
		testName        string
		errorCode       reasoncode.ReasonCode
		errorMessage    string
		wrappedMessages []string
		properties      map[string]string
	}{
		{
			// ErrorUnclassified - General unclassified error
			testName:  "ErrorUnclassified",
			errorCode: reasoncode.ErrorUnclassified,
		},
		{
			// Default error code
			testName:  "DefaultCode",
			errorCode: "",
		},
		{
			// Example of a specific errorCode
			testName:  "ErrorUnknownProvider",
			errorCode: reasoncode.ErrorUnknownProvider,
		},
		{
			testName:        "Wrapped",
			errorCode:       reasoncode.ErrorUnclassified,
			wrappedMessages: []string{"This is a wrapped exception"},
		},
		{
			testName:        "MultiWrapped",
			errorCode:       reasoncode.ErrorUnclassified,
			wrappedMessages: []string{"This is a wrapped exception", "This is another wrapped exception"},
		},
		{
			testName:   "Properties",
			errorCode:  reasoncode.ErrorUnclassified,
			properties: map[string]string{"prop1": "val1", "prop2": "val2"},
		},
	}

	for _, testCase := range testCases {
		t.Run(fmt.Sprintf("Case %v", testCase.testName), func(t *testing.T) {
			err := Error{
				Fault: Fault{
					Message:    testCase.errorMessage,
					ReasonCode: testCase.errorCode,
					Wrapped:    testCase.wrappedMessages,
					Properties: testCase.properties,
				},
			}
			assert.Equal(t, testCase.errorMessage, err.Error())
			if testCase.errorCode == "" {
				assert.Equal(t, reasoncode.ErrorUnclassified, err.Code())
			} else {
				assert.Equal(t, testCase.errorCode, err.Code())
			}
			assert.Equal(t, testCase.wrappedMessages, err.Wrapped())
			assert.Equal(t, testCase.properties, err.Properties())
		})
	}
}
