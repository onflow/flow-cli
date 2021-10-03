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

type flagsViewAccount struct {
}

var viewAccountFlags = flagsViewAccount{}

var ViewAccountCommand = &command.Command{
	Cmd: &cobra.Command{
		Use:     "account",
		Short:   "View a list of accounts in the configuration / View the properties of a particular account",
		Example: "flow config view account \nflow config view account <accountname>",
		Args:    cobra.MaximumNArgs(1),
	},
	Flags: &viewAccountFlags,
	RunS:  viewAccount,
}

func viewAccount(args []string,
	_ flowkit.ReaderWriter,
	globalFlags command.GlobalFlags,
	_ *services.Services,
	state *flowkit.State,
) (command.Result, error) {

	// Flag for marking account existence.
	var flag int = 0
	// Count variable for printing format.
	var count int = 0

	// IF CONDITION:
	// If there are zero arguments in the command i.e. command looks like --> flow config view account,
	// Then we print the list of all the present accounts in the configuration.
	// ELSE IF CONDITION:
	// If there are arguments == 1 i.e. command looks like --> flow config view account <accountname>,
	// Then we print all the details of the account "<accountname>".
	// 	If the <accountname> doesn't exist in the configuration, then we print "Account <accountname> does not exist".
	if len(args) == 0 {
		fmt.Print("List of Accounts: ")
		for _, value := range *state.Accounts() {
			if count == 0 {
				fmt.Print(value.Name())
				count = count + 1
			} else if count > 0 {
				fmt.Print(", ", value.Name())
			}
		}
	} else if len(args) == 1 {
		for _, value := range state.Config().Accounts {
			if value.Name == args[0] {
				fmt.Print("Account Name: ", value.Name, "\n")
				fmt.Print("Address: ", value.Address, "\n")
				fmt.Print("Key Properties: ", "\n")
				fmt.Print("Type: ", value.Key.Type, ", ")
				fmt.Print("Index: ", value.Key.Index, ", ")
				fmt.Print("Signature Algorithm: ", value.Key.SigAlgo, ", ")
				fmt.Print("Hash Algorithm: ", value.Key.HashAlgo, ", ")
				fmt.Print("Private Key: ", value.Key.PrivateKey)
				flag = 1
			}
		}
		if flag == 0 {
			fmt.Print("Account ", args[0], " does not exist")
		}
	}

	return &Result{}, nil
}
