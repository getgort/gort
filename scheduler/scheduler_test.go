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

	"github.com/getgort/gort/data"
	"github.com/getgort/gort/dataaccess/memory"
	"github.com/stretchr/testify/assert"
)

const OneSecond = 1*time.Second + 50*time.Millisecond

func TestCommandSchedulerFull(t *testing.T) {
	cs := NewCommandScheduler(memory.NewInMemoryDataAccess())
	cs.Start()
	cs.Add(context.Background(), "@every 1s", data.CommandEntry{
		Bundle:  data.Bundle{},
		Command: data.BundleCommand{},
	}, make(data.CommandParameters, 0), "id", "email", "name", "test adapter", "channel")

	select {
	case <-time.After(OneSecond):
		t.Fail()
	case req := <-cs.Commands:
		assert.NotEqual(t, 0, req.RequestID)
		// success
	}
}
