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
	"strconv"

	"github.com/getgort/gort/client"
	"github.com/spf13/cobra"
)

const (
	schedulDeleteUse    = "delete"
	scheduleDeleteShort = "Deletes a scheduled command"
	scheduleDeleteLong  = "Deletes a scheduled command."
	scheduleDeleteUsage = `Usage:
gort schedule delete [flags] id

id is the id of the scheduled command to delete.

Flags:
  -h, --help  Show this message and exit
`
)

func GetScheduleDeleteCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   schedulDeleteUse,
		Short: scheduleDeleteShort,
		Long:  scheduleDeleteLong,
		RunE:  scheduleDeleteCommand,
		Args:  cobra.ExactArgs(1),
	}

	cmd.SetUsageTemplate(scheduleDeleteUsage)

	return cmd
}

func scheduleDeleteCommand(cmd *cobra.Command, args []string) error {
	c, err := client.Connect(FlagGortProfile, FlagConfigBaseDir)
	if err != nil {
		return err
	}

	id, err := strconv.ParseInt(args[0], 0, 64)
	if err != nil {
		return fmt.Errorf("invalid id: %w", err)
	}
	err = c.ScheduleDelete(id)
	if err != nil {
		return fmt.Errorf("failed to delete schedule %d: %w", id, err)
	}

	fmt.Printf("Successfully deleted schedule %d\n", id)

	return nil
}
