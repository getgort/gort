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
	"os"

	"github.com/getgort/gort/client"
	"github.com/getgort/gort/service"
	"github.com/spf13/cobra"
)

const (
	scheduleCreateUse   = "create"
	scheduleCreateShort = "Schedule a new command"
	scheduleCreateLong  = "Schedule a new command."
	scheduleCreateUsage = `Usage:
gort schedule create [flags] cron command
`
)

func GetScheduleCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   scheduleCreateUse,
		Short: scheduleCreateShort,
		Long:  scheduleCreateLong,
		RunE:  scheduleCreateCmd,
		Args:  cobra.ExactArgs(2),
	}

	cmd.SetUsageTemplate(scheduleCreateUsage)

	return cmd
}

func scheduleCreateCmd(cmd *cobra.Command, args []string) error {
	req := service.ScheduleRequest{
		Command:   args[1],
		Cron:      args[0],
		Adapter:   os.Getenv("GORT_ADAPTER"),
		ChannelID: os.Getenv("GORT_ROOM"),
	}

	c, err := client.Connect(FlagGortProfile, FlagConfigBaseDir)
	if err != nil {
		return err
	}

	err = c.ScheduleCreate(req)
	if err != nil {
		return err
	}

	return nil
}
