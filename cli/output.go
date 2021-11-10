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
	"io"
	"os"
	"strings"
	"text/template"

	"github.com/Masterminds/sprig"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

const ErrorTemplate = `Error: {{ .Error }}`

type Errorable interface {
	Error() string
	HasErr() bool
	SetErr(err error) Errorable
	SetSuccess(b bool)
}

type CommandResult struct {
	Success bool
	Err     string `json:"Error,omitempty" yaml:"error,omitempty"`
}

func (c *CommandResult) Error() string {
	return c.Err
}

func (c *CommandResult) HasErr() bool {
	return c.Err != ""
}

func (c *CommandResult) SetErr(err error) Errorable {
	c.Err = err.Error()
	return c
}

func (c *CommandResult) SetSuccess(b bool) {
	c.Success = b
}

func OutputError(cmd *cobra.Command, o Errorable, e error) error {
	// Silence command errors, because we want to control the output.
	cmd.SilenceErrors = true
	o.SetErr(e)

	if err := Output(cmd, o, ErrorTemplate); err != nil {
		return err
	}

	return o
}

func OutputErrorMessage(cmd *cobra.Command, o Errorable, message string) error {
	// Silence command errors, because we want to control the output.
	cmd.SilenceErrors = true
	o.SetErr(fmt.Errorf(message))

	if err := Output(cmd, o, `{{ .Error }}`); err != nil {
		return err
	}

	return o
}

func OutputSuccess(cmd *cobra.Command, o Errorable, tmpl string) error {
	o.SetSuccess(true)
	return Output(cmd, o, tmpl)
}

func Output(cmd *cobra.Command, o Errorable, tmpl string) error {
	var text string

	switch f := strings.ToLower(FlagGortFormat); f {
	case "json":
		b, err := json.MarshalIndent(o, "", "  ")
		if err != nil {
			err = fmt.Errorf("Output error: failed to marshal as json: %w", err)
			fmt.Fprintln(os.Stderr, err.Error())
			return err
		}
		text = string(b)

	case "yaml":
		b, err := yaml.Marshal(o)
		if err != nil {
			err = fmt.Errorf("Output error: failed to marshal as yaml: %w", err)
			fmt.Fprintln(os.Stderr, err.Error())
			return err
		}
		text = string(b)

	case "text":
		t, err := template.New("template").Funcs(template.FuncMap(sprig.FuncMap())).Parse(tmpl)
		if err != nil {
			err = fmt.Errorf("Output error: failed to parse response template: %w", err)
			fmt.Fprintln(os.Stderr, err.Error())
			return err
		}
		b := &bytes.Buffer{}
		if err := t.Execute(b, o); err != nil {
			err = fmt.Errorf("Output error: failed to execute response template: %w", err)
			fmt.Fprintln(os.Stderr, err.Error())
			return err
		}
		text = b.String()

	default:
		err := fmt.Errorf("Output error: unsupported output format: %s", f)
		fmt.Fprintln(os.Stderr, err.Error())
		return err
	}

	var w io.Writer
	if o.HasErr() {
		w = os.Stderr
	} else {
		w = os.Stdout
	}

	fmt.Fprintln(w, strings.TrimSpace(text))

	return nil
}
