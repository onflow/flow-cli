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
	"context"
	flowsdk "github.com/onflow/flow-go-sdk"

	"github.com/spf13/cobra"

	"github.com/onflow/flow-cli/internal/command"
	"github.com/onflow/flow-cli/pkg/flowkit"
	"github.com/onflow/flow-cli/pkg/flowkit/output"
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

	var events []flowsdk.BlockEvents
	if blockFlags.Events != "" {
		events, err = flow.GetEvents(
			context.Background(),
			[]string{blockFlags.Events},
			block.Height,
			block.Height,
			nil,
		)
		if err != nil {
			return nil, err
		}
	}

	collections := make([]*flowsdk.Collection, 0)
	if command.ContainsFlag(blockFlags.Include, "transactions") {
		for _, guarantee := range block.CollectionGuarantees {
			collection, err := flow.GetCollection(context.Background(), guarantee.CollectionID)
			if err != nil {
				return nil, err
			}
			collections = append(collections, collection)
		}
	}

	return &BlockResult{
		block:       block,
		events:      events,
		collections: collections,
		included:    blockFlags.Include,
	}, nil
}
