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

package postgres

import (
	"context"
	"fmt"

	"github.com/getgort/gort/data"
)

func (da PostgresDataAccess) DynamicConfigurationCreate(_ context.Context, config data.DynamicConfiguration) error {
	return fmt.Errorf("not implemented")
}

func (da PostgresDataAccess) DynamicConfigurationDelete(_ context.Context, layer data.ConfigurationLayer, bundle, owner, key string) error {
	return fmt.Errorf("not implemented")
}

func (da PostgresDataAccess) DynamicConfigurationExists(_ context.Context, layer data.ConfigurationLayer, bundle, owner, key string) (bool, error) {
	return false, fmt.Errorf("not implemented")
}

func (da PostgresDataAccess) DynamicConfigurationGet(_ context.Context, layer data.ConfigurationLayer, bundle, owner, key string) (data.DynamicConfiguration, error) {
	return data.DynamicConfiguration{}, fmt.Errorf("not implemented")
}

func (da PostgresDataAccess) DynamicConfigurationList(_ context.Context, layer data.ConfigurationLayer, bundle, owner, key string) ([]data.DynamicConfiguration, error) {
	return nil, fmt.Errorf("not implemented")
}
