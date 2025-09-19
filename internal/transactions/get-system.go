/*
 * Flow CLI
 *
 * Copyright Flow Foundation
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
	"strings"

	flowsdk "github.com/onflow/flow-go-sdk"
	"github.com/spf13/cobra"

	"github.com/onflow/flowkit/v2"
	"github.com/onflow/flowkit/v2/output"

	"github.com/onflow/flow-cli/internal/command"
)

type flagsGetSystem struct {
	Include []string `default:"" flag:"include" info:"Fields to include in the output. Valid values: signatures, code, payload, fee-events."`
	Exclude []string `default:"" flag:"exclude" info:"Fields to exclude from the output. Valid values: events."`
}

var getSystemFlags = flagsGetSystem{}

var getSystemCommand = &command.Command{
	Cmd: &cobra.Command{
		Use:     "get-system <block_id|latest|block_height> [tx_id]",
		Short:   "Get the system transaction by block and optional ID",
		Example: "flow transactions get-system latest\nflow transactions get-system latest 07a8...b433",
		Args:    cobra.RangeArgs(1, 2),
	},
	Flags: &getSystemFlags,
	Run:   getSystemTransaction,
}

func getSystemTransaction(
	args []string,
	_ command.GlobalFlags,
	logger output.Logger,
	_ flowkit.ReaderWriter,
	flow flowkit.Services,
) (command.Result, error) {
	query, err := flowkit.NewBlockQuery(args[0])
	if err != nil {
		return nil, err
	}

	logger.StartProgress("Fetching Block...")
	defer logger.StopProgress()
	block, err := flow.GetBlock(context.Background(), query)
	if err != nil {
		return nil, err
	}

	var tx *flowsdk.Transaction
	var result *flowsdk.TransactionResult

	if len(args) == 2 {
		// Parse transaction ID if provided
		id := flowsdk.HexToID(strings.TrimPrefix(args[1], "0x"))

		// Fetch transaction and result by ID
		t, err := flow.GetSystemTransactionWithID(context.Background(), block.ID, id)
		if err != nil {
			return nil, err
		}
		r, err := flow.GetSystemTransactionResultWithID(context.Background(), block.ID, id)
		if err != nil {
			return nil, err
		}
		tx = t
		result = r
	} else {
		// Fallback to last system transaction in the block
		t, r, err := flow.GetSystemTransaction(context.Background(), block.ID)
		if err != nil {
			return nil, err
		}
		tx = t
		result = r
	}

	return &transactionResult{
		result:  result,
		tx:      tx,
		include: getSystemFlags.Include,
		exclude: getSystemFlags.Exclude,
		network: flow.Network().Name,
	}, nil
}
