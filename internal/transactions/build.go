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
}

var buildFlags = flagsBuild{}

var BuildCommand = &command.Command{
	Cmd: &cobra.Command{
		Use:     "build <code filename>",
		Short:   "Build a transaction for later signing",
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

		globalFlags.Filter = "payload"

		build, err := services.Transactions.Build(
			buildFlags.Proposer,
			buildFlags.Authorizer,
			buildFlags.Payer,
			buildFlags.ProposerKeyIndex,
			args[0], // code filename
			buildFlags.Args,
			buildFlags.ArgsJSON,
		)
		if err != nil {
			return nil, err
		}

		return &TransactionResult{
			tx: build.FlowTransaction(),
		}, nil
	},
}
