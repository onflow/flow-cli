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

	"github.com/onflow/flow-cli/pkg/flowkit"

	"github.com/spf13/cobra"

	"github.com/onflow/flow-cli/internal/command"
	"github.com/onflow/flow-cli/pkg/flowkit/services"
)

type flagsSendSigned struct {
	Include []string `default:"" flag:"include" info:"Fields to include in the output. Valid values: signatures, code, payload."`
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
	RunS:  sendSigned,
}

func sendSigned(
	args []string,
	readerWriter flowkit.ReaderWriter,
	globalFlags command.GlobalFlags,
	services *services.Services,
	_ *flowkit.State,
) (command.Result, error) {
	filename := args[0]

	code, err := readerWriter.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("error loading transaction payload: %w", err)
	}

	tx, result, err := services.Transactions.SendSigned(code, globalFlags.Yes)
	if err != nil {
		return nil, err
	}

	return &TransactionResult{
		result:  result,
		tx:      tx,
		include: sendSignedFlags.Include,
		exclude: sendSignedFlags.Exclude,
	}, nil
}
