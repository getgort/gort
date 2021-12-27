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

	"github.com/getgort/gort/data"
	"github.com/getgort/gort/data/rest"
)

var dataAccess = &InMemoryDataAccess{
	bundles: make(map[string]*data.Bundle),
	configs: make(map[string]*data.DynamicConfiguration),
	groups:  make(map[string]*rest.Group),
	roles:   make(map[string]*rest.Role),
	users:   make(map[string]*rest.User),
}

// InMemoryDataAccess is an entirely in-memory representation of a data access layer.
// Great for testing and development. Terrible for production.
type InMemoryDataAccess struct {
	bundles map[string]*data.Bundle
	configs map[string]*data.DynamicConfiguration
	groups  map[string]*rest.Group
	roles   map[string]*rest.Role
	users   map[string]*rest.User
}

// NewInMemoryDataAccess returns a new InMemoryDataAccess instance.
func NewInMemoryDataAccess() *InMemoryDataAccess {
	return dataAccess
}

// Initialize initializes an InMemoryDataAccess instance.
func (da *InMemoryDataAccess) Initialize(ctx context.Context) error {
	return nil
}

func Reset() {
	dataAccess.bundles = make(map[string]*data.Bundle)
	dataAccess.configs = make(map[string]*data.DynamicConfiguration)
	dataAccess.groups = make(map[string]*rest.Group)
	dataAccess.roles = make(map[string]*rest.Role)
	dataAccess.users = make(map[string]*rest.User)
}
