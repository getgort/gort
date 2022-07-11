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

package scheduler

import (
	"context"
	"testing"
	"time"

	"github.com/getgort/gort/command"
	"github.com/getgort/gort/config"
	"github.com/getgort/gort/data"
	"github.com/getgort/gort/data/rest"
	"github.com/getgort/gort/dataaccess"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCommandSchedulerFull(t *testing.T) {
	// Init Gort
	require.NoError(t, config.Initialize("../testing/config/no-database.yml"))

	da, err := dataaccess.Get()
	require.NoError(t, err)

	require.NoError(t, da.Initialize(context.Background()))
	user, err := dataaccess.DoBootstrap(context.Background(), rest.User{
		Email:    "user@getgort.io",
		Username: "user",
	})
	require.NoError(t, err)

	scheduledCommands := StartScheduler()
	entries, err := da.FindCommandEntry(context.Background(), "gort", "whoami")
	require.NoError(t, err)
	require.NotEmpty(t, entries)
	require.NoError(t, Schedule(context.Background(), data.ScheduledCommand{
		CommandEntry: entries[0],
		Command: command.Command{
			Bundle:     "gort",
			Command:    "whoami",
			Options:    make(map[string]command.CommandOption),
			Parameters: make(command.CommandParameters, 0),
		},
		Cron:      "@every 1s",
		UserID:    "id",
		UserName:  user.Username,
		UserEmail: user.Email,
		Adapter:   "test adapter",
		ChannelID: "channel",
	}))

	select {
	case <-time.After(1*time.Second + 50*time.Millisecond):
		t.Fail()
	case req := <-scheduledCommands:
		assert.NotEqual(t, 0, req.RequestID)
		// success
	}
}
