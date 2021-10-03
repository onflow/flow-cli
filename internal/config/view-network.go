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

type flagsViewNetwork struct {
}

var viewNetworkFlags = flagsViewNetwork{}

var ViewNetworkCommand = &command.Command{
	Cmd: &cobra.Command{
		Use:     "network",
		Short:   "View a list of networks in configuration / View the properties of a paticular network",
		Example: "flow config view network \nflow config view network <networkname>",
		Args:    cobra.MaximumNArgs(1),
	},
	Flags: &viewNetworkFlags,
	RunS:  viewNetwork,
}

func viewNetwork(args []string,
	_ flowkit.ReaderWriter,
	globalFlags command.GlobalFlags,
	_ *services.Services,
	state *flowkit.State,
) (command.Result, error) {

	// Flag for marking network existence
	var flag int = 0
	// Count variable for printing format
	var count int = 0

	// IF CONDITION:
	// If there are zero arguments in the command i.e. command looks like --> flow config view network
	// Then we print the list of all the present networks in the configuration
	// ELSE IF CONDITION:
	// If there are arguments == 1 i.e. command looks like --> flow config view network <networkname>
	// Then we print all the details of the network "<networkname>"
	// 	If the <networkname> doesn't exist in the configuration, then we print "Network <networkname> does not exist"
	if len(args) == 0 {
		fmt.Print("List of Networks: ")
		for _, value := range *state.Networks() {
			if count == 0 {
				fmt.Print(value.Name)
				count = count + 1
			} else if count > 0 {
				fmt.Print(", ", value.Name)
			}
		}
	} else if len(args) == 1 {
		for _, value := range state.Config().Networks {
			if value.Name == args[0] {
				fmt.Print("Network Name: ", value.Name, "\n")
				fmt.Print("Host: ", value.Host)
				flag = 1
			}
		}
		if flag == 0 {
			fmt.Print("Network ", args[0], " does not exist")
		}
	}

	return &Result{}, nil
}
