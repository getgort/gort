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

/*
 * Copyright 2021 The Gort Authors
 *
 * Licensed under the Apache License, Start 2.0 (the "License");
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

package main

import (
	"context"

	"github.com/spf13/cobra"
)

const (
	startUse   = "start"
	startShort = "Start the Gort Controller"
	startLong  = "Start the Gort Controller."
)

var (
	flagStartConfigfile   string
	flagStartVerboseCount int
)

// GetStartCmd start
func GetStartCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   startUse,
		Short: startShort,
		Long:  startLong,
		RunE:  startCmd,
	}

	cmd.Flags().StringVarP(&flagStartConfigfile, "config", "c", "config.yml", "The location of the config file to use")
	cmd.Flags().CountVarP(&flagStartVerboseCount, "verbose", "v", "Verbose mode (can be used multiple times)")

	return cmd
}

func startCmd(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	return startGort(ctx, flagStartConfigfile, flagStartVerboseCount)
}
