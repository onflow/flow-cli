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
package transactions

import (
	"fmt"

	"github.com/onflow/flow-cli/pkg/flowkit"

	"github.com/onflow/flow-go-sdk"
	"github.com/spf13/cobra"

	"github.com/onflow/flow-cli/internal/command"
	"github.com/onflow/flow-cli/pkg/flowkit/services"
)

type flagsBuild struct {
	ArgsJSON         string   `default:"" flag:"args-json" info:"arguments in JSON-Cadence format"`
	Args             []string `default:"" flag:"arg" info:"argument in Type:Value format"`
	Proposer         string   `default:"emulator-account" flag:"proposer" info:"transaction proposer"`
	ProposerKeyIndex int      `default:"0" flag:"proposer-key-index" info:"proposer key index"`
	Payer            string   `default:"emulator-account" flag:"payer" info:"transaction payer"`
	Authorizer       []string `default:"emulator-account" flag:"authorizer" info:"transaction authorizer"`
	GasLimit         uint64   `default:"1000" flag:"gas-limit" info:"transaction gas limit"`
}

var buildFlags = flagsBuild{}

var BuildCommand = &command.Command{
	Cmd: &cobra.Command{
		Use:     "build <code filename>",
		Short:   "Build an unsigned transaction",
		Example: "flow transactions build ./transaction.cdc --proposer alice --authorizer alice --payer bob",
		Args:    cobra.ExactArgs(1),
	},
	Flags: &buildFlags,
	RunS: func(
		cmd *cobra.Command,
		args []string,
		readerWriter flowkit.ReaderWriter,
		globalFlags command.GlobalFlags,
		services *services.Services,
		state *flowkit.State,
	) (command.Result, error) {
		proposer, err := getAddress(buildFlags.Proposer, state)
		if err != nil {
			return nil, err
		}

		// get all authorizers
		var authorizers []flow.Address
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
		code, err := loader.ReadFile(filename)
		if err != nil {
			return nil, fmt.Errorf("error loading transaction file: %w", err)
		}

		txArgs, err := flowkit.ParseArguments(buildFlags.Args, buildFlags.ArgsJSON) // todo refactor flowkit
		if err != nil {
			return nil, err
		}

		build, err := services.Transactions.Build(
			proposer,
			authorizers,
			payer,
			buildFlags.ProposerKeyIndex,
			code,
			filename,
			buildFlags.GasLimit,
			txArgs,
			globalFlags.Network,
		)
		if err != nil {
			return nil, err
		}

		return &TransactionResult{
			tx:      build.FlowTransaction(),
			include: []string{"code", "payload", "signatures"},
		}, nil
	},
}

func getAddress(address string, state *flowkit.State) (flow.Address, error) {
	addr := flow.HexToAddress(address)
	if addr == flow.EmptyAddress {
		acc := state.Accounts().ByName(address)
		if acc == nil {
			return flow.EmptyAddress, fmt.Errorf("account not found, make sure to pass valid account name from configuration or valid flow address")
		}
		addr = acc.Address()
	}

	return addr, nil
}
