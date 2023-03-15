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
	"errors"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/onflow/flow-cli/internal/command"
	"github.com/onflow/flow-cli/pkg/flowkit"
	"github.com/onflow/flow-cli/pkg/flowkit/config"
	"github.com/onflow/flow-cli/pkg/flowkit/output"
	"github.com/onflow/flow-cli/pkg/flowkit/project"
	"github.com/onflow/flow-cli/pkg/flowkit/services"
)

type flagsDeploy struct {
	Update   bool `flag:"update" default:"false" info:"use update flag to update existing contracts"`
	ShowDiff bool `flag:"show-diff" default:"false" info:"use show-diff flag to show diff between existing and new contracts on update"`
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

func updateWithPrompt(existing []byte, new []byte) bool {
	return output.ShowContractDiffPrompt(existing, new)
}

func deploy(
	_ []string,
	_ flowkit.ReaderWriter,
	globalFlags command.GlobalFlags,
	srv *services.Services,
	_ *flowkit.State,
) (command.Result, error) {

	//precheck for standard contract on Mainnet
	if globalFlags.Network == config.DefaultMainnetNetwork().Name {
		err := srv.Project.CheckForStandardContractUsageOnMainnet()
		if err != nil {
			return nil, err
		}

	}

	deployFunc := services.UpdateExisting(deployFlags.Update)
	if deployFlags.ShowDiff {
		deployFunc = updateWithPrompt
	}

	c, err := srv.Project.Deploy(globalFlags.Network, deployFunc)
	if err != nil {
		var projectErr *services.ProjectDeploymentError
		if errors.As(err, &projectErr) {
			for name, err := range projectErr.Contracts() {
				fmt.Printf(
					"%s Failed to deploy contract %s: %s\n",
					output.ErrorEmoji(),
					name,
					err.Error(),
				)
			}
			return nil, fmt.Errorf("failed deploying all contracts")
		}
		return nil, err
	}

	return &DeployResult{c}, nil
}

type DeployResult struct {
	contracts []*project.Contract
}

func (r *DeployResult) JSON() interface{} {
	result := make(map[string]string)

	for _, contract := range r.contracts {
		result[contract.Name] = contract.AccountAddress.String()
	}

	return result
}

func (r *DeployResult) String() string {
	return ""
}

func (r *DeployResult) Oneliner() string {
	return ""
}
