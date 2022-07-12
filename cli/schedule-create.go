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
	"os"

	"github.com/getgort/gort/client"
	"github.com/getgort/gort/data/rest"

	"github.com/spf13/cobra"
)

const (
	scheduleCreateUse   = "create"
	scheduleCreateShort = "Schedule a new command"
	scheduleCreateLong  = "Schedule a new command."
	scheduleCreateUsage = `Usage:
gort schedule create [flags] cron command

cron is the specification of when the command should be run in cron format
command is a string containing the command to run.

Flags:
  -a, --adapter	string  The name of the adapter to schedule for.
                        Uses the GORT_ADAPTER env var if not provided.
  -c, --channel string  The id of the channel to schedule in.
                        Uses the GORT_ROOM env var if not provided.
  -h, --help            Show this message and exit
`
)

var (
	flagAdapter   string
	flagChannelID string
)

func GetScheduleCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   scheduleCreateUse,
		Short: scheduleCreateShort,
		Long:  scheduleCreateLong,
		RunE:  scheduleCreateCmd,
		Args:  cobra.ExactArgs(2),
	}

	cmd.Flags().StringVarP(&flagAdapter, "adapter", "a", os.Getenv("GORT_ADAPTER"), "The name of the adapter to schedule for.")
	cmd.Flags().StringVarP(&flagChannelID, "channel", "c", os.Getenv("GORT_ROOM"), "The id of the channel to schedule for.")

	cmd.SetUsageTemplate(scheduleCreateUsage)

	return cmd
}

func scheduleCreateCmd(cmd *cobra.Command, args []string) error {
	req := rest.ScheduleRequest{
		Command:   args[1],
		Cron:      args[0],
		Adapter:   flagAdapter,
		ChannelID: flagChannelID,
	}

	c, err := client.Connect(FlagGortProfile, FlagConfigBaseDir)
	if err != nil {
		return err
	}

	err = c.ScheduleCreate(req)
	if err != nil {
		return err
	}

	fmt.Println("Successfully scheduled!")

	return nil
}
