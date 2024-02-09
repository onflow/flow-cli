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

package migration

import (
	"context"
	"fmt"

	"github.com/onflow/cadence"
	flowsdk "github.com/onflow/flow-go-sdk"
	"github.com/onflow/flowkit"
	"github.com/onflow/flowkit/accounts"
	"github.com/onflow/flowkit/output"
	"github.com/onflow/flowkit/transactions"
	"github.com/spf13/cobra"

	"github.com/onflow/flow-cli/internal/command"
)

var stageContractflags interface{}

var stageContractCommand = &command.Command{
	Cmd: &cobra.Command{
		Use:     "flow stage-contract <NAME> --network <NETWORK> --signer <HOST_ACCOUNT>",
		Short:   "stage a contract for migration",
		Example: `flow stage-contract HelloWorld --network testnet --signer emulator-account`,
		Args:    cobra.MinimumNArgs(1),
	},
	Flags: &stageContractflags,
	RunS:  stageContract,
}

func stageContract(
	args []string,
	globalFlags command.GlobalFlags,
	_ output.Logger,
	flow flowkit.Services,
	state *flowkit.State,
) (command.Result, error) {
	code, err := RenderContractTemplate(UnstageContractTransactionFilepath, globalFlags.Network)
	if err != nil {
		return nil, fmt.Errorf("error loading staging contract file: %w", err)
	}
	
	contractName := args[0]

	contracts := state.Contracts()
	if contracts == nil {
		return nil, fmt.Errorf("no contracts found in state")
	}

	var contractPath string
	for _, c := range *contracts {
		if contractName == c.Name {
			contractPath = c.Location
			break
		}
	}

	contractCode, err := state.ReadFile(contractPath)
	if err != nil {
		return nil, fmt.Errorf("error loading contract file: %w", err)
	}

	deployments := state.Deployments().ByNetwork(globalFlags.Network)
	var accountName string
	for _, d := range deployments {
		for _, c := range d.Contracts {
			if c.Name == contractName {
				accountName = d.Account
				break
			}
		}
	}

	accs := state.Accounts()
	if accs == nil {
		return nil, fmt.Errorf("no accounts found in state")
	}

	var accountToDeploy *accounts.Account
	for _, a := range *accs {
		if accountName == a.Name {
			accountToDeploy = &a
			break
		}
	}
	if accountToDeploy == nil {
		return nil, fmt.Errorf("account %s not found in state", accountName)
	}

	cName, err := cadence.NewString(contractName)
	if err != nil {
		return nil, fmt.Errorf("failed to get cadence string from contract name: %w", err)
	}

	cCode, err := cadence.NewString(string(contractCode))
	if err != nil {
		return nil, fmt.Errorf("failed to get cadence string from contract code: %w", err)
	}

	_, _, err = flow.SendTransaction(
		context.Background(),
		transactions.AccountRoles{
			Proposer:    *accountToDeploy,
			Authorizers: []accounts.Account{*accountToDeploy},
			Payer:       *accountToDeploy,
		},
		flowkit.Script{
			Code: code,
			Args: []cadence.Value{cName, cCode},
		},
		flowsdk.DefaultTransactionGasLimit,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to send transaction: %w", err)
	}

	return nil, nil
}
