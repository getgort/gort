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

func TestLoadBundle(t *testing.T) {
	b, err := LoadBundle("../testing/test-bundle.yml")
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
	assert.Equal(t, "echox", b.Commands["echox"].Name)
	assert.Equal(t, "Echos back anything sent to it, all at once.", b.Commands["echox"].Description)
	assert.Equal(t, []string{"/bin/echo"}, b.Commands["echox"].Executable)
	assert.Len(t, b.Commands["echox"].Rules, 1)
	assert.Equal(t, "must have test:echox", b.Commands["echox"].Rules[0])
}
