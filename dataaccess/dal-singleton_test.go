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

package dataaccess

import (
	"context"
	"testing"
	"time"

	"github.com/getgort/gort/config"
	"github.com/getgort/gort/dataaccess/memory"
	"github.com/getgort/gort/dataaccess/postgres"
	"github.com/stretchr/testify/assert"
)

func TestInitializeDataAccess(t *testing.T) {
	err := config.Initialize("../testing/config/no-database.yml")
	assert.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	initializeDataAccess(ctx)

	for state := range Updates() {
		if state == StateInitialized {
			break
		}

		if state == StateError {
			t.Error("Failed to initialize DAL")
		}
	}

	dal, err := Get()
	assert.NoError(t, err)
	assert.NotNil(t, dal)
}

// If there's no "database" section in the config, use the in-memopry DAL.
func TestUseInMemory(t *testing.T) {
	err := config.Initialize("../testing/config/no-database.yml")
	assert.NoError(t, err)

	da := getCorrectDataAccess()
	expected := memory.NewInMemoryDataAccess()
	assert.IsType(t, expected, da)
}

// If there's a "database" section in the config, use the Postgres DAL.
func TestUsePostgres(t *testing.T) {
	err := config.Initialize("../testing/config/complete.yml")
	assert.NoError(t, err)

	da := getCorrectDataAccess()
	expected := postgres.NewPostgresDataAccess(config.GetDatabaseConfigs())
	assert.IsType(t, expected, da)
}
