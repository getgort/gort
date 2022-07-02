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
	"fmt"
	"reflect"
	"strings"
)

type IntFunc func(i int) int

type StringFunc func(i int) string

type Columnizer struct {
	columns []column
}

type column struct {
	name string
	f    interface{}
}

func (c *Columnizer) IntColumn(name string, f IntFunc) *Columnizer {
	c.columns = append(c.columns, column{name: name, f: f})
	return c
}

func (c *Columnizer) StringColumn(name string, f StringFunc) *Columnizer {
	c.columns = append(c.columns, column{name: name, f: f})
	return c
}

// Will panic if o isn't a slice.
func (c *Columnizer) Format(o interface{}) []string {
	// Build meta-format string
	headers := strings.Builder{}
	meta := strings.Builder{}

	columnWidths := make([]interface{}, len(c.columns))
	columnNames := make([]interface{}, len(c.columns))

	rv := reflect.ValueOf(o)
	length := rv.Len()

	// Loop over each column
	for i, col := range c.columns {
		columnNames[i] = col.name
		columnWidths[i] = len(col.name)

		headers.WriteString("%%-%ds")

		switch f := col.f.(type) {
		case IntFunc:
			meta.WriteString("%%-%dd")
			for j := 0; j < length; j++ {
				w := len(fmt.Sprintf("%d", f(j)))
				if w > columnWidths[i].(int) {
					columnWidths[i] = w
				}
			}

		case StringFunc:
			meta.WriteString("%%-%ds")
			for j := 0; j < length; j++ {
				w := len(f(j))
				if w > columnWidths[i].(int) {
					columnWidths[i] = w
				}
			}

		default:
			panic(fmt.Sprintf("unknown type: %T", f))
		}
	}

	for i := range columnWidths {
		w := columnWidths[i].(int)
		columnWidths[i] = w + 3
	}

	headerFormat := fmt.Sprintf(headers.String(), columnWidths...)
	metaFormat := fmt.Sprintf(meta.String(), columnWidths...)

	var lines []string
	lines = append(lines, fmt.Sprintf(headerFormat, columnNames...))

	for row := 0; row < length; row++ {
		values := make([]interface{}, len(c.columns))

		for i, col := range c.columns {
			switch f := col.f.(type) {
			case IntFunc:
				values[i] = f(row)

			case StringFunc:
				values[i] = f(row)

			default:
				panic(fmt.Sprintf("unknown type: %T", f))
			}
		}

		lines = append(lines, fmt.Sprintf(metaFormat, values...))
	}

	return lines
}

// Print is a convenience function that prints the table to stdout.
func (c *Columnizer) Print(o interface{}) {
	for _, line := range c.Format(o) {
		fmt.Println(line)
	}
}
