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
	"github.com/getgort/gort/dataaccess/memory"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestSchedulerMultiple tests that the one CommandRequest is issued per second
// over a specified number of seconds.
func TestSchedulerMultiple(t *testing.T) {
	const iterations = 5

	// Init Gort
	memory.Reset()
	ctx := context.Background()
	require.NoError(t, config.Initialize("../testing/config/no-database.yml"))

	da, err := dataaccess.Get()
	require.NoError(t, err)

	require.NoError(t, da.Initialize(ctx))
	user, err := dataaccess.Bootstrap(ctx, rest.User{
		Email:    "user@getgort.io",
		Username: "user",
	})
	require.NoError(t, err)

	scheduledCommands := StartScheduler()
	defer StopScheduler()
	entries, err := da.FindCommandEntry(ctx, "gort", "whoami")
	require.NoError(t, err)
	require.NotEmpty(t, entries)
	id, err := Schedule(ctx, data.ScheduledCommand{
		CommandEntry: entries[0],
		Command: command.Command{
			Bundle:     "gort",
			Command:    "whoami",
			Options:    make(map[string]command.CommandOption),
			Parameters: make(command.CommandParameters, 0),
			Original:   "whoami",
		},
		Cron:      "@every 1s",
		UserID:    "id",
		UserName:  user.Username,
		UserEmail: user.Email,
		Adapter:   "test adapter",
		ChannelID: "channel",
	})
	require.NoError(t, err)
	require.NotZero(t, id)

	for i := 0; i < iterations; {
		select {
		case <-time.After(iterations*time.Second + 50*time.Millisecond):
			t.FailNow()
		case req := <-scheduledCommands:
			assert.NotEqual(t, 0, req.RequestID)
			i++
		}
	}
}

func TestSchedulerPersistence(t *testing.T) {
	memory.Reset()
	ctx := context.Background()
	require.NoError(t, config.Initialize("../testing/config/no-database.yml"))

	da, err := dataaccess.Get()
	require.NoError(t, err)

	require.NoError(t, da.Initialize(ctx))
	user, err := dataaccess.Bootstrap(ctx, rest.User{
		Email:    "user@getgort.io",
		Username: "user",
	})
	require.NoError(t, err)

	_ = StartScheduler()

	entries, err := da.FindCommandEntry(ctx, "gort", "whoami")
	require.NoError(t, err)
	require.NotEmpty(t, entries)
	id, err := Schedule(ctx, data.ScheduledCommand{
		CommandEntry: entries[0],
		Command: command.Command{
			Bundle:     "gort",
			Command:    "whoami",
			Options:    make(map[string]command.CommandOption),
			Parameters: make(command.CommandParameters, 0),
			Original:   "whoami",
		},
		Cron:      "@every 1s",
		UserID:    "id",
		UserName:  user.Username,
		UserEmail: user.Email,
		Adapter:   "test adapter",
		ChannelID: "channel",
	})
	require.NoError(t, err)
	require.NotZero(t, id)

	require.Equal(t, 1, len(cron.Jobs()))

	StopScheduler()

	require.Equal(t, 0, len(cron.Jobs()))

	_ = StartScheduler()

	require.Equal(t, 1, len(cron.Jobs()))

	// TODO: StopScheduler bloack indefinitely due to
	// https://github.com/go-co-op/gocron/issues/355
	// Once this issue is fixed, we can stop the scheduler in this test.
	//StopScheduler()
	//require.Equal(t, 0, len(cron.Jobs()))
}
