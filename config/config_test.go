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

package config

import (
	"os"
	"testing"
	"time"

	"github.com/getgort/gort/data"
	"github.com/stretchr/testify/assert"
)

func TestLoad(t *testing.T) {
	config, err := load("../testing/config/complete.yml")
	if err != nil {
		t.Error(err.Error())
		t.FailNow()
	}

	cglobal := config.GlobalConfigs
	assert.NotNil(t, cglobal)
	assert.Equal(t, time.Minute, cglobal.CommandTimeout)

	cgort := config.GortServerConfigs
	assert.NotNil(t, cgort)
	assert.Equal(t, true, cgort.AllowSelfRegistration)
	assert.Equal(t, ":4000", cgort.APIAddress)
	assert.Equal(t, "localhost", cgort.APIURLBase)
	assert.Equal(t, true, cgort.DevelopmentMode)
	assert.Equal(t, true, cgort.EnableSpokenCommands)

	cdb := config.DatabaseConfigs
	assert.NotNil(t, cdb)
	assert.Equal(t, "localhost", cdb.Host)
	assert.Equal(t, 5432, cdb.Port)
	assert.Equal(t, "gort", cdb.User)
	assert.Equal(t, "veryKleverPassw0rd!", cdb.Password)
	assert.Equal(t, true, cdb.SSLEnabled)

	assert.Equal(t, 30*time.Second, cdb.ConnectionMaxIdleTime)
	assert.Equal(t, 5*time.Minute, cdb.ConnectionMaxLifetime)
	assert.Equal(t, 2, cdb.MaxIdleConnections)
	assert.Equal(t, 4, cdb.MaxOpenConnections)

	cd := config.DockerConfigs
	assert.NotNil(t, cd)
	assert.Equal(t, "unix:///var/run/docker.sock", cd.DockerHost)

	cs := config.SlackProviders
	assert.NotNil(t, cs)
	assert.NotEmpty(t, cs)
	assert.Equal(t, "MyWorkspace", cs[0].Name)
	assert.Equal(t, "xoxb-210987654321-123456789012-nyWJ3U4JoWuUtaUkRPKn0dJR", cs[0].APIToken)
	assert.Equal(t, "https://emoji.slack-edge.com/T023V8ZFQEQ/gort/78a0c1607eeb1f29.png", cs[0].IconURL)
	assert.Equal(t, "Gort", cs[0].BotName)

	cj := config.JaegerConfigs
	assert.NotNil(t, cj)
	assert.NotEmpty(t, cj)
	assert.Equal(t, cj.Endpoint, "http://localhost:14268/api/traces")
	assert.Equal(t, cj.Username, "gort")
	assert.Equal(t, cj.Password, "veryKleverPassw0rd!")
}

func TestIsUndefinedNil(t *testing.T) {
	id := IsUndefined(nil)
	assert.True(t, id)
}

func TestIsUndefinedFalse(t *testing.T) {
	c, err := load("../testing/config/complete.yml")
	if err != nil {
		t.Error(err.Error())
		t.FailNow()
	}

	id := IsUndefined(c)
	assert.False(t, id)
}

func TestIsUndefinedTrue(t *testing.T) {
	c := data.GortConfig{}
	id := IsUndefined(c)
	assert.True(t, id)
}

func TestIsUndefinedTrue2(t *testing.T) {
	c, err := load("../testing/config/no-database.yml")
	if err != nil {
		t.Error(err.Error())
		t.FailNow()
	}

	id := IsUndefined(c.DatabaseConfigs)
	assert.True(t, id)
}

func TestStandardizeDatabaseConfigNone(t *testing.T) {
	config, err := load("../testing/config/no-database-password.yml")
	if err != nil {
		t.Error(err.Error())
		t.FailNow()
	}

	standardizeDatabaseConfig(&config.DatabaseConfigs)

	dbp := config.DatabaseConfigs.Password

	assert.Empty(t, dbp)
}

func TestStandardizeDatabaseConfigDefined(t *testing.T) {
	const expected = "someRandomPassword"

	err := os.Setenv(EnvDatabasePassword, expected)
	if err != nil {
		t.Error(err.Error())
		t.FailNow()
	}

	config, err := load("../testing/config/no-database-password.yml")
	if err != nil {
		t.Error(err.Error())
		t.FailNow()
	}

	standardizeDatabaseConfig(&config.DatabaseConfigs)

	dbp := config.DatabaseConfigs.Password

	assert.NotEmpty(t, dbp)
	assert.Equal(t, dbp, expected)
}
