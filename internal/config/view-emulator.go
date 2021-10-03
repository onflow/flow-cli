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

type flagsViewEmulator struct {
}

var viewEmulatorFlags = flagsViewEmulator{}

var ViewEmulatorCommand = &command.Command{
	Cmd: &cobra.Command{
		Use:     "emulator",
		Short:   "View a list of emulators in configuration",
		Example: "flow config view emulator",
		Args:    cobra.MaximumNArgs(1),
	},
	Flags: &viewEmulatorFlags,
	RunS:  viewEmulator,
}

func viewEmulator(args []string,
	_ flowkit.ReaderWriter,
	globalFlags command.GlobalFlags,
	_ *services.Services,
	state *flowkit.State,
) (command.Result, error) {

	// Flag for marking account existence
	var flag int = 0
	// Count variable for printing format
	var count int = 0

	// IF CONDITION:
	// If there are zero arguments in the command i.e. command looks like --> flow config view account
	// Then we print the list of all the present accounts in the configuration
	// ELSE IF CONDITION:
	// If there are arguments == 1 i.e. command looks like --> flow config view account <accountname>
	// Then we print all the details of the account "<accountname>"
	// 	If the <accountname> does'nt exist in the configuration, then we print "Account does not exist"
	if len(args) == 0 {
		fmt.Print("List of Emulators: ")
		for _, value := range state.Config().Emulators {
			if count == 0 {
				fmt.Print(value.Name)
				count = count + 1
			} else if count > 0 {
				fmt.Print(", ", value.Name)
			}
		}
	} else if len(args) == 1 {
		for _, value := range state.Config().Emulators {
			if value.Name == args[0] {
				fmt.Print("Emulator Name: ", value.Name, "\n")
				fmt.Print("Port: ", value.Port, "\n")
				fmt.Print("Service Account: ", value.ServiceAccount, "\n")
				flag = 1
			}
		}
		if flag == 0 {
			fmt.Print("Emulator does not exist")
		}
	}

	return &Result{}, nil
}
