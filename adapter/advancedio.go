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

package adapter

import (
	"github.com/getgort/gort/command"
	"github.com/getgort/gort/types"
)

// Input is used for advanced command input.
type Input struct {
	Channel  ChannelInfo
	Command  CommandInfo
	Provider ProviderInfo
	User     UserInfo
}

// CommandInfo represents a command typed in by a user. Unlike
// command.Command, it can be marshaled into meaningful JSON.
type CommandInfo struct {
	Bundle     string
	Command    string
	Options    map[string]string
	Parameters []string
}

// NewCommandInfo creates a new CommandInfo from a command.Command value.
func NewCommandInfo(c command.Command) CommandInfo {
	var options = map[string]string{}
	var params []string

	for _, o := range c.Options {
		options[o.Name] = o.Value.String()
	}
	for _, p := range c.Parameters {
		var s string

		if v, ok := p.(types.StringValue); ok {
			s = v.V
		} else {
			s = p.String()
		}

		params = append(params, s)
	}

	return CommandInfo{
		Bundle:     c.Bundle,
		Command:    c.Command,
		Options:    options,
		Parameters: params,
	}
}
