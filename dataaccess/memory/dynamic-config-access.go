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

package memory

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/getgort/gort/data"
	"github.com/getgort/gort/dataaccess/errs"
)

func (da *InMemoryDataAccess) DynamicConfigurationCreate(_ context.Context, config data.DynamicConfiguration) error {
	lookupKey, err := generateLookupKey(config.Layer, config.Bundle, config.Owner, config.Key)
	if err != nil {
		return err
	}

	if err := config.Layer.Validate(); err != nil {
		return err
	}

	if da.configs[lookupKey] != nil {
		return errs.ErrConfigExists
	}

	da.configs[lookupKey] = &config

	return nil
}

func (da *InMemoryDataAccess) DynamicConfigurationDelete(_ context.Context, layer data.ConfigurationLayer, bundle, owner, key string) error {
	lookupKey, err := generateLookupKey(layer, bundle, owner, key)
	if err != nil {
		return err
	}

	if err := layer.Validate(); err != nil {
		return err
	}

	if da.configs[lookupKey] == nil {
		return errs.ErrNoSuchConfig
	}

	delete(da.configs, lookupKey)

	return nil
}

func (da *InMemoryDataAccess) DynamicConfigurationExists(_ context.Context, layer data.ConfigurationLayer, bundle, owner, key string) (bool, error) {
	lookupKey, err := generateLookupKey(layer, bundle, owner, key)
	if err != nil {
		return false, err
	}

	if err := layer.Validate(); err != nil {
		return false, err
	}

	return da.configs[lookupKey] != nil, nil
}

func (da *InMemoryDataAccess) DynamicConfigurationGet(_ context.Context, layer data.ConfigurationLayer, bundle, owner, key string) (data.DynamicConfiguration, error) {
	lookupKey, err := generateLookupKey(layer, bundle, owner, key)
	if err != nil {
		return data.DynamicConfiguration{}, err
	}

	if err := layer.Validate(); err != nil {
		return data.DynamicConfiguration{}, err
	}

	dc := da.configs[lookupKey]
	if dc == nil {
		return data.DynamicConfiguration{}, errs.ErrNoSuchConfig
	}

	return *dc, nil
}

// DynamicConfigurationList will list matching configurations. Empty values
// are treated as wildcards. Bundle (at a minimum) must be not empty.
func (da *InMemoryDataAccess) DynamicConfigurationList(_ context.Context, layer data.ConfigurationLayer, bundle, owner, key string) ([]data.DynamicConfiguration, error) {
	const wildcard = `([^\|]*)`

	if bundle == "" {
		return nil, errs.ErrEmptyConfigBundle
	}
	if layer == "" {
		layer = wildcard
	}
	if owner == "" {
		owner = wildcard
	}
	if key == "" {
		key = wildcard
	}

	str, _ := generateLookupKey(layer, bundle, owner, key)
	p := regexp.MustCompile(fmt.Sprintf(`^%s$`, strings.ReplaceAll(str, "|", `\|`)))

	var cc []data.DynamicConfiguration

	for k, v := range da.configs {
		if p.Match([]byte(k)) {
			cc = append(cc, *v)
		}
	}

	return cc, nil
}

func generateLookupKey(layer data.ConfigurationLayer, bundle, owner, key string) (string, error) {
	switch {
	case bundle == "":
		return "", errs.ErrEmptyConfigBundle
	case layer == "":
		return "", errs.ErrEmptyConfigLayer
	case owner == "" && layer != data.LayerBundle:
		return "", errs.ErrEmptyConfigOwner
	case key == "":
		return "", errs.ErrEmptyConfigKey
	}

	lookupKey := fmt.Sprintf("%s||%s||%s||%s", bundle, layer, owner, key)

	return strings.ToLower(lookupKey), nil
}
