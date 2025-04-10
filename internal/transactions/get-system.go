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
		Use:     "get-system <block_id>",
		Short:   "Get the system transaction by block ID",
		Example: "flow transactions get-system a1b2c3...",
		Args:    cobra.ExactArgs(1),
	},
	Flags: &getSystemFlags,
	Run:   getSystemTransaction,
}

func getSystemTransaction(
	args []string,
	_ command.GlobalFlags,
	_ output.Logger,
	_ flowkit.ReaderWriter,
	flow flowkit.Services,
) (command.Result, error) {
	blockID := flowsdk.HexToID(strings.TrimPrefix(args[0], "0x"))

	tx, result, err := flow.GetSystemTransaction(context.Background(), blockID)
	if err != nil {
		return nil, err
	}

	return &transactionResult{
		result:  result,
		tx:      tx,
		include: getSystemFlags.Include,
		exclude: getSystemFlags.Exclude,
	}, nil
}
