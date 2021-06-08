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

	"github.com/onflow/flow-cli/pkg/flowcli"

	"github.com/onflow/flow-cli/pkg/flowcli/util"
	"github.com/onflow/flow-go-sdk"
	"github.com/spf13/cobra"

	"github.com/onflow/flow-cli/internal/command"
	"github.com/onflow/flow-cli/pkg/flowcli/services"
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
	Run: func(
		cmd *cobra.Command,
		args []string,
		globalFlags command.GlobalFlags,
		services *services.Services,
	) (command.Result, error) {
		proposer := flow.HexToAddress(buildFlags.Proposer)
		if proposer == flow.EmptyAddress {
			// todo get from project
		}

		// get all authorizers
		var authorizers []flow.Address
		for _, auth := range buildFlags.Authorizer {
			addr := flow.HexToAddress(auth)
			if addr == flow.EmptyAddress {
				// todo get from project
			}
			authorizers = append(authorizers, addr)
		}

		payer := flow.HexToAddress(buildFlags.Payer)
		if proposer == flow.EmptyAddress {
			// todo get from project
		}

		filename := args[0]
		code, err := util.LoadFile(filename)
		if err != nil {
			return nil, fmt.Errorf("error loading transaction file: %w", err)
		}

		txArgs, err := flowcli.ParseArguments(buildFlags.Args, buildFlags.ArgsJSON) // todo refactor flowcli
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
