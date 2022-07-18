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
	"strings"

	"github.com/getgort/gort/client"

	"github.com/spf13/cobra"
)

const (
	scheduleGetUse   = "get"
	scheduleGetShort = "Get a list of schedules commands"
	scheduleGetLong  = "Get a list of schedules commands."
	scheduleGetUsage = `Usage:
gort schedule get [flags]

Flags:
  -h, --help  Show this message and exit`
)

func GetScheduleGetCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   scheduleGetUse,
		Short: scheduleGetShort,
		Long:  scheduleGetLong,
		RunE:  scheduleGetCmd,
	}

	cmd.SetUsageTemplate(scheduleGetUsage)

	return cmd
}

func scheduleGetCmd(*cobra.Command, []string) error {
	c, err := client.Connect(FlagGortProfile, FlagConfigBaseDir)
	if err != nil {
		return err
	}

	info, err := c.SchedulesGet()
	if err != nil {
		return err
	}

	maxId, maxAdapter, maxChannel, maxCron := 2, 7, 7, 4

	for _, s := range info {
		if len(fmt.Sprintf("%d", s.Id)) > maxId {
			maxId = len(fmt.Sprintf("%d", s.Id))
		}
		if len(s.Adapter) > maxAdapter {
			maxAdapter = len(s.Adapter)
		}
		if len(s.ChannelName) > maxChannel {
			maxChannel = len(s.ChannelName)
		}
		if len(s.Cron) > maxCron {
			maxCron = len(s.Cron)
		}
	}

	template := fmt.Sprintf(" %%%dd | %%%ds | %%%ds | %%%ds | %%s\n", maxId, maxAdapter, maxChannel, maxCron)

	header := fmt.Sprintf(fmt.Sprintf(" %%%ds | %%%ds | %%%ds | %%%ds | %%s\n", maxId, maxAdapter, maxChannel, maxCron), "ID", "Adapter", "Channel", "Cron", "Command")
	fmt.Print(header)
	fmt.Println(strings.Repeat("-", len(header)))
	for _, s := range info {
		fmt.Printf(template, s.Id, s.Adapter, s.ChannelName, s.Cron, s.Command)
	}

	return nil
}
