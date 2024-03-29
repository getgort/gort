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

package templates

import (
	"testing"

	"github.com/getgort/gort/bundles"
	"github.com/getgort/gort/data"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadBundleFromFile(t *testing.T) {
	bundle, err := bundles.LoadBundleFromFile("../testing/test-bundle.yml")
	if err != nil {
		t.Error(err.Error())
	}
	cmd := *bundle.Commands["echox"]

	template, err := Get(cmd, bundle, data.TemplateType("foo"))
	require.Equal(t, "", template)
	require.Error(t, err)

	template, err = Get(cmd, bundle, data.Command)
	assert.Equal(t, "Template:Command:Command", template)
	assert.NoError(t, err)

	template, err = Get(cmd, bundle, data.CommandError)
	assert.Equal(t, "Template:Command:CommandError", template)
	assert.NoError(t, err)

	template, err = Get(cmd, bundle, data.Message)
	assert.Equal(t, "Template:Command:Message", template)
	assert.NoError(t, err)

	template, err = Get(cmd, bundle, data.MessageError)
	assert.Equal(t, "Template:Command:MessageError", template)
	assert.NoError(t, err)

	cmd.Templates = data.Templates{}

	template, err = Get(cmd, bundle, data.Command)
	assert.Equal(t, "Template:Bundle:Command", template)
	assert.NoError(t, err)

	template, err = Get(cmd, bundle, data.CommandError)
	assert.Equal(t, "Template:Bundle:CommandError", template)
	assert.NoError(t, err)

	template, err = Get(cmd, bundle, data.Message)
	assert.Equal(t, "Template:Bundle:Message", template)
	assert.NoError(t, err)

	template, err = Get(cmd, bundle, data.MessageError)
	assert.Equal(t, "Template:Bundle:MessageError", template)
	assert.NoError(t, err)

	bundle.Templates = data.Templates{}

	template, err = Get(cmd, bundle, data.Command)
	assert.Equal(t, DefaultCommand, template)
	assert.NoError(t, err)

	template, err = Get(cmd, bundle, data.CommandError)
	assert.Equal(t, DefaultCommandError, template)
	assert.NoError(t, err)

	template, err = Get(cmd, bundle, data.Message)
	assert.Equal(t, DefaultMessage, template)
	assert.NoError(t, err)

	template, err = Get(cmd, bundle, data.MessageError)
	assert.Equal(t, DefaultMessageError, template)
	assert.NoError(t, err)
}
