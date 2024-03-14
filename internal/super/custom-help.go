/*
 * Flow CLI
 *
 * Copyright 2019 Dapper Labs, Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *   http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package super

import (
	"github.com/spf13/cobra"
)

type flagsHelp struct {
	All bool
}

var helpFlags = flagsHelp{}

var CustomHelp = &cobra.Command{
	Use:     "help [command]",
	Short:   "Help about any command",
	Example: "flow help [command] --all",
	Run: func(cmd *cobra.Command, args []string) {
		if helpFlags.All {
			AllHelp.Run(cmd, args)
			return
		}
		// Fallback to default help behavior
		cmd.Help()
	},
}

func init() {
	// Register the --all flag for the custom help command
	CustomHelp.Flags().BoolVarP(&helpFlags.All, "all", "a", false, "Show help for all commands")
}
