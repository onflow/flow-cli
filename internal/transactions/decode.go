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

	"github.com/onflow/flowkit/transactions"

	"github.com/spf13/cobra"

	"github.com/onflow/flowkit"
	"github.com/onflow/flowkit/output"

	"github.com/onflow/flow-cli/internal/command"
)

type flagsDecode struct {
	Include []string `default:"" flag:"include" info:"Fields to include in the output. Valid values: signatures, code, payload."`
}

var decodeFlags = flagsDecode{}

var decodeCommand = &command.Command{
	Cmd: &cobra.Command{
		Use:     "decode <transaction filename>",
		Short:   "Decode a transaction",
		Example: "flow transactions decode ./transaction.rlp",
		Args:    cobra.ExactArgs(1),
	},
	Flags: &decodeFlags,
	Run:   decode,
}

func decode(
	args []string,
	_ command.GlobalFlags,
	_ output.Logger,
	reader flowkit.ReaderWriter,
	_ flowkit.Services,
) (command.Result, error) {
	filename := args[0]
	payload, err := reader.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read transaction from %s: %v", filename, err)
	}

	tx, err := transactions.NewFromPayload(payload)
	if err != nil {
		return nil, err
	}

	return &transactionResult{
		tx:      tx.FlowTransaction(),
		include: decodeFlags.Include,
	}, nil
}
