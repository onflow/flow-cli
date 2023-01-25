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

package services

import (
	"errors"
	"fmt"

	"github.com/onflow/flow-go-sdk"
	"github.com/onflow/flow-go-sdk/crypto"

	"github.com/onflow/flow-cli/pkg/flowkit"
	"github.com/onflow/flow-cli/pkg/flowkit/config"
	"github.com/onflow/flow-cli/pkg/flowkit/gateway"
	"github.com/onflow/flow-cli/pkg/flowkit/output"
	"github.com/onflow/flow-cli/pkg/flowkit/project"
)

// Project is a service that handles all interactions for a state.
type Project struct {
	gateway gateway.Gateway
	state   *flowkit.State
	logger  output.Logger
}

// NewProject returns a new state service.
func NewProject(
	gateway gateway.Gateway,
	state *flowkit.State,
	logger output.Logger,
) *Project {
	return &Project{
		gateway: gateway,
		state:   state,
		logger:  logger,
	}
}

// Init initializes a new project using the properties provided.
func (p *Project) Init(
	readerWriter flowkit.ReaderWriter,
	reset bool,
	global bool,
	sigAlgo crypto.SignatureAlgorithm,
	hashAlgo crypto.HashAlgorithm,
	serviceKey crypto.PrivateKey,
) (*flowkit.State, error) {
	path := config.DefaultPath
	if global {
		path = config.GlobalPath()
	}

	if flowkit.Exists(path) && !reset {
		return nil, fmt.Errorf(
			"configuration already exists at: %s, if you want to reset configuration use the reset flag",
			path,
		)
	}

	state, err := flowkit.Init(readerWriter, sigAlgo, hashAlgo)
	if err != nil {
		return nil, err
	}

	if serviceKey != nil {
		state.SetEmulatorKey(serviceKey)
	}

	err = state.Save(path)
	if err != nil {
		return nil, err
	}

	return state, nil
}

// Defines a Mainnet Standard Contract ( e.g Core Deployments, FungibleToken, NonFungibleToken )
type StandardContract struct {
	Name     string
	Address  flow.Address
	InfoLink string
}

func (p *Project) ReplaceStandardContractReferenceToAlias(standardContract StandardContract) error {
	//replace contract with alias
	c, err := p.state.Config().Contracts.ByNameAndNetwork(standardContract.Name, config.DefaultMainnetNetwork().Name)
	if err != nil {
		return err
	}
	c.Alias = standardContract.Address.String()

	//remove from deploy
	for di, d := range p.state.Config().Deployments {
		if d.Network != config.DefaultMainnetNetwork().Name {
			continue
		}
		for ci, c := range d.Contracts {
			if c.Name == standardContract.Name {
				p.state.Config().Deployments[di].Contracts = append((d.Contracts)[0:ci], (d.Contracts)[ci+1:]...)
				break
			}
		}
	}
	return nil
}

func (p *Project) CheckForStandardContractUsageOnMainnet() error {

	mainnetContracts := map[string]StandardContract{
		"FungibleToken": {
			Name:     "FungibleToken",
			Address:  flow.HexToAddress("0xf233dcee88fe0abe"),
			InfoLink: "https://developers.flow.com/flow/core-contracts/fungible-token",
		},
		"FlowToken": {
			Name:     "FlowToken",
			Address:  flow.HexToAddress("0x1654653399040a61"),
			InfoLink: "https://developers.flow.com/flow/core-contracts/flow-token",
		},
		"FlowFees": {
			Name:     "FlowFees",
			Address:  flow.HexToAddress("0xf919ee77447b7497"),
			InfoLink: "https://developers.flow.com/flow/core-contracts/flow-fees",
		},
		"FlowServiceAccount": {
			Name:     "FlowServiceAccount",
			Address:  flow.HexToAddress("0xe467b9dd11fa00df"),
			InfoLink: "https://developers.flow.com/flow/core-contracts/service-account",
		},
		"FlowStorageFees": {
			Name:     "FlowStorageFees",
			Address:  flow.HexToAddress("0xe467b9dd11fa00df"),
			InfoLink: "https://developers.flow.com/flow/core-contracts/service-account",
		},
		"FlowIDTableStaking": {
			Name:     "FlowIDTableStaking",
			Address:  flow.HexToAddress("0x8624b52f9ddcd04a"),
			InfoLink: "https://developers.flow.com/flow/core-contracts/staking-contract-reference",
		},
		"FlowEpoch": {
			Name:     "FlowEpoch",
			Address:  flow.HexToAddress("0x8624b52f9ddcd04a"),
			InfoLink: "https://developers.flow.com/flow/core-contracts/epoch-contract-reference",
		},
		"FlowClusterQC": {
			Name:     "FlowClusterQC",
			Address:  flow.HexToAddress("0x8624b52f9ddcd04a"),
			InfoLink: "https://developers.flow.com/flow/core-contracts/epoch-contract-reference",
		},
		"FlowDKG": {
			Name:     "FlowDKG",
			Address:  flow.HexToAddress("0x8624b52f9ddcd04a"),
			InfoLink: "https://developers.flow.com/flow/core-contracts/epoch-contract-reference",
		},
		"NonFungibleToken": {
			Name:     "NonFungibleToken",
			Address:  flow.HexToAddress("0x1d7e57aa55817448"),
			InfoLink: "https://developers.flow.com/flow/core-contracts/non-fungible-token",
		},
		"MetadataViews": {
			Name:     "MetadataViews",
			Address:  flow.HexToAddress("0x1d7e57aa55817448"),
			InfoLink: "https://developers.flow.com/flow/core-contracts/nft-metadata",
		},
	}

	contracts, err := p.state.DeploymentContractsByNetwork("mainnet")
	if err != nil {
		return err
	}

	for _, contract := range contracts {
		standardContract, ok := mainnetContracts[contract.Name]
		if !ok {
			continue
		}

		p.logger.Info(fmt.Sprintf("It seems like you are trying to deploy %s to Mainnet \n", contract.Name))
		p.logger.Info(fmt.Sprintf("It is a standard contract already deployed at address 0x%s \n", standardContract.Address.String()))
		p.logger.Info(fmt.Sprintf("You can read more about it here: %s \n", standardContract.InfoLink))

		if output.WantToUseMainnetVersionPrompt() {
			err := p.ReplaceStandardContractReferenceToAlias(standardContract)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// Deploy the project for the provided network.
//
// Retrieve all the contracts for specified network, sort them for deployment
// deploy one by one and replace the imports in the contract source so it corresponds
// to the account name the contract was deployed to.
func (p *Project) Deploy(network string, update bool) ([]*project.Contract, error) {
	if p.state == nil {
		return nil, config.ErrDoesNotExist
	}

	contracts, err := p.state.DeploymentContractsByNetwork(network)
	if err != nil {
		return nil, err
	}

	deployment, err := project.NewDeployment(contracts)
	if err != nil {
		return nil, err
	}

	sorted, err := deployment.Sort()
	if err != nil {
		return nil, err
	}

	p.logger.Info(fmt.Sprintf(
		"\nDeploying %d contracts for accounts: %s\n",
		len(sorted),
		p.state.AccountsForNetwork(network).String(),
	))
	defer p.logger.StopProgress()

	// todo refactor service layer so it can be shared
	accounts := NewAccounts(p.gateway, p.state, output.NewStdoutLogger(output.NoneLog))

	deployErr := &ProjectDeploymentError{}
	for _, contract := range sorted {
		targetAccount, err := p.state.Accounts().ByName(contract.AccountName)
		if err != nil {
			return nil, fmt.Errorf("target account for deploying contract not found in configuration")
		}

		// special case for emulator updates, where we remove and add a contract because it allows us to have more freedom in changes.
		// Updating contracts is limited as described in https://developers.flow.com/cadence/language/contract-updatability
		if update && network == config.DefaultEmulatorNetwork().Name {
			_, err = accounts.RemoveContract(targetAccount, contract.Name)
			if err != nil {
				deployErr.add(contract, err, fmt.Sprintf("failed to remove the contract %s before the update", contract.Name))
				continue
			}
		}

		txID, updated, err := accounts.AddContract(
			targetAccount,
			flowkit.NewScript(contract.Code(), contract.Args, contract.Location()),
			network,
			update,
		)
		if err != nil && errors.Is(err, errUpdateNoDiff) {
			p.logger.Info(fmt.Sprintf(
				"%s -> 0x%s [skipping, no changes found]",
				output.Italic(contract.Name),
				contract.AccountAddress.String(),
			))
			continue
		} else if err != nil {
			deployErr.add(contract, err, fmt.Sprintf("failed to deploy contract %s", contract.Name))
			continue
		}

		p.logger.Info(fmt.Sprintf(
			"%s -> 0x%s (%s) %s",
			output.Green(contract.Name),
			contract.AccountAddress,
			txID.String(),
			map[bool]string{true: "[updated]", false: ""}[updated],
		))
	}

	if len(deployErr.contracts) > 0 {
		return nil, deployErr
	}

	p.logger.Info(fmt.Sprintf("\n%s All contracts deployed successfully", output.SuccessEmoji()))
	return sorted, nil
}

type ProjectDeploymentError struct {
	contracts map[string]error
}

func (d *ProjectDeploymentError) add(contract *project.Contract, err error, msg string) {
	if d.contracts == nil {
		d.contracts = make(map[string]error)
	}
	d.contracts[contract.Name] = fmt.Errorf("%s: %w", msg, err)
}

func (d *ProjectDeploymentError) Contracts() map[string]error {
	return d.contracts
}

func (d *ProjectDeploymentError) Error() string {
	err := ""
	for c, e := range d.contracts {
		err = fmt.Sprintf("%s %s: %s,", err, c, e.Error())
	}
	return err
}
