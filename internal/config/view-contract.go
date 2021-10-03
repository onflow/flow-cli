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

package config

import (
	"fmt"

	"github.com/onflow/flow-cli/internal/command"
	"github.com/spf13/cobra"

	"github.com/onflow/flow-cli/pkg/flowkit"
	"github.com/onflow/flow-cli/pkg/flowkit/services"
)

type flagsViewContract struct {
}

var viewContractFlags = flagsViewContract{}

var ViewContractCommand = &command.Command{
	Cmd: &cobra.Command{
		Use:     "contract",
		Short:   "View a list of contracts in the configuration / View the properties of a particular contract",
		Example: "flow config view contract \nflow config view contract <contractname>",
		Args:    cobra.MaximumNArgs(1),
	},
	Flags: &viewContractFlags,
	RunS:  viewContract,
}

func viewContract(args []string,
	_ flowkit.ReaderWriter,
	globalFlags command.GlobalFlags,
	_ *services.Services,
	state *flowkit.State,
) (command.Result, error) {

	// Flag for marking contract existence.
	var flag int = 0
	// Count variable for printing format.
	var count int = 0

	// IF CONDITION:
	// If there are zero arguments in the command i.e. command looks like --> flow config view contract,
	// Then we print the list of all the present contracts in the configuration.
	// ELSE IF CONDITION:
	// If there are arguments == 1 i.e. command looks like --> flow config view contract <contractname>,
	// Then we print all the details of the contract "<contractname>".
	// 	If the <contractname> doesn't exist in the configuration, then we print "Contract <contractname> does not exist".
	if len(args) == 0 {
		fmt.Print("List of Contracts: ")
		for _, value := range *state.Contracts() {
			if count == 0 {
				fmt.Print(value.Name)
				count = count + 1
			} else if count > 0 {
				fmt.Print(", ", value.Name)
			}
		}
	} else if len(args) == 1 {
		for _, value := range state.Config().Contracts {
			if value.Name == args[0] {
				fmt.Print("Contract Name: ", value.Name, "\n")
				fmt.Print("Source: ", value.Source, "\n")
				fmt.Print("Network: ", value.Network, "\n")
				fmt.Print("Alias: ", value.Alias)
				flag = 1
			}
		}
		if flag == 0 {
			fmt.Print("Contract ", args[0], " does not exist")
		}
	}

	return &Result{}, nil
}
