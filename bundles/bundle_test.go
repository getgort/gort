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

package bundles

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoadBundleFromFile(t *testing.T) {
	b, err := LoadBundleFromFile("../testing/test-bundle.yml")
	if err != nil {
		t.Error(err.Error())
	}

	assert.Equal(t, "test", b.Name)
	assert.Equal(t, "0.0.1", b.Version)
	assert.Equal(t, "Matt Titmus <matthew.titmus@gmail.com>", b.Author)
	assert.Equal(t, "https://guide.getgort.io", b.Homepage)
	assert.Equal(t, "A test bundle.", b.Description)
	assert.Equal(t, "This is test bundle.\nThere are many like it, but this one is mine.", b.LongDescription)
	assert.Len(t, b.Permissions, 1)
	assert.Equal(t, "ubuntu", b.Docker.Image)
	assert.Equal(t, "20.04", b.Docker.Tag)
	assert.Len(t, b.Commands, 1)

	// Bundle templates
	assert.Equal(t, "Template:Bundle:CommandError", b.Templates.CommandError)
	assert.Equal(t, "Template:Bundle:Command", b.Templates.Command)
	assert.Equal(t, "Template:Bundle:MessageError", b.Templates.MessageError)
	assert.Equal(t, "Template:Bundle:Message", b.Templates.Message)

	cmd := b.Commands["echox"]
	assert.Equal(t, "echox", cmd.Name)
	assert.Equal(t, "Write arguments to the standard output.", cmd.Description)
	assert.Equal(t, `Write arguments to the standard output.

Usage:
  test:echox [string ...]`, cmd.LongDescription)
	assert.Equal(t, []string{"/bin/echo"}, cmd.Executable)
	assert.Len(t, cmd.Rules, 1)
	assert.Equal(t, "must have test:echox", cmd.Rules[0])

	// Command templates
	assert.Equal(t, "Template:Command:CommandError", cmd.Templates.CommandError)
	assert.Equal(t, "Template:Command:Command", cmd.Templates.Command)
	assert.Equal(t, "Template:Command:MessageError", cmd.Templates.MessageError)
	assert.Equal(t, "Template:Command:Message", cmd.Templates.Message)
}
