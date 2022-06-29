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

package client_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/getgort/gort/client"
)

func TestAllowInsecure(t *testing.T) {
	var tests = []struct {
		Name         string
		ProfileEntry client.ProfileEntry
		ExpectErr    bool
	}{
		{
			Name: "allows secure URL",
			ProfileEntry: client.ProfileEntry{
				URLString: "https://example.com",
			},
		},
		{
			Name: "does not allow insecure URL",
			ProfileEntry: client.ProfileEntry{
				URLString: "http://example.com",
			},
			ExpectErr: true,
		},
		{
			Name: "allows insecure URL if allowInsecure==true",
			ProfileEntry: client.ProfileEntry{
				URLString:     "http://example.com",
				AllowInsecure: true,
			},
			ExpectErr: false,
		},
	}
	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			_, err := client.ConnectWithNewProfile(test.ProfileEntry, ".")
			if test.ExpectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
