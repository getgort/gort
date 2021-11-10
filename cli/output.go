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

package cli

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
	"text/template"

	"github.com/Masterminds/sprig"
	yaml "gopkg.in/yaml.v3"
)

func Output(format string, o interface{}, tmpl string) error {
	var text string

	switch f := strings.ToLower(format); f {
	case "json":
		b, err := json.MarshalIndent(o, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal as json: %w", err)
		}
		text = string(b)

	case "yaml":
		b, err := yaml.Marshal(o)
		if err != nil {
			return fmt.Errorf("failed to marshal as yaml: %w", err)
		}
		text = string(b)

	case "text":
		t, err := template.New("template").Funcs(template.FuncMap(sprig.FuncMap())).Parse(tmpl)
		if err != nil {
			return fmt.Errorf("failed to parse template: %w", err)
		}
		b := &bytes.Buffer{}
		if err := t.Execute(b, o); err != nil {
			return fmt.Errorf("failed to execute template: %w", err)
		}
		text = b.String()

	default:
		return fmt.Errorf("unsupported output format: %s", f)
	}

	fmt.Println(strings.TrimSpace(text))
	return nil
}
