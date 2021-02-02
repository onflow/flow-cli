/*
 * Flow CLI
 *
 * Copyright 2019-2020 Dapper Labs, Inc.
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

package initialize

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/onflow/flow-cli/flow/beta/cli"
)

var Cmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize a new Flow project",
	Run: func(cmd *cobra.Command, args []string) {
		if !cli.ProjectExists() {
			proj := cli.InitProject()
			proj.Save()

			fmt.Printf("Initialized a new Flow project in \"%s\"!\n\n", cli.DefaultConfigPath)
			fmt.Printf("Start the emulator by running: flow beta start-emulator\n")
		} else {
			fmt.Printf("Flow configuration file already exists! Begin by running: flow beta start-emulator\n")
		}
	},
}
