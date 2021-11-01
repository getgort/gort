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

package data

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

var request = CommandRequest{
	RequestID: 1,
	UserName:  "user",
}

func TestNewCommandResponseEnvelope(t *testing.T) {
	e := NewCommandResponseEnvelope(request)

	assert.Equal(t, int64(1), e.Request.RequestID)
	assert.Equal(t, "user", e.Request.UserName)
}

func TestNewCommandResponseEnvelope_WithExitCode0(t *testing.T) {
	e := NewCommandResponseEnvelope(request, WithExitCode(0))

	assert.Equal(t, int16(0), e.Data.ExitCode)
	assert.False(t, e.Data.IsError)
}

func TestNewCommandResponseEnvelope_WithExitCode1(t *testing.T) {
	e := NewCommandResponseEnvelope(request, WithExitCode(1))

	assert.Equal(t, int16(1), e.Data.ExitCode)
	assert.True(t, e.Data.IsError)
}

func TestNewCommandResponseEnvelope_WithError(t *testing.T) {
	const title = "Error"
	const msg = "this is an error"
	const code int16 = 1
	err := fmt.Errorf(msg)

	e := NewCommandResponseEnvelope(request, WithError(title, err, code))

	assert.Equal(t, e.Data.Error, err)
	assert.Equal(t, e.Data.ExitCode, code)
	assert.True(t, e.Data.IsError)
	assert.Equal(t, e.Response.Lines, []string{msg})
	assert.Equal(t, e.Response.Out, msg)
	assert.Equal(t, e.Response.Payload, msg)
	assert.Equal(t, e.Response.Title, title)
}

func TestNewCommandResponseEnvelope_WithResponseLines(t *testing.T) {
	message := "this is a\ntwo-line message"
	lines := []string{"this is a", "two-line message"}

	e := NewCommandResponseEnvelope(request, WithResponseLines(lines))

	assert.Equal(t, e.Response.Lines, lines)
	assert.Equal(t, e.Response.Out, message)
	assert.False(t, e.Response.IsStructured)
	assert.Equal(t, e.Response.Payload, message)
	assert.False(t, e.Data.IsError)
}

func TestNewCommandResponseEnvelope_WithStructuredResponseLines(t *testing.T) {
	message := `{ "Name": "Matt" }`
	lines := []string{message}

	e := NewCommandResponseEnvelope(request, WithResponseLines(lines))

	assert.Equal(t, e.Response.Lines, lines)
	assert.Equal(t, e.Response.Out, message)
	assert.True(t, e.Response.IsStructured)
	assert.False(t, e.Data.IsError)

	p, ok := e.Response.Payload.(map[string]interface{})
	assert.True(t, ok)
	assert.Equal(t, "Matt", p["Name"])
}
