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

package client

import (
	"fmt"
	"testing"

	yaml "gopkg.in/yaml.v3"
)

// [defaults]
// profile=gort.mycompany.com

// [gort.mycompany.com]
// url=http://gort.mycompany.com:4000
// password=BgEocTFGzON$U39srRt5^fi(ZD0KxV*1
// user=admin

func TestProfileHumanEyeball(t *testing.T) {
	defaultProfile := "gort.mycompany.com"

	profile := Profile{Profiles: make(map[string]ProfileEntry)}
	profile.Defaults.Profile = defaultProfile
	profile.Profiles[defaultProfile] = ProfileEntry{
		URLString: "http://gort.mycompany.com:4000",
		Password:  "BgEocTFGzON$U39srRt5^fi(ZD0KxV*1",
		Username:  "admin",
	}
	profile.Profiles["profile2"] = ProfileEntry{
		URLString: "http://gort.myothercompany.com:4000",
		Password:  "SomeWackyPassword!11!1",
		Username:  "gort-admin",
	}

	y, err := yaml.Marshal(profile)
	if err != nil {
		fmt.Printf("err: %v\n", err)
		return
	}

	fmt.Println(string(y))
}
