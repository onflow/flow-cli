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

type flagsViewDeployment struct {
}

var viewDeploymentFlags = flagsViewDeployment{}

var ViewDeploymentCommand = &command.Command{
	Cmd: &cobra.Command{
		Use:     "deployment",
		Short:   "View a list of networks deployed in configuration / View the properties of the deployed network",
		Example: "flow config view deployment \nflow config view deployment <networkname>",
		Args:    cobra.MaximumNArgs(1),
	},
	Flags: &viewDeploymentFlags,
	RunS:  viewDeployment,
}

func viewDeployment(args []string,
	_ flowkit.ReaderWriter,
	globalFlags command.GlobalFlags,
	_ *services.Services,
	state *flowkit.State,
) (command.Result, error) {

	// Flag for marking deployed network existence
	var flag int = 0
	// Count variable for printing format
	var count int = 0

	// IF CONDITION:
	// If there are zero arguments in the command i.e. command looks like --> flow config view deployment
	// Then we print the list of all the deployed networks in the configuration
	// ELSE IF CONDITION:
	// If there are arguments == 1 i.e. command looks like --> flow config view deployment <networkname>
	// Then we print all the details of the deployed network "<networkname>"
	// 	If the <networkname> isn't deployed in the configuration, then we print "Network <networkname> is not deployed"
	if len(args) == 0 {
		fmt.Print("List of Networks Deployed: ")
		for _, value := range *state.Deployments() {
			if count == 0 {
				fmt.Print(value.Network)
				count = count + 1
			} else if count > 0 {
				fmt.Print(", ", value.Network)
			}
		}
	} else if len(args) == 1 {
		for _, value := range state.Config().Deployments {
			if value.Network == args[0] {
				fmt.Print("Network Name: ", value.Network, "\n")
				fmt.Print("Account: ", value.Account, "\n")
				fmt.Print("Contracts with Cadence Value: ", value.Contracts)
				flag = 1
			}
		}
		if flag == 0 {
			fmt.Print("Network ", args[0], " is not deployed")
		}
	}

	return &Result{}, nil
}
