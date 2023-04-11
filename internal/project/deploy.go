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

	flowsdk "github.com/onflow/flow-go-sdk"
	"github.com/spf13/cobra"
	"golang.org/x/exp/slices"

	"github.com/onflow/flow-cli/internal/command"
	"github.com/onflow/flow-cli/internal/util"
	"github.com/onflow/flow-cli/pkg/flowkit"
	"github.com/onflow/flow-cli/pkg/flowkit/config"
	"github.com/onflow/flow-cli/pkg/flowkit/output"
	"github.com/onflow/flow-cli/pkg/flowkit/project"
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
		deployFunc = util.ShowContractDiffPrompt
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

	return &DeployResult{c}, nil
}

type DeployResult struct {
	contracts []*project.Contract
}

func (r *DeployResult) JSON() any {
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

// checkForStandardContractUsageOnMainnet checks if any contract defined to be used on mainnet
// are referencing standard contract and if so warn the use that they should use the already
// deployed contracts as an alias on mainnet instead of deploying their own copy.
func checkForStandardContractUsageOnMainnet(state *flowkit.State, logger output.Logger, replace bool) error {
	mainnetContracts := map[string]standardContract{
		"FungibleToken": {
			name:     "FungibleToken",
			address:  flowsdk.HexToAddress("0xf233dcee88fe0abe"),
			infoLink: "https://developers.flow.com/flow/core-contracts/fungible-token",
		},
		"FlowToken": {
			name:     "FlowToken",
			address:  flowsdk.HexToAddress("0x1654653399040a61"),
			infoLink: "https://developers.flow.com/flow/core-contracts/flow-token",
		},
		"FlowFees": {
			name:     "FlowFees",
			address:  flowsdk.HexToAddress("0xf919ee77447b7497"),
			infoLink: "https://developers.flow.com/flow/core-contracts/flow-fees",
		},
		"FlowServiceAccount": {
			name:     "FlowServiceAccount",
			address:  flowsdk.HexToAddress("0xe467b9dd11fa00df"),
			infoLink: "https://developers.flow.com/flow/core-contracts/service-account",
		},
		"FlowStorageFees": {
			name:     "FlowStorageFees",
			address:  flowsdk.HexToAddress("0xe467b9dd11fa00df"),
			infoLink: "https://developers.flow.com/flow/core-contracts/service-account",
		},
		"FlowIDTableStaking": {
			name:     "FlowIDTableStaking",
			address:  flowsdk.HexToAddress("0x8624b52f9ddcd04a"),
			infoLink: "https://developers.flow.com/flow/core-contracts/staking-contract-reference",
		},
		"FlowEpoch": {
			name:     "FlowEpoch",
			address:  flowsdk.HexToAddress("0x8624b52f9ddcd04a"),
			infoLink: "https://developers.flow.com/flow/core-contracts/epoch-contract-reference",
		},
		"FlowClusterQC": {
			name:     "FlowClusterQC",
			address:  flowsdk.HexToAddress("0x8624b52f9ddcd04a"),
			infoLink: "https://developers.flow.com/flow/core-contracts/epoch-contract-reference",
		},
		"FlowDKG": {
			name:     "FlowDKG",
			address:  flowsdk.HexToAddress("0x8624b52f9ddcd04a"),
			infoLink: "https://developers.flow.com/flow/core-contracts/epoch-contract-reference",
		},
		"NonFungibleToken": {
			name:     "NonFungibleToken",
			address:  flowsdk.HexToAddress("0x1d7e57aa55817448"),
			infoLink: "https://developers.flow.com/flow/core-contracts/non-fungible-token",
		},
		"MetadataViews": {
			name:     "MetadataViews",
			address:  flowsdk.HexToAddress("0x1d7e57aa55817448"),
			infoLink: "https://developers.flow.com/flow/core-contracts/nft-metadata",
		},
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
	contract := state.Config().Contracts.ByName(standardContract.name)
	if contract == nil {
		return fmt.Errorf("contract not found") // shouldn't occur
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
