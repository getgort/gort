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
)

type TemplateType string

const (
	// Command templates are used to format the outputs from successfully
	// executed commands.
	Command TemplateType = "command"

	// CommandError templates are used to format the error messages produced
	// by commands that return with a non-zero status.
	CommandError TemplateType = "command_error"

	// Message templates are used to format standard informative (non-error)
	// messages from the Gort system (not commands).
	Message TemplateType = "message"

	// MessageError templates are used to format error messages from the Gor
	// system (not commands).
	MessageError TemplateType = "message_error"
)

// Templates describes (or not) a set of templates that can be used to format
// command output. It is used in several places, including bundles, bundle
// commands, and the application config.
type Templates struct {
	// Command templates are used to format the outputs from successfully
	// executed commands.
	Command string `yaml:"command,omitempty" json:"command,omitempty"`

	// CommandError templates are used to format the error messages produced
	// by commands that return with a non-zero status.
	CommandError string `yaml:"command_error,omitempty" json:"command_error,omitempty"`

	// Message templates are used to format standard informative (non-error)
	// messages from the Gort system (not commands).
	Message string `yaml:"message,omitempty" json:"message,omitempty"`

	// MessageError templates are used to format error messages from the Gort
	// system (not commands).
	MessageError string `yaml:"message_error,omitempty" json:"message_error,omitempty"`
}

// Get returns a template string. If no template is defined for the given
// name/type, an empty string is returned. An invalid type returns an error.
func (t Templates) Get(tt TemplateType) (string, error) {
	switch tt {
	case Command:
		return t.Command, nil
	case CommandError:
		return t.CommandError, nil
	case Message:
		return t.Message, nil
	case MessageError:
		return t.MessageError, nil
	default:
		return "", fmt.Errorf("invalid template type %q", string(tt))
	}
}
