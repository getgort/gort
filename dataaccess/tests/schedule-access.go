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

package tests

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/getgort/gort/command"
	"github.com/getgort/gort/data"

	"github.com/stretchr/testify/require"
)

func (da DataAccessTester) testScheduleAccess(t *testing.T) {
	t.Run("testScheduleCreate", da.testScheduleCreate)
	t.Run("testScheduleDelete", da.testScheduleDelete)
	t.Run("testSchedulesGet", da.testSchedulesGet)
}

func (da DataAccessTester) testScheduleCreate(t *testing.T) {
	cmd, _ := command.TokenizeAndParse("whoami")
	s1 := data.ScheduledCommand{
		CommandEntry: data.CommandEntry{},
		Command:      cmd,
		Adapter:      "MySlack",
		ChannelID:    "channel",
		UserID:       "user",
		UserEmail:    "email",
		UserName:     "name",
		Cron:         "@every 1m",
	}

	err := da.ScheduleCreate(da.ctx, &s1)
	require.NoError(t, err)
	require.NotZero(t, s1.ScheduleID)
	defer da.ScheduleDelete(da.ctx, s1.ScheduleID)
}

func (da DataAccessTester) testSchedulesGet(t *testing.T) {
	cmd, _ := command.TokenizeAndParse("whoami")
	s1 := data.ScheduledCommand{
		CommandEntry: data.CommandEntry{},
		Command:      cmd,
		Adapter:      "MySlack",
		ChannelID:    "channel",
		UserID:       "user",
		UserEmail:    "email",
		UserName:     "name",
		Cron:         "@every 1m",
	}

	s2 := data.ScheduledCommand{
		CommandEntry: data.CommandEntry{},
		Command:      cmd,
		Adapter:      "",
		ChannelID:    "",
		UserID:       "",
		UserEmail:    "",
		UserName:     "",
		Cron:         "",
	}

	err := da.ScheduleCreate(da.ctx, &s1)
	require.NoError(t, err)
	assert.NotZero(t, s1.ScheduleID)
	defer da.ScheduleDelete(da.ctx, s1.ScheduleID)

	list, err := da.SchedulesGet(da.ctx)
	require.NoError(t, err)
	assert.Equal(t, 1, len(list))

	err = da.ScheduleCreate(da.ctx, &s2)
	require.NoError(t, err)
	assert.NotZero(t, s2.ScheduleID)
	defer da.ScheduleDelete(da.ctx, s2.ScheduleID)
	assert.NotEqual(t, s2.ScheduleID, s1.ScheduleID)

	list, err = da.SchedulesGet(da.ctx)
	require.NoError(t, err)
	assert.Equal(t, 2, len(list))
}

func (da DataAccessTester) testScheduleDelete(t *testing.T) {
	cmd, _ := command.TokenizeAndParse("whoami")
	s1 := data.ScheduledCommand{
		CommandEntry: data.CommandEntry{},
		Command:      cmd,
		Adapter:      "MySlack",
		ChannelID:    "channel",
		UserID:       "user",
		UserEmail:    "email",
		UserName:     "name",
		Cron:         "@every 1m",
	}

	err := da.ScheduleCreate(da.ctx, &s1)
	require.NoError(t, err)
	assert.NotZero(t, s1.ScheduleID)

	list, err := da.SchedulesGet(da.ctx)
	require.NoError(t, err)
	assert.Equal(t, 1, len(list))

	err = da.ScheduleDelete(da.ctx, s1.ScheduleID)
	require.NoError(t, err)

	list, err = da.SchedulesGet(da.ctx)
	require.NoError(t, err)
	assert.Zero(t, len(list))
}
