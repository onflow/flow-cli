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
	"fmt"
	"strings"

	"github.com/onflow/flow-cli/pkg/flowkit"

	"github.com/onflow/flow-cli/pkg/flowkit/config"

	"github.com/onflow/flow-go-sdk/crypto"

	"github.com/onflow/flow-cli/pkg/flowkit/contracts"
	"github.com/onflow/flow-cli/pkg/flowkit/gateway"
	"github.com/onflow/flow-cli/pkg/flowkit/output"
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

// Deploy the project for the provided network.
//
// Retrieve all the contracts for specified network, sort them for deployment
// deploy one by one and replace the imports in the contract source so it corresponds
// to the account name the contract was deployed to.
func (p *Project) Deploy(network string, update bool) ([]*contracts.Contract, error) {
	if p.state == nil {
		return nil, config.ErrDoesNotExist
	}

	// check there are not multiple accounts with same contract
	if p.state.ContractConflictExists(network) {
		return nil, fmt.Errorf( // TODO(sideninja) specify which contract by name is a problem
			"the same contract cannot be deployed to multiple accounts on the same network",
		)
	}

	// create new processor for contract
	processor := contracts.NewPreprocessor(
		contracts.FilesystemLoader{
			Reader: p.state.ReaderWriter(),
		},
		p.state.AliasesForNetwork(network),
	)

	// add all contracts needed to deploy to processor
	contractsNetwork, err := p.state.DeploymentContractsByNetwork(network)
	if err != nil {
		return nil, err
	}

	for _, contract := range contractsNetwork {
		err := processor.AddContractSource(
			contract.Name,
			contract.Source,
			contract.AccountAddress,
			contract.AccountName,
			contract.Args,
		)
		if err != nil {
			return nil, err
		}
	}

	// resolve imports assigns accounts to imports
	err = processor.ResolveImports()
	if err != nil {
		return nil, err
	}

	// sort correct deployment order of contracts so we don't have import that is not yet deployed
	orderedContracts, err := processor.ContractDeploymentOrder()
	if err != nil {
		return nil, err
	}

	p.logger.Info(fmt.Sprintf(
		"\nDeploying %d contracts for accounts: %s\n",
		len(orderedContracts),
		strings.Join(p.state.AccountNamesForNetwork(network), ","),
	))
	defer p.logger.StopProgress()

	deployErr := false
	for _, contract := range orderedContracts {
		block, err := p.gateway.GetLatestBlock()
		if err != nil {
			return nil, err
		}

		targetAccount, err := p.state.Accounts().ByName(contract.AccountName())

		if err != nil {
			return nil, fmt.Errorf("target account for deploying contract not found in configuration")
		}

		// get deployment account
		targetAccountInfo, err := p.gateway.GetAccount(targetAccount.Address())
		if err != nil {
			return nil, fmt.Errorf("failed to fetch information for account %s with error %s", targetAccount.Address(), err.Error())
		}

		// create transaction to deploy new contract with args
		tx, err := flowkit.NewAddAccountContractTransaction(
			targetAccount,
			contract.Name(),
			contract.TranspiledCode(),
			contract.Args(),
		)
		if err != nil {
			return nil, err
		}
		// check if contract exists on account
		_, exists := targetAccountInfo.Contracts[contract.Name()]
		if exists && !update {
			p.logger.Error(fmt.Sprintf(
				"contract %s is already deployed to this account. Use the --update flag to force update",
				contract.Name(),
			))
			deployErr = true
			continue
		} else if exists && len(contract.Args()) > 0 { // TODO(sideninja) discuss removing the contract and redeploying it
			p.logger.Error(fmt.Sprintf(
				"contract %s is already deployed and can not be updated with initialization arguments",
				contract.Name(),
			))
			deployErr = true
			continue
		} else if exists {
			tx, err = flowkit.NewUpdateAccountContractTransaction(targetAccount, contract.Name(), contract.TranspiledCode())
			if err != nil {
				return nil, err
			}
		}

		tx.SetBlockReference(block)

		if err = tx.SetProposer(targetAccountInfo, targetAccount.Key().Index()); err != nil {
			return nil, err
		}

		tx, err = tx.Sign()
		if err != nil {
			p.logger.Error(fmt.Sprintf("%s error: %s", contract.Name(), err))
			deployErr = true
			continue
		}

		p.logger.StartProgress(
			fmt.Sprintf("%s deploying...", output.Bold(contract.Name())),
		)

		sentTx, err := p.gateway.SendSignedTransaction(tx)
		if err != nil {
			p.logger.StopProgress()
			p.logger.Error(fmt.Sprintf("%s error: %s", contract.Name(), err))
			deployErr = true
			continue
		}

		result, err := p.gateway.GetTransactionResult(sentTx, true)
		if err != nil {
			p.logger.StopProgress()
			p.logger.Error(fmt.Sprintf("%s error: %s", contract.Name(), err))
			deployErr = true
			continue
		}
		if result == nil {
			p.logger.Error("could not fetch the result of deployment, skipping")
			deployErr = true
			continue
		}
		if result.Error != nil {
			deployErr = true
			p.logger.StopProgress()
			p.logger.Error(fmt.Sprintf(
				"Error deploying %s: (%s)\n",
				output.Red(contract.Name()),
				result.Error.Error(),
			))
		}

		if result.Error == nil && !deployErr {
			p.logger.StopProgress()
			p.logger.Info(fmt.Sprintf(
				"%s -> 0x%s (%s)\n",
				output.Green(contract.Name()),
				contract.Target(),
				sentTx.ID().String(),
			))
		}
	}

	if !deployErr {
		p.logger.Info(fmt.Sprintf("\n%s All contracts deployed successfully", output.SuccessEmoji()))
	} else {
		err = fmt.Errorf("failed to deploy all contracts")
		p.logger.Error(err.Error())
		return nil, err
	}

	return orderedContracts, nil
}
