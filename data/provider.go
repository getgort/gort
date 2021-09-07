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

//// The wrappers for the "slack" section.
//// Other providers will eventually get their own sections

// Provider is the general interface for all providers.
// Currently only Slack is supported, with HipChat coming in time.
type Provider interface{}

// AbstractProvider is used to contain the general properties shared by
// all providers.
type AbstractProvider struct {
	BotName string `yaml:"bot_name,omitempty"`
	Name    string `yaml:"name,omitempty"`
}

// SlackProvider is the data wrapper for a Slack App provider.
type SlackProvider struct {
	AbstractProvider `yaml:",inline"`
	IconURL          string `yaml:"icon_url,omitempty"`

	// App and Bot tokens, used for Socket mode.
	AppToken string `yaml:"app_token,omitempty"`
	BotToken string `yaml:"bot_token,omitempty"`

	// Deprecated, used for Classic Slack apps
	APIToken string `yaml:"api_token,omitempty"`
}
