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

package services

import (
	"fmt"

	"strings"

	"github.com/onflow/flow-cli/pkg/flow"
	"github.com/onflow/flow-cli/pkg/flow/util"

	"github.com/onflow/flow-go-sdk/crypto"

	"github.com/onflow/flow-go-sdk/templates"

	"github.com/onflow/flow-cli/pkg/flow/contracts"
	"github.com/onflow/flow-cli/pkg/flow/gateway"
	flowsdk "github.com/onflow/flow-go-sdk"
)

// Project service handles all interactions for project
type Project struct {
	gateway gateway.Gateway
	project *flow.Project
	logger  util.Logger
}

// NewProject create new project service
func NewProject(
	gateway gateway.Gateway,
	project *flow.Project,
	logger util.Logger,
) *Project {
	return &Project{
		gateway: gateway,
		project: project,
		logger:  logger,
	}
}

func (p *Project) Init(reset bool, serviceKeySigAlgo string, serviceKeyHashAlgo string, servicePrivateKey string) (*flow.Project, error) {
	if !flow.ProjectExists(flow.DefaultConfigPath) || reset {

		sigAlgo, hashAlgo, err := util.ConvertSigAndHashAlgo(serviceKeySigAlgo, serviceKeyHashAlgo)
		if err != nil {
			return nil, err
		}

		project, err := flow.InitProject(sigAlgo, hashAlgo)
		if err != nil {
			return nil, err
		}

		if len(servicePrivateKey) > 0 {
			serviceKey, err := crypto.DecodePrivateKeyHex(sigAlgo, servicePrivateKey)
			if err != nil {
				return nil, fmt.Errorf("could not decode private key for a service account, provided private key: %s", servicePrivateKey)
			}

			project.SetEmulatorServiceKey(serviceKey)
		}

		err = project.Save(flow.DefaultConfigPath)
		if err != nil {
			return nil, err
		}

		return project, nil
	} else {
		return nil, fmt.Errorf("configuration already exists at: %s, if you want to reset configuration use the reset flag", flow.DefaultConfigPath)
	}
}

func (p *Project) Deploy(network string, update bool) ([]*contracts.Contract, error) {
	// check there are not multiple accounts with same contract
	// TODO: specify which contract by name is a problem
	if p.project.ContractConflictExists(network) {
		return nil, fmt.Errorf("currently it is not possible to deploy same contract with multiple accounts, please check Deployments in config and make sure a contract is only present in one account")
	}

	processor := contracts.NewPreprocessor(
		contracts.FilesystemLoader{},
		p.project.GetAliases(network),
	)

	for _, contract := range p.project.GetContractsByNetwork(network) {
		err := processor.AddContractSource(
			contract.Name,
			contract.Source,
			contract.Target,
		)
		if err != nil {
			return nil, err
		}
	}

	err := processor.ResolveImports()
	if err != nil {
		return nil, err
	}

	contracts, err := processor.ContractDeploymentOrder()
	if err != nil {
		return nil, err
	}

	p.logger.Info(fmt.Sprintf(
		"Deploying %v contracts for accounts: %s",
		len(contracts),
		strings.Join(p.project.GetAllAccountNames(), ","),
	))

	var errs []error
	for _, contract := range contracts {
		targetAccount := p.project.GetAccountByAddress(contract.Target().String())

		targetAccountInfo, err := p.gateway.GetAccount(targetAccount.Address())
		if err != nil {
			return nil, fmt.Errorf("failed to fetch information for account %s", targetAccount.Address())
		}

		var tx *flowsdk.Transaction

		_, exists := targetAccountInfo.Contracts[contract.Name()]
		if exists {
			if !update {
				p.logger.Error(fmt.Sprintf(
					"Contract %s is already deployed to account, use --update flag to force update.", contract.Name(),
				))
				continue
			}

			tx = prepareUpdateContractTransaction(targetAccount.Address(), contract)
		} else {
			tx = prepareAddContractTransaction(targetAccount.Address(), contract)
		}

		tx, err = p.gateway.SendTransaction(tx, targetAccount)

		p.logger.StartProgress(
			fmt.Sprintf("%s deploying...", util.Bold(contract.Name())),
		)

		result, err := p.gateway.GetTransactionResult(tx, true)

		p.logger.StopProgress("")

		if result.Error == nil {
			p.logger.Info(
				fmt.Sprintf("%s -> 0x%s", util.Green(contract.Name()), contract.Target()),
			)
		} else {
			p.logger.Error(
				fmt.Sprintf("%s error", contract.Name()),
			)

			errs = append(errs, result.Error)
		}
	}

	if len(errs) == 0 {
		p.logger.Info("✨  All contracts deployed successfully")
	} else {
		p.logger.Error("Failed to deploy all contracts")
		return nil, fmt.Errorf(`%v`, errs)
	}

	return contracts, nil
}

func prepareUpdateContractTransaction(
	targetAccount flowsdk.Address,
	contract *contracts.Contract,
) *flowsdk.Transaction {
	return templates.UpdateAccountContract(
		targetAccount,
		templates.Contract{
			Name:   contract.Name(),
			Source: contract.TranspiledCode(),
		},
	)
}

func prepareAddContractTransaction(
	targetAccount flowsdk.Address,
	contract *contracts.Contract,
) *flowsdk.Transaction {
	return templates.AddAccountContract(
		targetAccount,
		templates.Contract{
			Name:   contract.Name(),
			Source: contract.TranspiledCode(),
		},
	)
}
