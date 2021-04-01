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

	"github.com/onflow/flow-cli/internal/command"
	"github.com/onflow/flow-cli/pkg/flowcli/services"
	"github.com/spf13/cobra"
)

type flagsSend struct {
	ArgsJSON string   `default:"" flag:"args-json" info:"arguments in JSON-Cadence format"`
	Args     []string `default:"" flag:"arg" info:"argument in Type:Value format"`
	Signer   string   `default:"emulator-account" flag:"signer"`
	Code     string   `default:"" flag:"code" info:"⚠️ No longer supported: use filename argument"`
	Results  bool     `default:"" flag:"results" info:"⚠️ No longer supported: all transactions will provide result"`
}

var sendFlags = flagsSend{}

var SendCommand = &command.Command{
	Cmd: &cobra.Command{
		Use:     "send <filename>",
		Short:   "Send a transaction",
		Args:    cobra.ExactArgs(1),
		Example: `flow transactions send tx.cdc --arg String:"Hello world"`,
	},
	Flags: &sendFlags,
	Run: func(
		cmd *cobra.Command,
		args []string,
		globalFlags command.GlobalFlags,
		services *services.Services,
	) (command.Result, error) {
		if sendFlags.Code != "" {
			return nil, fmt.Errorf("⚠️ No longer supported: use filename argument")
		}

		if sendFlags.Results {
			return nil, fmt.Errorf("⚠️ No longer supported: all transactions will provide results")
		}

		tx, result, err := services.Transactions.Send(
			args[0], // filename
			sendFlags.Signer,
			sendFlags.Args,
			sendFlags.ArgsJSON,
		)
		if err != nil {
			return nil, err
		}

		return &TransactionResult{
			result: result,
			tx:     tx,
		}, nil
	},
}
