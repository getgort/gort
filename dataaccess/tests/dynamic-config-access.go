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

package tests

import (
	"testing"

	"github.com/getgort/gort/data"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func (da DataAccessTester) testDynamicConfigurationAccess(t *testing.T) {
	t.Run("testDynamicConfigurationCreate", da.testDynamicConfigurationCreate)
	t.Run("testDynamicConfigurationDelete", da.testDynamicConfigurationDelete)
	t.Run("testDynamicConfigurationExists", da.testDynamicConfigurationExists)
	t.Run("testDynamicConfigurationGet", da.testDynamicConfigurationGet)
	t.Run("testDynamicConfigurationList", da.testDynamicConfigurationList)
}

func (da DataAccessTester) testDynamicConfigurationCreate(t *testing.T) {
	tests := []struct {
		dc  data.DynamicConfiguration
		err bool
	}{
		{
			dc:  data.DynamicConfiguration{},
			err: true,
		}, {
			dc: data.DynamicConfiguration{
				Bundle: "some-bundle",
			},
			err: true,
		}, {
			dc: data.DynamicConfiguration{
				Bundle: "some-bundle",
				Layer:  data.LayerBundle,
			},
			err: true,
		}, {
			dc: data.DynamicConfiguration{
				Bundle: "some-bundle",
				Layer:  data.LayerBundle,
				Owner:  "some-bundle",
			},
			err: true,
		}, {
			dc: data.DynamicConfiguration{
				Bundle: "some-bundle",
				Layer:  data.LayerBundle,
				Owner:  "some-bundle",
				Key:    "username",
			},
			err: false,
		}, {
			dc: data.DynamicConfiguration{
				Bundle: "some-bundle",
				Layer:  data.ConfigurationLayer("nonstandard-bundle"),
				Owner:  "some-bundle",
				Key:    "username",
			},
			err: true,
		},
	}

	for i, test := range tests {
		err := da.DynamicConfigurationCreate(da.ctx, test.dc)
		defer da.DynamicConfigurationDelete(da.ctx, test.dc.Layer, test.dc.Bundle, test.dc.Owner, test.dc.Key)

		if test.err {
			assert.Error(t, err, "test %d: %v", i, test.dc)
		} else {
			assert.NoError(t, err, "test %d: %v", i, test.dc)
		}
	}
}

func (da DataAccessTester) testDynamicConfigurationDelete(t *testing.T) {
	pos := data.DynamicConfiguration{
		Bundle: "positive-bundle-delete",
		Layer:  data.LayerBundle,
		Owner:  "some-bundle",
		Key:    "username",
	}

	err := da.DynamicConfigurationCreate(da.ctx, pos)
	require.NoError(t, err)
	defer da.DynamicConfigurationDelete(da.ctx, pos.Layer, pos.Bundle, pos.Owner, pos.Key)

	tests := []struct {
		dc  data.DynamicConfiguration
		err bool
	}{
		{
			dc:  data.DynamicConfiguration{},
			err: true,
		}, {
			dc: data.DynamicConfiguration{
				Bundle: "some-bundle",
			},
			err: true,
		}, {
			dc: data.DynamicConfiguration{
				Bundle: "some-bundle",
				Layer:  data.LayerBundle,
			},
			err: true,
		}, {
			dc: data.DynamicConfiguration{
				Bundle: "some-bundle",
				Layer:  data.LayerBundle,
				Owner:  "some-bundle",
			},
			err: true,
		}, {
			dc: data.DynamicConfiguration{
				Bundle: "some-bundle",
				Layer:  data.LayerBundle,
				Owner:  "some-bundle",
				Key:    "username",
			},
			err: true,
		}, {
			dc: data.DynamicConfiguration{
				Bundle: "some-bundle",
				Layer:  data.ConfigurationLayer("nonstandard-bundle"),
				Owner:  "some-bundle",
				Key:    "username",
			},
			err: true,
		}, {
			dc:  pos,
			err: false,
		}, {
			dc:  pos,
			err: true,
		},
	}

	for i, test := range tests {
		err := da.DynamicConfigurationDelete(da.ctx, test.dc.Layer, test.dc.Bundle, test.dc.Owner, test.dc.Key)

		if test.err {
			assert.Error(t, err, "test %d: %v", i, test.dc)
		} else {
			assert.NoError(t, err, "test %d: %v", i, test.dc)
		}
	}
}

func (da DataAccessTester) testDynamicConfigurationExists(t *testing.T) {
	pos := data.DynamicConfiguration{
		Bundle: "positive-bundle-exists",
		Layer:  data.LayerBundle,
		Owner:  "some-bundle",
		Key:    "username",
	}

	err := da.DynamicConfigurationCreate(da.ctx, pos)
	require.NoError(t, err)
	defer da.DynamicConfigurationDelete(da.ctx, pos.Layer, pos.Bundle, pos.Owner, pos.Key)

	tests := []struct {
		dc       data.DynamicConfiguration
		expected bool
		err      bool
	}{
		{
			dc:  data.DynamicConfiguration{},
			err: true,
		}, {
			dc: data.DynamicConfiguration{
				Bundle: "no-such-bundle",
			},
			err: true,
		}, {
			dc: data.DynamicConfiguration{
				Bundle: "no-such-bundle",
				Layer:  data.LayerBundle,
			},
			err: true,
		}, {
			dc: data.DynamicConfiguration{
				Bundle: "no-such-bundle",
				Layer:  data.LayerBundle,
				Owner:  "no-such-bundle",
			},
			err: true,
		}, {
			dc: data.DynamicConfiguration{
				Bundle: "no-such-bundle",
				Layer:  data.LayerBundle,
				Owner:  "no-such-bundle",
				Key:    "username",
			},
			err:      false,
			expected: false,
		}, {
			dc: data.DynamicConfiguration{
				Bundle: "no-such-bundle",
				Layer:  data.ConfigurationLayer("nonstandard-bundle"),
				Owner:  "no-such-bundle",
				Key:    "username",
			},
			err: true,
		}, {
			dc:       pos,
			err:      false,
			expected: true,
		},
	}

	for i, test := range tests {
		exists, err := da.DynamicConfigurationExists(da.ctx, test.dc.Layer, test.dc.Bundle, test.dc.Owner, test.dc.Key)

		if test.err {
			assert.Error(t, err, "test %d: %v", i, test.dc)
		} else {
			assert.NoError(t, err, "test %d: %v", i, test.dc)
			assert.Equal(t, test.expected, exists, "test %d: %v", i, test.dc)
		}
	}
}

func (da DataAccessTester) testDynamicConfigurationGet(t *testing.T) {
	dcs := []data.DynamicConfiguration{
		{Bundle: "bundle-create", Layer: data.LayerRoom, Owner: "foo", Key: "username"},
		{Bundle: "bundle-create", Layer: data.LayerRoom, Owner: "foo", Key: "password", Secret: true},
	}

	for _, dc := range dcs {
		err := da.DynamicConfigurationCreate(da.ctx, dc)
		require.NoError(t, err)
	}
	defer func() {
		for _, dc := range dcs {
			da.DynamicConfigurationDelete(da.ctx, dc.Layer, dc.Bundle, dc.Owner, dc.Key)
		}
	}()

	tests := []struct {
		bundle   string
		layer    data.ConfigurationLayer
		owner    string
		key      string
		expected data.DynamicConfiguration
		err      bool
	}{
		{
			bundle: "",
			err:    true,
		},
		{
			bundle: "bundle-create",
			err:    true,
		},
		{
			bundle: "bundle-create",
			layer:  data.LayerRoom,
			err:    true,
		},
		{
			bundle: "bundle-create",
			layer:  data.LayerRoom,
			owner:  "foo",
			err:    true,
		},
		{
			bundle:   "bundle-create",
			layer:    data.LayerRoom,
			owner:    "foo",
			key:      "username",
			expected: dcs[0],
		},
		{
			bundle:   "bundle-create",
			layer:    data.LayerRoom,
			owner:    "foo",
			key:      "password",
			expected: dcs[1],
		},
		{
			bundle: "bundle-create",
			layer:  data.LayerRoom,
			owner:  "foo",
			key:    "doesnt-exist",
			err:    true,
		},
	}

	const msg = "Index=%d Layer=%q Bundle=%q Owner=%q Key=%q"
	for i, test := range tests {
		dc, err := da.DynamicConfigurationGet(da.ctx, test.layer, test.bundle, test.owner, test.key)

		if test.err {
			assert.Error(t, err, msg, i, test.layer, test.bundle, test.owner, test.key)
		} else {
			assert.NoError(t, err, msg, i, test.layer, test.bundle, test.owner, test.key)
			assert.Equal(t, test.expected, dc, msg, i, test.layer, test.bundle, test.owner, test.key)
		}
	}
}

func (da DataAccessTester) testDynamicConfigurationList(t *testing.T) {
	dcs := []data.DynamicConfiguration{
		{Bundle: "bundle-list", Layer: data.LayerRoom, Owner: "foo", Key: "username"},
		{Bundle: "bundle-list", Layer: data.LayerRoom, Owner: "foo", Key: "password"},
		{Bundle: "bundle-list", Layer: data.LayerRoom, Owner: "bar", Key: "username"},
		{Bundle: "bundle-list", Layer: data.LayerUser, Owner: "bar", Key: "username"},
	}

	for _, dc := range dcs {
		err := da.DynamicConfigurationCreate(da.ctx, dc)
		require.NoError(t, err)
	}
	defer func() {
		for _, dc := range dcs {
			da.DynamicConfigurationDelete(da.ctx, dc.Layer, dc.Bundle, dc.Owner, dc.Key)
		}
	}()

	tests := []struct {
		bundle   string
		layer    data.ConfigurationLayer
		owner    string
		key      string
		expected []data.DynamicConfiguration
		err      bool
	}{
		{
			bundle: "",
			err:    true,
		},
		{
			bundle: "",
			layer:  data.LayerRoom,
			owner:  "foo",
			key:    "username",
			err:    true,
		},
		{
			bundle:   "bundle-non-existent",
			expected: []data.DynamicConfiguration{},
			err:      false,
		},
		{
			bundle:   "bundle-list",
			expected: []data.DynamicConfiguration{dcs[0], dcs[1], dcs[2], dcs[3]},
			err:      false,
		},
		{
			bundle:   "bundle-list",
			layer:    data.LayerRoom,
			expected: []data.DynamicConfiguration{dcs[0], dcs[1], dcs[2]},
			err:      false,
		},
		{
			bundle:   "bundle-list",
			layer:    data.LayerUser,
			expected: []data.DynamicConfiguration{dcs[3]},
			err:      false,
		},
		{
			bundle:   "bundle-list",
			layer:    data.LayerRoom,
			owner:    "foo",
			expected: []data.DynamicConfiguration{dcs[0], dcs[1]},
			err:      false,
		},
		{
			bundle:   "bundle-list",
			owner:    "foo",
			expected: []data.DynamicConfiguration{dcs[0], dcs[1]},
			err:      false,
		},
		{
			bundle:   "bundle-list",
			owner:    "bar",
			expected: []data.DynamicConfiguration{dcs[2], dcs[3]},
			err:      false,
		},
		{
			bundle:   "bundle-list",
			layer:    data.LayerRoom,
			owner:    "foo",
			key:      "username",
			expected: []data.DynamicConfiguration{dcs[0]},
			err:      false,
		},
	}

	const msg = "Index=%d Layer=%q Bundle=%q Owner=%q Key=%q"
	for i, test := range tests {
		list, err := da.DynamicConfigurationList(da.ctx, test.layer, test.bundle, test.owner, test.key)

		if test.err {
			assert.Error(t, err, msg, i, test.layer, test.bundle, test.owner, test.key)
		} else {
			assert.NoError(t, err, msg, i, test.layer, test.bundle, test.owner, test.key)
			assert.ElementsMatch(t, test.expected, list, msg, i, test.layer, test.bundle, test.owner, test.key)
		}
	}
}
