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

package transactions

import (
	"context"
	"fmt"

	"github.com/onflow/flow-cli/internal/prompt"

	"github.com/onflow/cadence"
	flowsdk "github.com/onflow/flow-go-sdk"
	"github.com/spf13/cobra"

	"github.com/onflow/flowkit/v2"
	"github.com/onflow/flowkit/v2/arguments"
	"github.com/onflow/flowkit/v2/output"
	"github.com/onflow/flowkit/v2/transactions"

	"github.com/onflow/flow-cli/internal/command"
)

type flagsBuild struct {
	ArgsJSON         string   `default:"" flag:"args-json" info:"arguments in JSON-Cadence format"`
	Proposer         string   `default:"emulator-account" flag:"proposer" info:"transaction proposer"`
	ProposerKeyIndex int      `default:"0" flag:"proposer-key-index" info:"proposer key index"`
	Payer            string   `default:"emulator-account" flag:"payer" info:"transaction payer"`
	Authorizer       []string `default:"emulator-account" flag:"authorizer" info:"transaction authorizer"`
	GasLimit         uint64   `default:"1000" flag:"gas-limit" info:"transaction gas limit"`
}

var buildFlags = flagsBuild{}

var buildCommand = &command.Command{
	Cmd: &cobra.Command{
		Use:     "build <code filename>  [<argument> <argument> ...]",
		Short:   "Build an unsigned transaction",
		Example: `flow transactions build ./transaction.cdc "Hello" --proposer alice --authorizer alice --payer bob`,
		Args:    cobra.MinimumNArgs(1),
	},
	Flags: &buildFlags,
	RunS:  build,
}

func build(
	args []string,
	globalFlags command.GlobalFlags,
	_ output.Logger,
	flow flowkit.Services,
	state *flowkit.State,
) (command.Result, error) {
	proposer, err := getAddress(buildFlags.Proposer, state)
	if err != nil {
		return nil, err
	}

	// get all authorizers
	var authorizers []flowsdk.Address
	for _, auth := range buildFlags.Authorizer {
		addr, err := getAddress(auth, state)
		if err != nil {
			return nil, err
		}
		authorizers = append(authorizers, addr)
	}

	payer, err := getAddress(buildFlags.Payer, state)
	if err != nil {
		return nil, err
	}

	filename := args[0]
	code, err := state.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("error loading transaction file: %w", err)
	}

	var transactionArgs []cadence.Value
	if buildFlags.ArgsJSON != "" {
		transactionArgs, err = arguments.ParseJSON(buildFlags.ArgsJSON)
	} else {
		transactionArgs, err = arguments.ParseWithoutType(args[1:], code, filename)
	}
	if err != nil {
		return nil, fmt.Errorf("error parsing transaction arguments: %w", err)
	}

	tx, err := flow.BuildTransaction(
		context.Background(),
		transactions.AddressesRoles{
			Proposer:    proposer,
			Authorizers: authorizers,
			Payer:       payer,
		},
		buildFlags.ProposerKeyIndex,
		flowkit.Script{
			Code:     code,
			Args:     transactionArgs,
			Location: filename,
		},
		buildFlags.GasLimit,
	)
	if err != nil {
		return nil, err
	}

	if !globalFlags.Yes && !prompt.ApproveTransactionForBuildingPrompt(tx.FlowTransaction()) {
		return nil, fmt.Errorf("transaction was not approved")
	}

	return &transactionResult{
		tx:      tx.FlowTransaction(),
		include: []string{"code", "payload", "signatures"},
	}, nil
}

func getAddress(address string, state *flowkit.State) (flowsdk.Address, error) {
	addr, valid := parseAddress(address)
	if valid {
		return addr, nil
	}

	// if address is not valid then try using the string as an account name.
	acc, err := state.Accounts().ByName(address)
	if err != nil {
		return flowsdk.EmptyAddress, err
	}
	return acc.Address, nil
}

func parseAddress(value string) (flowsdk.Address, bool) {
	address := flowsdk.HexToAddress(value)

	// valid on any chain
	return address, address.IsValid(flowsdk.Mainnet) ||
		address.IsValid(flowsdk.Testnet) ||
		address.IsValid(flowsdk.Emulator)
}
