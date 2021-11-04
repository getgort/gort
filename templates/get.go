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
	DefaultDefault = `{{ text }}{{ .Response.Out }}{{ endtext }}`

	DefaultMessage = `{{ text }}{{ .Response.Out }}{{ endtext }}`

	DefaultMessageError = `{{ header | color "#FF0000" }}{{ .Response.Title }}{{ endheader }}
{{ text }}{{ .Response.Out }}{{ endtext }}`

	DefaultCommand = `{{ text | monospace true }}{{ .Response.Out }}{{ endtext }}`

	DefaultCommandError = `{{ header | color "#FF0000" }}{{ .Response.Title }}{{ endheader }}
{{ text }}The pipeline failed planning the invocation:{{ endtext }}
{{ text | monospace true }}{{ .Request.Bundle.Name }}:{{ .Request.Command.Name }} {{ .Request.Parameters }}{{ endtext }}
{{ text }}The specific error was:{{ endtext }}
{{ text | monospace true }}{{ .Response.Out }}{{ endtext }}`
)

const (
	Default      TemplateType = "default"
	Message      TemplateType = "message"
	MessageError TemplateType = "message_error"
	Command      TemplateType = "command"
	CommandError TemplateType = "command_error"
)

type TemplateType string

var templateDefaults = data.Templates{
	Default:      DefaultDefault,
	Message:      DefaultMessage,
	MessageError: DefaultMessageError,
	Command:      DefaultCommand,
	CommandError: DefaultCommandError,
}

// Get returns the first defined template found in the following sequence:
// 1. Command (type-specific, then default)
// 2. Bundle (type-specific, then default)
// 3. Config (type-specific, then default)
// 4. Default (type-specific, then default)
func Get(cmd data.BundleCommand, bundle data.Bundle, tt TemplateType) (string, error) {
	// We really only need to check for an error on the first call. The
	// outcome won't change after this.
	switch template, err := cmd.Templates.Get(string(tt)); {
	case err != nil:
		return "", err
	case template != "":
		return template, nil
	}

	if template, _ := cmd.Templates.Get(string(Default)); template != "" {
		return template, nil
	}

	if template, _ := bundle.Templates.Get(string(tt)); template != "" {
		return template, nil
	}

	if template, _ := bundle.Templates.Get(string(Default)); template != "" {
		return template, nil
	}

	if template, _ := config.GetTemplates().Get(string(tt)); template != "" {
		return template, nil
	}

	if template, _ := config.GetTemplates().Get(string(Default)); template != "" {
		return template, nil
	}

	if template, _ := templateDefaults.Get(string(tt)); template != "" {
		return template, nil
	}

	if template, _ := templateDefaults.Get(string(Default)); template != "" {
		return template, nil
	}

	return "", fmt.Errorf("no default template for %s found", tt)
}
