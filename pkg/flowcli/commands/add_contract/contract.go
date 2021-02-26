/*
 * Flow CLI
 *
 * Copyright 2019-2021 Dapper Labs, Inc.
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

package add_contract

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/onflow/flow-cli/flow/project/cli"
	"github.com/onflow/flow-cli/flow/project/cli/config"
)

var Cmd = &cobra.Command{
	Use:   "add-contract <name> <path>",
	Short: "Add a contract to the flow.json file",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		project := cli.LoadProject()

		name := args[0]
		path := args[1]

		if name == "" {
			fmt.Println("\n❌ Contract name is required")
			cli.Exit(1, "")
		}
		if project.ContractExist(name) {
			fmt.Printf("\n❌ Contract with name %s already exist\n", name)
			cli.Exit(1, "")
		}

		_, err := os.Open(path)
		if err != nil {
			if os.IsNotExist(err) {
				fmt.Printf("\n❌ Path %s does not exist\n", path)
			}
			os.Exit(1)
		}

		contract := config.Contract{
			Name:   name,
			Source: path,
		}

		project.AddContract(contract)

		project.Save()
	},
}
