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
	debugger "github.com/onflow/execution-debugger"
	"github.com/onflow/execution-debugger/debuggers"
	"github.com/onflow/flow-cli/flowkit"
	"github.com/onflow/flow-cli/flowkit/output"
	"github.com/onflow/flow-cli/internal/command"
	"github.com/onflow/flow-emulator/convert/sdk"
	flowsdk "github.com/onflow/flow-go-sdk"
	flowGo "github.com/onflow/flow-go/model/flow"
	"github.com/rs/zerolog"
	"github.com/spf13/cobra"
	"os"
)

type flagsDebug struct {
	Include []string `default:"" flag:"include" info:"Fields to include in the output. Valid values: signatures, code, payload."`
	Exclude []string `default:"" flag:"exclude" info:"Fields to exclude from the output. Valid values: events."`
}

var debugFlags = flagsDebug{}

var debugCommand = &command.Command{
	Cmd: &cobra.Command{
		Use:     "debug <transaction id | transaction filename>",
		Short:   "Debug an existing on-chain transaction by ID or a non-existing RLP encoded transaction from a file",
		Example: "flow transactions debug ./test.rlp",
		Args:    cobra.ExactArgs(1),
	},
	Flags: &debugFlags,
	Run:   debug,
}

func debug(
	args []string,
	_ command.GlobalFlags,
	_ output.Logger,
	_ flowkit.ReaderWriter,
	_ flowkit.Services,
) (command.Result, error) {
	idArg := args[0]

	logger := zerolog.New(zerolog.ConsoleWriter{Out: os.Stderr})
	dbg, err := debuggers.NewMainnetExecutionDebugger(logger)
	if err != nil {
		return nil, err
	}

	id, err := flowGo.HexStringToIdentifier(idArg)
	if err != nil {
		return nil, err
	}

	txResolver := &debugger.NetworkTransactions{
		Client: dbg.Client(),
		ID:     id,
	}

	result, err := dbg.DebugTransaction(txResolver)
	if err != nil {
		return nil, err
	}

	txBody, err := txResolver.TransactionBody()
	if err != nil {
		return nil, err
	}

	tx := sdk.FlowTransactionToSDK(*txBody)

	height, err := txResolver.BlockHeight()
	events, err := sdk.FlowEventsToSDK(result.Execution.Output.Events)

	txResult := &flowsdk.TransactionResult{
		Status:      flowsdk.TransactionStatusUnknown,
		Error:       result.Execution.Output.Err,
		Events:      events,
		BlockHeight: height,
	}

	return &transactionResult{
		result: txResult,
		tx:     &tx,
	}, nil
}
