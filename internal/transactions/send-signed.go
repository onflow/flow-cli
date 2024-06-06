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

	"github.com/onflow/flowkit/v2/transactions"

	"github.com/spf13/cobra"

	"github.com/onflow/flowkit/v2"
	"github.com/onflow/flowkit/v2/output"

	"github.com/onflow/flow-cli/internal/command"
)

type flagsSendSigned struct {
	Include []string `default:"" flag:"include" info:"Fields to include in the output. Valid values: signatures, code, payload."`
	Exclude []string `default:"" flag:"exclude" info:"Fields to exclude from the output (events)"`
}

var sendSignedFlags = flagsSendSigned{}

var sendSignedCommand = &command.Command{
	Cmd: &cobra.Command{
		Use:     "send-signed <signed transaction filename>",
		Short:   "Send signed transaction",
		Args:    cobra.ExactArgs(1),
		Example: `flow transactions send-signed signed.rlp`,
	},
	Flags: &sendSignedFlags,
	Run:   sendSigned,
}

func sendSigned(
	args []string,
	globalFlags command.GlobalFlags,
	logger output.Logger,
	reader flowkit.ReaderWriter,
	flow flowkit.Services,
) (command.Result, error) {
	filename := args[0]

	code, err := reader.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("error loading transaction payload: %w", err)
	}

	tx, err := transactions.NewFromPayload(code)
	if err != nil {
		return nil, err
	}

	if !globalFlags.Yes && !prompt.ApproveTransactionForSendingPrompt(tx.FlowTransaction()) {
		return nil, fmt.Errorf("transaction was not approved for sending")
	}

	logger.StartProgress(fmt.Sprintf("Sending transaction with ID: %s", tx.FlowTransaction().ID()))
	defer logger.StopProgress()

	sentTx, result, err := flow.SendSignedTransaction(context.Background(), tx)
	if err != nil {
		return nil, err
	}

	return &transactionResult{
		result:  result,
		tx:      sentTx,
		include: sendSignedFlags.Include,
		exclude: sendSignedFlags.Exclude,
	}, nil
}
