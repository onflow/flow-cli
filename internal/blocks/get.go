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

package blocks

import (
	"github.com/spf13/cobra"

	"github.com/onflow/flow-cli/internal/command"
	"github.com/onflow/flow-cli/pkg/flowkit"
	"github.com/onflow/flow-cli/pkg/flowkit/services"
)

type flagsBlocks struct {
	Events  string   `default:"" flag:"events" info:"List events of this type for the block"`
	Include []string `default:"" flag:"include" info:"Fields to include in the output. Valid values: transactions."`
}

var blockFlags = flagsBlocks{}

var GetCommand = &command.Command{
	Cmd: &cobra.Command{
		Use:     "get <block_id|latest|block_height>",
		Short:   "Get block info",
		Example: "flow blocks get latest --network testnet",
		Args:    cobra.ExactArgs(1),
	},
	Flags: &blockFlags,
	Run:   get,
}

func get(
	args []string,
	_ flowkit.ReaderWriter,
	_ command.GlobalFlags,
	services *services.Services,
) (command.Result, error) {
	block, events, collections, err := services.Blocks.GetBlock(
		args[0], // block id
		blockFlags.Events,
		command.ContainsFlag(blockFlags.Include, "transactions"),
	)
	if err != nil {
		return nil, err
	}

	return &BlockResult{
		block:       block,
		events:      events,
		collections: collections,
		included:    blockFlags.Include,
	}, nil
}
