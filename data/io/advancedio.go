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

package io

import (
	"encoding/json"

	"github.com/getgort/gort/command"
	"github.com/getgort/gort/data/rest"
	"github.com/getgort/gort/types"
)

const (
	ActionReply = "reply"
	ActionReact = "react"
)

// CommandInfo represents a command typed in by a user. Unlike
// command.Command, it can be marshaled into meaningful JSON.
type CommandInfo struct {
	// Bundle is the name of the bundle used
	Bundle string

	// Command is the name of the command
	Command string

	// Options is the list of command options (i.e., flags) entered
	// by the user
	Options map[string]string

	// Parameters is the tokenized list of non-option command parameters
	Parameters []string
}

// AdvancedInput is used to construct the JSON blob that's passed into a
// command in lieu of parameters when the executed bundle command's advanced
// input is set to true.
type AdvancedInput struct {
	// Channel contains the basic information for the provider channel.
	Channel ChannelInfo

	// Command wraps the command execution request, as entered by the user
	Command CommandInfo

	// User contains the basic information about the provider
	Provider ProviderInfo

	// ProviderUser contains the basic information about the provider user
	ProviderUser UserInfo

	// GortUser is the Gort user triggering the command (any
	// credentials are scrubbed).
	GortUser rest.User

	MessageRef string
}

// String returns an un-indented JSON representation of the AdvancedInput
// instance.
func (ai AdvancedInput) String() string {
	b, _ := json.Marshal(ai)
	return string(b)
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

// AdvancedOutput is parsed from json output from bundle commands, and describes
// a variety of actions that can be performed by an adapter. Aside from Action,
// most fields are optional.
type AdvancedOutput struct {
	// Action is the thing for the adapter to do.
	Action string

	// ChannelId is the channel in which the thing should be done, if
	// applicable.
	ChannelID string

	// MessageRef is a way to refer to a unique message across any adapter. It
	// is a string to simplify transport and storage for users, but should
	// always be a json-encoded adapter.MessageRef.
	MessageRef string

	// Content is the content associated with an action, for example text for a
	// reply or the emoji for a reaction.
	Content string
}
