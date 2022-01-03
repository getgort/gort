/*
 * Copyright 2021 The Gort Authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *      http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package service

import (
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/getgort/gort/data"
	"github.com/stretchr/testify/assert"
)

func TestDeleteDynamicConfig(t *testing.T) {
	const url = "http://example.com/v2/configs/bundle-service-delete/user/admin/foo"

	dc := data.DynamicConfiguration{
		Bundle: "bundle-service-delete",
		Layer:  data.LayerUser,
		Owner:  "admin",
		Key:    "foo",
		Value:  "test-value",
	}

	router := createTestRouter()

	NewResponseTester("PUT", url).WithBody(dc).WithStatus(http.StatusOK).Test(t, router)

	output := []data.DynamicConfiguration{}
	NewResponseTester("GET", url).WithOutput(&output).WithStatus(http.StatusOK).Test(t, router)
	assert.Equal(t, []data.DynamicConfiguration{dc}, output)

	NewResponseTester("DELETE", url).WithStatus(http.StatusOK).Test(t, router)

	NewResponseTester("GET", url).WithStatus(http.StatusNoContent).Test(t, router)
}

func TestGetDynamicConfigs(t *testing.T) {
	router := createTestRouter()

	dcs := []data.DynamicConfiguration{
		{Bundle: "bundle-service-list", Layer: data.LayerRoom, Owner: "foo", Key: "username", Value: "one"},
		{Bundle: "bundle-service-list", Layer: data.LayerRoom, Owner: "foo", Key: "password", Value: "two"},
		{Bundle: "bundle-service-list", Layer: data.LayerRoom, Owner: "bar", Key: "username", Value: "three"},
		{Bundle: "bundle-service-list", Layer: data.LayerUser, Owner: "bar", Key: "username", Value: "four"},
	}

	for _, dc := range dcs {
		url := fmt.Sprintf("http://example.com/v2/configs/%s/%s/%s/%s", dc.Bundle, dc.Layer, dc.Owner, dc.Key)
		NewResponseTester("PUT", url).WithBody(dc).WithStatus(http.StatusOK).Test(t, router)
	}

	tests := []struct {
		bundle   string
		layer    data.ConfigurationLayer
		owner    string
		key      string
		expected []data.DynamicConfiguration
		status   int
	}{
		{
			bundle: "",
			status: http.StatusNotFound,
		},
		{
			bundle: "*",
			layer:  data.LayerRoom,
			owner:  "foo",
			key:    "username",
			status: http.StatusExpectationFailed,
		},
		{
			bundle: "*",
			layer:  data.LayerRoom,
			owner:  "foo",
			status: http.StatusExpectationFailed,
		},
		{
			bundle:   "bundle-non-existent",
			expected: []data.DynamicConfiguration{},
			status:   http.StatusNoContent,
		},
		{
			bundle:   "bundle-service-list",
			expected: []data.DynamicConfiguration{dcs[0], dcs[1], dcs[2], dcs[3]},
			status:   http.StatusOK,
		},
		{
			bundle:   "bundle-service-list",
			layer:    data.LayerRoom,
			expected: []data.DynamicConfiguration{dcs[0], dcs[1], dcs[2]},
			status:   http.StatusOK,
		},
		{
			bundle:   "bundle-service-list",
			layer:    data.LayerUser,
			expected: []data.DynamicConfiguration{dcs[3]},
			status:   http.StatusOK,
		},
		{
			bundle:   "bundle-service-list",
			layer:    data.LayerRoom,
			owner:    "foo",
			expected: []data.DynamicConfiguration{dcs[0], dcs[1]},
			status:   http.StatusOK,
		},
		{
			bundle:   "bundle-service-list",
			layer:    data.ConfigurationLayer("*"),
			owner:    "foo",
			expected: []data.DynamicConfiguration{dcs[0], dcs[1]},
			status:   http.StatusOK,
		},
		{
			bundle:   "bundle-service-list",
			layer:    data.ConfigurationLayer("*"),
			owner:    "bar",
			expected: []data.DynamicConfiguration{dcs[2], dcs[3]},
			status:   http.StatusOK,
		},
		{
			bundle:   "bundle-service-list",
			layer:    data.LayerRoom,
			owner:    "foo",
			key:      "username",
			expected: []data.DynamicConfiguration{dcs[0]},
			status:   http.StatusOK,
		},
	}

	const msg = "Index=%d Layer=%q Bundle=%q Owner=%q Key=%q"
	for i, test := range tests {
		url := fmt.Sprintf("http://example.com/v2/configs/%s/%s/%s/%s", test.bundle, test.layer, test.owner, test.key)
		url = strings.TrimRight(url, "/")

		list := []data.DynamicConfiguration{}
		NewResponseTester("GET", url).WithStatus(test.status).WithOutput(&list).Test(t, router, msg, i, test.layer, test.bundle, test.owner, test.key)
		assert.ElementsMatch(t, test.expected, list, msg, i, test.layer, test.bundle, test.owner, test.key)
	}
}
