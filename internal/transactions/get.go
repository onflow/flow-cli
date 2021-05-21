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

	"github.com/spf13/cobra"

	"github.com/onflow/flow-cli/internal/command"
	"github.com/onflow/flow-cli/pkg/flowcli/services"
)

type flagsGet struct {
	Sealed  bool     `default:"true" flag:"sealed" info:"Wait for a sealed result"`
	Code    bool     `default:"false" flag:"code" info:"⚠️  Deprecated: use include flag"`
	Include []string `default:"" flag:"include" info:"Fields to include in the output"`
	Exclude []string `default:"" flag:"exclude" info:"Fields to exclude from the output (events)"`
}

var getFlags = flagsGet{}

var GetCommand = &command.Command{
	Cmd: &cobra.Command{
		Use:     "get <tx_id>",
		Aliases: []string{"status"},
		Short:   "Get the transaction by ID",
		Example: "flow transactions get 07a8...b433",
		Args:    cobra.ExactArgs(1),
	},
	Flags: &getFlags,
	Run: func(
		cmd *cobra.Command,
		args []string,
		globalFlags command.GlobalFlags,
		services *services.Services,
	) (command.Result, error) {
		if cmd.CalledAs() == "status" {
			fmt.Println("⚠️  DEPRECATION WARNING: use \"flow transactions get\" instead")
		}

		if getFlags.Code {
			fmt.Println("⚠️  DEPRECATION WARNING: use include flag instead")
		}

		tx, result, err := services.Transactions.GetStatus(
			args[0], // transaction id
			getFlags.Sealed,
		)
		if err != nil {
			return nil, err
		}

		return &TransactionResult{
			result:  result,
			tx:      tx,
			code:    getFlags.Code,
			include: getFlags.Include,
			exclude: getFlags.Exclude,
		}, nil
	},
}
