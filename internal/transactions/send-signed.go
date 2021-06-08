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
	"github.com/onflow/flow-cli/pkg/flowcli/project"
	"github.com/onflow/flow-cli/pkg/flowcli/services"
)

type flagsSendSigned struct {
	Include []string `default:"" flag:"include" info:"Fields to include in the output"`
	Exclude []string `default:"" flag:"exclude" info:"Fields to exclude from the output (events)"`
}

var sendSignedFlags = flagsSendSigned{}

var SendSignedCommand = &command.Command{
	Cmd: &cobra.Command{
		Use:     "send-signed <signed transaction filename>",
		Short:   "Send signed transaction",
		Args:    cobra.ExactArgs(1),
		Example: `flow transactions send-signed signed.rlp`,
	},
	Flags: &sendSignedFlags,
	Run: func(
		cmd *cobra.Command,
		args []string,
		globalFlags command.GlobalFlags,
		services *services.Services,
		project *project.Project,
	) (command.Result, error) {

		tx, result, err := services.Transactions.SendSigned(
			args[0], // signed filename
		)
		if err != nil {
			return nil, err
		}

		return &TransactionResult{
			result:  result,
			tx:      tx,
			include: sendSignedFlags.Include,
			exclude: sendSignedFlags.Exclude,
		}, nil
	},
}
