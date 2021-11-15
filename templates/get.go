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
	"fmt"

	"github.com/getgort/gort/config"
	"github.com/getgort/gort/data"
)

const (
	// DefaultCommand is a template used to format the outputs from successfully
	// executed commands.
	DefaultCommand = `{{ text | monospace true }}{{ .Response.Out }}{{ endtext }}`

	// CommandError is a template used to format the error messages produced
	// by commands that return with a non-zero status.
	DefaultCommandError = `{{ header | color "#FF0000" | title .Response.Title }}
{{ text }}The pipeline failed planning the invocation:{{ endtext }}
{{ text | monospace true }}{{ .Request.Bundle.Name }}:{{ .Request.Command.Name }} {{ .Request.Parameters }}{{ endtext }}
{{ text }}The specific error was:{{ endtext }}
{{ text | monospace true }}{{ .Response.Out }}{{ endtext }}`

	// Message is a template used to format standard informative (non-error)
	// messages from the Gort system (not commands).
	DefaultMessage = `{{ text }}{{ .Response.Out }}{{ endtext }}`

	// MessageError is a template used to format error messages from the Gor
	// system (not commands).
	DefaultMessageError = `{{ header | color "#FF0000" | title .Response.Title }}
{{ text }}{{ .Response.Out }}{{ endtext }}`
)

var templateDefaults = data.Templates{
	Message:      DefaultMessage,
	MessageError: DefaultMessageError,
	Command:      DefaultCommand,
	CommandError: DefaultCommandError,
}

// Get returns the first defined template found in the following sequence:
// 1. Command
// 2. Bundle
// 3. Config
// 4. Default
func Get(cmd data.BundleCommand, bundle data.Bundle, tt data.TemplateType) (string, error) {
	// We really only need to check for an error on the first call. The
	// outcome won't change after this.
	switch template, err := cmd.Templates.Get(tt); {
	case err != nil:
		return "", err
	case template != "":
		return template, nil
	}

	if template, _ := bundle.Templates.Get(tt); template != "" {
		return template, nil
	}

	if template, _ := config.GetTemplates().Get(tt); template != "" {
		return template, nil
	}

	if template, _ := templateDefaults.Get(tt); template != "" {
		return template, nil
	}

	return "", fmt.Errorf("no default template for %s found", tt)
}
