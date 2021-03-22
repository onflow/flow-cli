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
	"github.com/onflow/flow-cli/internal/command"
	"github.com/onflow/flow-cli/pkg/flow"
	"github.com/onflow/flow-cli/pkg/flow/services"
	"github.com/spf13/cobra"
)

type flagsStatus struct {
	Sealed bool `default:"true" flag:"sealed" info:"Wait for a sealed result"`
	Code   bool `default:"false" flag:"code" info:"Display transaction code"`
}

var statusFlags = &flagsStatus{}

var GetCommand = &command.Command{
	Cmd: &cobra.Command{
		Use:   "get <tx_id>",
		Short: "Get the transaction status",
		Args:  cobra.ExactArgs(1),
	},
	Flags: &statusFlags,
	Run: func(
		cmd *cobra.Command,
		args []string,
		project *flow.Project,
		services *services.Services,
	) (command.Result, error) {
		tx, result, err := services.Transactions.GetStatus(args[0], statusFlags.Sealed)
		if err != nil {
			return nil, err
		}

		return &TransactionResult{
			result: result,
			tx:     tx,
			code:   statusFlags.Code,
		}, nil
	},
}
