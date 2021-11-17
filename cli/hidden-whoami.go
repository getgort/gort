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

	"github.com/spf13/cobra"
)

const (
	hiddenWhoamiUse   = "whoami"
	hiddenWhoamiShort = "Provides your basic identity and account information"
	hiddenWhoamiLong  = `Provides your basic identity and account information.`
	hiddenWhoamiUsage = `Usage:
	!gort:whoami

  Flags:
	-h, --help   Show this message and exit
  `
)

// GetHiddenWhoamiCmd is a command
func GetHiddenWhoamiCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:          hiddenWhoamiUse,
		Short:        hiddenWhoamiShort,
		Long:         hiddenWhoamiLong,
		RunE:         hiddenWhoamiCmd,
		SilenceUsage: true,
	}

	cmd.SetUsageTemplate(hiddenWhoamiUsage)

	return cmd
}

func hiddenWhoamiCmd(cmd *cobra.Command, args []string) error {
	if _, ok := os.LookupEnv("GORT_SERVICE_TOKEN"); !ok {
		return fmt.Errorf("whoami can only be run from chat")
	}

	var adapter, chatUserID, gortUser string

	if adapter = os.Getenv("GORT_ADAPTER"); adapter == "" {
		adapter = "*UNDEFINED!*"
	}

	if chatUserID = os.Getenv("GORT_CHAT_ID"); chatUserID == "" {
		chatUserID = "*UNDEFINED!*"
	}

	if gortUser = os.Getenv("GORT_USER"); gortUser == "" {
		gortUser = "*UNDEFINED!*"
	} else if gortUser == "nobody" {
		gortUser = "Nobody"
	}

	tmpl := `Adapter:   %s
User ID:   %s
Mapped to: %s
`

	fmt.Printf(tmpl, adapter, chatUserID, gortUser)
	return nil
}
