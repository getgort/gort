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

package dynamic

import (
	"context"
	"fmt"

	"github.com/getgort/gort/config"
	"github.com/getgort/gort/data"
	"github.com/getgort/gort/dynamic/dal"
)

// DynamicConfiguration is the interface used to interact with the dynamic
// configuration backend.
type DynamicConfiguration interface {
	Delete(ctx context.Context, layer data.ConfigurationLayer, bundle, owner, key string) error
	Exists(ctx context.Context, layer data.ConfigurationLayer, bundle, owner, key string) (bool, error)
	Get(ctx context.Context, layer data.ConfigurationLayer, bundle, owner, key string) (data.DynamicConfiguration, error)
	List(ctx context.Context, layer data.ConfigurationLayer, bundle, owner, key string) ([]data.DynamicConfiguration, error)
	Set(ctx context.Context, config data.DynamicConfiguration) error
}

// Get provides an interface to the dynamic configuration backend. If no
// backend is specified, the default (DAL) implementation is used.
func Get() (DynamicConfiguration, error) {
	dynamicConfigs := config.GetDynamicConfigs()

	if config.IsUndefined(dynamicConfigs) {
		return dal.NewDALDynamicConfiguration(dynamicConfigs)
	}

	return nil, fmt.Errorf("unsupported dynamic configuration backend: %s", dynamicConfigs.Backend)
}
