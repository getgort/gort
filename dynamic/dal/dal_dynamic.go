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

package dal

import (
	"context"
	"fmt"

	"github.com/getgort/gort/data"
	"github.com/getgort/gort/dataaccess"
)

func NewDALDynamicConfiguration(data.DynamicConfigs) (*DALDynamicConfiguration, error) {
	da, err := dataaccess.Get()
	if err != nil {
		return nil, fmt.Errorf("couldn't create DAL-based dynamic config: %w", err)
	}

	return &DALDynamicConfiguration{da: da}, nil
}

type DALDynamicConfiguration struct {
	da dataaccess.DataAccess
}

func (c *DALDynamicConfiguration) Delete(ctx context.Context, layer data.ConfigurationLayer, bundle, owner, key string) error {
	return c.da.DynamicConfigurationDelete(ctx, layer, bundle, owner, key)
}

func (c *DALDynamicConfiguration) Exists(ctx context.Context, layer data.ConfigurationLayer, bundle, owner, key string) (bool, error) {
	return c.da.DynamicConfigurationExists(ctx, layer, bundle, owner, key)
}

func (c *DALDynamicConfiguration) Get(ctx context.Context, layer data.ConfigurationLayer, bundle, owner, key string) (data.DynamicConfiguration, error) {
	return c.da.DynamicConfigurationGet(ctx, layer, bundle, owner, key)
}

func (c *DALDynamicConfiguration) List(ctx context.Context, layer data.ConfigurationLayer, bundle, owner, key string) ([]data.DynamicConfiguration, error) {
	return c.da.DynamicConfigurationList(ctx, layer, bundle, owner, key)
}

func (c *DALDynamicConfiguration) Set(ctx context.Context, config data.DynamicConfiguration) error {
	exists, err := c.Exists(ctx, config.Layer, config.Bundle, config.Owner, config.Key)
	if err != nil {
		return err
	}

	if exists {
		if err := c.Delete(ctx, config.Layer, config.Bundle, config.Owner, config.Key); err != nil {
			return err
		}
	}

	return c.da.DynamicConfigurationCreate(ctx, config)
}
