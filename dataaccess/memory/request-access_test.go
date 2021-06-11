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
	"fmt"
	"testing"
	"time"

	"github.com/getgort/gort/data"
	"github.com/stretchr/testify/assert"
)

func testRequestAccess(t *testing.T) {
	t.Run("testRequestBegin", testRequestBegin)
	t.Run("testRequestUpdate", testRequestUpdate)
	t.Run("testRequestClose", testRequestClose)
}

func testRequestBegin(t *testing.T) {
	ctx = context.Background()
	bundle, err := getTestBundle()
	assert.NoError(t, err)

	entry := data.CommandEntry{
		Bundle:  bundle,
		Command: *bundle.Commands["echox"],
	}

	req := data.CommandRequest{
		CommandEntry: entry,
		Adapter:      "testAdapter",
		ChannelID:    "testChannelID",
		Parameters:   []string{"foo", "bar"},
		Timestamp:    time.Now(),
		UserID:       "testUserID   ",
		UserEmail:    "testUserEmail",
		UserName:     "testUserName ",
	}

	t.Log("Request ID before: ", req.RequestID)
	assert.Zero(t, req.RequestID)

	err = da.RequestBegin(ctx, &req)
	assert.NoError(t, err)

	t.Log("Request ID after: ", req.RequestID)
	assert.NotZero(t, req.RequestID)

	err = da.RequestBegin(ctx, &req)
	assert.Error(t, err)
}

func testRequestUpdate(t *testing.T) {
	ctx = context.Background()
	bundle, err := getTestBundle()
	assert.NoError(t, err)

	entry := data.CommandEntry{
		Bundle:  bundle,
		Command: *bundle.Commands["echox"],
	}

	req := data.CommandRequest{
		CommandEntry: entry,
		Adapter:      "testAdapter",
		ChannelID:    "testChannelID",
		Parameters:   []string{"foo", "bar"},
		Timestamp:    time.Now(),
		UserID:       "testUserID   ",
		UserEmail:    "testUserEmail",
		UserName:     "testUserName ",
	}

	err = da.RequestUpdate(ctx, req)
	assert.Error(t, err)

	req.RequestID = 1

	err = da.RequestUpdate(ctx, req)
	assert.NoError(t, err)
}

func testRequestClose(t *testing.T) {
	ctx = context.Background()
	bundle, err := getTestBundle()
	assert.NoError(t, err)

	entry := data.CommandEntry{
		Bundle:  bundle,
		Command: *bundle.Commands["echox"],
	}

	req := data.CommandRequest{
		CommandEntry: entry,
		Adapter:      "testAdapter",
		ChannelID:    "testChannelID",
		Parameters:   []string{"foo", "bar"},
		RequestID:    1,
		Timestamp:    time.Now(),
		UserID:       "testUserID   ",
		UserEmail:    "testUserEmail",
		UserName:     "testUserName ",
	}

	res := data.CommandResponse{
		Command:  req,
		Duration: time.Second,
		Status:   1,
		Error:    fmt.Errorf("Fake error"),
	}

	err = da.RequestClose(ctx, res)
	assert.NoError(t, err)
}
