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
	"testing"

	"github.com/getgort/gort/data"
	"github.com/stretchr/testify/assert"
)

func TestLoadConfiguration(t *testing.T) {
	config, err := loadConfiguration("../testing/config/complete.yml")
	if err != nil {
		t.Error(err.Error())
	}

	cglobal := config.GlobalConfigs
	assert.NotNil(t, cglobal)
	assert.Equal(t, 60, cglobal.CommandTimeoutSeconds)

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
	assert.Equal(t, 10, cdb.PoolSize)
	assert.Equal(t, 15000, cdb.PoolTimeout)
	assert.Equal(t, 15000, cdb.QueryTimeout)

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

	cb := config.BundleConfigs
	assert.NotNil(t, cb)
	assert.Len(t, cb, 2)
	assert.Equal(t, "echo", cb[0].Name)
	assert.Equal(t, "A default bundle with echo commands.", cb[0].Description)
	assert.Equal(t, "getgort/relaytest", cb[0].Docker.Image)
	assert.Equal(t, "latest", cb[0].Docker.Tag)

	assert.Len(t, cb[0].Commands, 2)
	assert.Equal(t, "echo", cb[0].Commands["echo"].Name)
	assert.Equal(t, "Echos back anything sent to it, all at once.", cb[0].Commands["echo"].Description)
	assert.Equal(t, "/bin/echo", cb[0].Commands["echo"].Executable)

	assert.Equal(t, "splitecho", cb[0].Commands["splitecho"].Name)
	assert.Equal(t, "Echos back anything sent to it, one parameter at a time.", cb[0].Commands["splitecho"].Description)
	assert.Equal(t, "/opt/app/splitecho.sh", cb[0].Commands["splitecho"].Executable)
}

func TestIsUndefinedNil(t *testing.T) {
	id := IsUndefined(nil)
	assert.True(t, id)
}

func TestIsUndefinedFalse(t *testing.T) {
	c, err := loadConfiguration("../testing/config/complete.yml")
	if err != nil {
		t.Error(err.Error())
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
	c, err := loadConfiguration("../testing/config/no-database.yml")
	if err != nil {
		t.Error(err.Error())
	}

	id := IsUndefined(c.DatabaseConfigs)
	assert.True(t, id)
}
