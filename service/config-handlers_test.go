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

	output := data.DynamicConfiguration{}
	NewResponseTester("GET", url).WithOutput(&output).WithStatus(http.StatusOK).Test(t, router)
	assert.Equal(t, dc, output)

	NewResponseTester("DELETE", url).WithStatus(http.StatusOK).Test(t, router)

	NewResponseTester("GET", url).WithStatus(http.StatusNotFound).Test(t, router)
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
			status: 404,
		},
		{
			bundle:   "bundle-non-existent",
			expected: []data.DynamicConfiguration{},
			status:   200,
		},
		{
			bundle:   "bundle-service-list",
			expected: []data.DynamicConfiguration{dcs[0], dcs[1], dcs[2], dcs[3]},
			status:   200,
		},
		{
			bundle:   "bundle-service-list",
			layer:    data.LayerRoom,
			expected: []data.DynamicConfiguration{dcs[0], dcs[1], dcs[2]},
			status:   200,
		},
		{
			bundle:   "bundle-service-list",
			layer:    data.LayerUser,
			expected: []data.DynamicConfiguration{dcs[3]},
			status:   200,
		},
		{
			bundle:   "bundle-service-list",
			layer:    data.LayerRoom,
			owner:    "foo",
			expected: []data.DynamicConfiguration{dcs[0], dcs[1]},
			status:   200,
		},
	}

	const msg = "Index=%d Layer=%q Bundle=%q Owner=%q Key=%q"
	for i, test := range tests {
		url := fmt.Sprintf("http://example.com/v2/configs/%s/%s/%s/%s", test.bundle, test.layer, test.owner, test.key)
		url = strings.TrimRight(url, "/")

		list := []data.DynamicConfiguration{}
		NewResponseTester("GET", url).WithStatus(test.status).WithOutput(&list).Test(t, router)
		assert.ElementsMatch(t, test.expected, list, msg, i, test.layer, test.bundle, test.owner, test.key)
	}
}

func TestPutAndGetDynamicConfiguration(t *testing.T) {
	router := createTestRouter()

	dc := data.DynamicConfiguration{
		Bundle: "bundle",
		Layer:  data.LayerUser,
		Owner:  "admin",
		Key:    "foo",
		Value:  "test-value",
	}

	NewResponseTester("GET", "http://example.com/v2/configs/bundle/user/admin/foo").WithStatus(http.StatusNotFound).Test(t, router)

	NewResponseTester("PUT", "http://example.com/v2/configs/bundle/user/admin/foo").WithBody(dc).WithStatus(http.StatusOK).Test(t, router)

	output := data.DynamicConfiguration{}
	NewResponseTester("GET", "http://example.com/v2/configs/bundle/user/admin/foo").WithOutput(&output).WithStatus(http.StatusOK).Test(t, router)
	assert.Equal(t, dc, output)
}
