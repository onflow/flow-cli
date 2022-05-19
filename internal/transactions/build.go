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
	"fmt"

	"github.com/onflow/cadence"
	"github.com/onflow/flow-go-sdk"
	"github.com/spf13/cobra"

	"github.com/onflow/flow-cli/internal/command"
	"github.com/onflow/flow-cli/pkg/flowkit"
	"github.com/onflow/flow-cli/pkg/flowkit/services"
	"github.com/onflow/flow-cli/pkg/flowkit/util"
)

type flagsBuild struct {
	ArgsJSON         string   `default:"" flag:"args-json" info:"arguments in JSON-Cadence format"`
	Args             []string `default:"" flag:"arg" info:"⚠️  Deprecated: use command arguments"`
	Proposer         string   `default:"emulator-account" flag:"proposer" info:"transaction proposer"`
	ProposerKeyIndex int      `default:"0" flag:"proposer-key-index" info:"proposer key index"`
	Payer            string   `default:"emulator-account" flag:"payer" info:"transaction payer"`
	Authorizer       []string `default:"emulator-account" flag:"authorizer" info:"transaction authorizer"`
	GasLimit         uint64   `default:"1000" flag:"gas-limit" info:"transaction gas limit"`
}

var buildFlags = flagsBuild{}

var BuildCommand = &command.Command{
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
	code, err := readerWriter.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("error loading transaction file: %w", err)
	}

	if len(buildFlags.Args) != 0 {
		fmt.Println("⚠️  DEPRECATION WARNING: use transaction arguments as command arguments: send <code filename> [<argument> <argument> ...]")
	}

	var transactionArgs []cadence.Value
	if buildFlags.ArgsJSON != "" || len(buildFlags.Args) != 0 {
		transactionArgs, err = flowkit.ParseArguments(buildFlags.Args, buildFlags.ArgsJSON)
	} else {
		transactionArgs, err = flowkit.ParseArgumentsWithoutType(filename, code, args[1:])
	}

	if err != nil {
		return nil, fmt.Errorf("error parsing transaction arguments: %w", err)
	}

	build, err := services.Transactions.Build(
		proposer,
		authorizers,
		payer,
		buildFlags.ProposerKeyIndex,
		code,
		filename,
		buildFlags.GasLimit,
		transactionArgs,
		globalFlags.Network,
		globalFlags.Yes,
	)
	if err != nil {
		return nil, err
	}

	return &TransactionResult{
		tx:      build.FlowTransaction(),
		include: []string{"code", "payload", "signatures"},
	}, nil
}

func getAddress(address string, state *flowkit.State) (flow.Address, error) {
	addr, valid := util.ParseAddress(address)
	if valid {
		return addr, nil
	}

	// if address is not valid then try using the string as an account name.
	acc, err := state.Accounts().ByName(address)
	if err != nil {
		return flow.EmptyAddress, err
	}
	return acc.Address(), nil
}
