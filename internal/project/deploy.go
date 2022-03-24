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

package project

import (
	"github.com/spf13/cobra"

	"github.com/onflow/flow-cli/internal/command"
	"github.com/onflow/flow-cli/pkg/flowkit"
	"github.com/onflow/flow-cli/pkg/flowkit/contracts"
	"github.com/onflow/flow-cli/pkg/flowkit/services"
)

type flagsDeploy struct {
	Update bool `flag:"update" default:"false" info:"use update flag to update existing contracts"`
}

var deployFlags = flagsDeploy{}

var DeployCommand = &command.Command{
	Cmd: &cobra.Command{
		Use:     "deploy",
		Short:   "Deploy Cadence contracts",
		Example: "flow project deploy --network testnet",
	},
	Flags: &deployFlags,
	RunS:  deploy,
}

func deploy(
	_ []string,
	_ flowkit.ReaderWriter,
	globalFlags command.GlobalFlags,
	services *services.Services,
	_ *flowkit.State,
) (command.Result, error) {
	c, err := services.Project.Deploy(globalFlags.Network, deployFlags.Update)
	if err != nil {
		return nil, err
	}

	return &DeployResult{c}, nil
}

type DeployResult struct {
	contracts []*contracts.Contract
}

func (r *DeployResult) JSON() interface{} {
	result := make(map[string]string)

	for _, contract := range r.contracts {
		result[contract.Name()] = contract.Target().String()
	}

	return result
}

func (r *DeployResult) String() string {
	return ""
}

func (r *DeployResult) Oneliner() string {
	return ""
}
