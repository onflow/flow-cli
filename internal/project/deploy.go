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
	"context"
	"errors"
	"fmt"
	"github.com/onflow/flow-go/fvm/systemcontracts"
	flowGo "github.com/onflow/flow-go/model/flow"

	flowsdk "github.com/onflow/flow-go-sdk"
	"github.com/spf13/cobra"
	"golang.org/x/exp/slices"

	"github.com/onflow/flowkit/v2"
	"github.com/onflow/flowkit/v2/config"
	"github.com/onflow/flowkit/v2/output"
	"github.com/onflow/flowkit/v2/project"

	"github.com/onflow/flow-cli/internal/command"
	"github.com/onflow/flow-cli/internal/util"
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

func deploy(
	_ []string,
	global command.GlobalFlags,
	logger output.Logger,
	flow flowkit.Services,
	state *flowkit.State,
) (command.Result, error) {

	if flow.Network() == config.MainnetNetwork { // if using mainnet check for standard contract usage
		err := checkForStandardContractUsageOnMainnet(state, logger, global.Yes)
		if err != nil {
			return nil, err
		}
	}

	deployFunc := flowkit.UpdateExistingContract(deployFlags.Update)
	if deployFlags.ShowDiff {
		deployFunc = util.ShowContractDiffPrompt(logger)
	}

	c, err := flow.DeployProject(context.Background(), deployFunc)
	if err != nil {
		var projectErr *flowkit.ProjectDeploymentError
		if errors.As(err, &projectErr) {
			for name, err := range projectErr.Contracts() {
				logger.Info(fmt.Sprintf(
					"%s Failed to deploy contract %s: %s",
					output.ErrorEmoji(),
					name,
					err.Error(),
				))
			}
			return nil, fmt.Errorf("failed deploying all contracts")
		}
		return nil, err
	}

	return &deployResult{c}, nil
}

type deployResult struct {
	contracts []*project.Contract
}

func (r *deployResult) JSON() any {
	result := make(map[string]any)

	for _, contract := range r.contracts {
		result[contract.Name] = contract.AccountAddress.String()
	}

	return result
}

func (r *deployResult) String() string {
	return ""
}

func (r *deployResult) Oneliner() string {
	return ""
}

// checkForStandardContractUsageOnMainnet checks if any contract defined to be used on mainnet
// are referencing standard contract and if so warn the use that they should use the already
// deployed contracts as an alias on mainnet instead of deploying their own copy.
func checkForStandardContractUsageOnMainnet(state *flowkit.State, logger output.Logger, replace bool) error {

	mainnetContracts := make(map[string]standardContract)
	sc := systemcontracts.SystemContractsForChain(flowGo.Mainnet)

	for _, coreContract := range sc.All() {
		mainnetContracts[coreContract.Name] = standardContract{
			name:     coreContract.Name,
			address:  flowsdk.HexToAddress(coreContract.Address.String()),
			infoLink: "https://developers.flow.com/flow/core-contracts/",
		}
	}

	contracts, err := state.DeploymentContractsByNetwork(config.MainnetNetwork)
	if err != nil {
		return err
	}

	for _, contract := range contracts {
		standardContract, ok := mainnetContracts[contract.Name]
		if !ok {
			continue
		}

		logger.Info(fmt.Sprintf("It seems like you are trying to deploy %s to Mainnet \n", contract.Name))
		logger.Info(fmt.Sprintf("It is a standard contract already deployed at address 0x%s \n", standardContract.address.String()))
		logger.Info(fmt.Sprintf("You can read more about it here: %s \n", standardContract.infoLink))

		if replace || util.WantToUseMainnetVersionPrompt() {
			err := replaceContractWithAlias(state, standardContract)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

type standardContract struct {
	name     string
	address  flowsdk.Address
	infoLink string
}

func replaceContractWithAlias(state *flowkit.State, standardContract standardContract) error {
	contract, err := state.Config().Contracts.ByName(standardContract.name)
	if err != nil {
		return err
	}
	contract.Aliases.Add(config.MainnetNetwork.Name, standardContract.address) // replace contract with an alias

	for di, d := range state.Config().Deployments.ByNetwork(config.MainnetNetwork.Name) {
		for ci, c := range d.Contracts {
			if c.Name == standardContract.name {
				state.Config().Deployments[di].Contracts = slices.Delete(state.Config().Deployments[di].Contracts, ci, ci+1)
				if len(state.Config().Deployments[di].Contracts) == 0 {
					_ = state.Config().Deployments.Remove(d.Account, d.Network)
				}
				break
			}
		}
	}
	return nil
}
