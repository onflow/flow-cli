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

package evm

import (
	_ "embed"

	"github.com/spf13/cobra"

	"github.com/onflow/flow-cli/flowkit"
	"github.com/onflow/flow-cli/flowkit/output"
	"github.com/onflow/flow-cli/internal/command"
	"github.com/onflow/flow-cli/internal/transactions"
)

//go:embed create.cdc
var createCode []byte

type flagsCreate struct {
	Signer string `default:"" flag:"signer" info:"Account name from configuration used to sign the transaction as proposer, payer and suthorizer"`
}

var createFlags = flagsCreate{}

var createCommand = &command.Command{
	Cmd: &cobra.Command{
		Use:     "create-account <amount>",
		Short:   "Create a new EVM account and fund it with the amount as well as store the bridged account resource",
		Args:    cobra.ExactArgs(1),
		Example: "flow evm create-account 1.0",
	},
	Flags: &createFlags,
	RunS:  create,
}

// todo only for demo, super hacky now

func create(
	args []string,
	g command.GlobalFlags,
	_ output.Logger,
	flow flowkit.Services,
	state *flowkit.State,
) (command.Result, error) {
	amount := args[0]

	return transactions.SendTransaction(
		createCode,
		[]string{"", amount},
		"",
		flow,
		state,
		transactions.Flags{
			Signer: deployFlags.Signer,
		},
	)
}
