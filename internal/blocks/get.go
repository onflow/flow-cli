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

package blocks

import (
	"github.com/onflow/flow-cli/internal/command"
	"github.com/onflow/flow-cli/pkg/flowcli/services"
	"github.com/spf13/cobra"
)

type flagsBlocks struct {
	Events  string `default:"" flag:"events" info:"List events of this type for the block"`
	Verbose bool   `default:"false" flag:"verbose" info:"Display transactions in block"`
}

var blockFlags = flagsBlocks{}

var GetCommand = &command.Command{
	Cmd: &cobra.Command{
		Use:   "get <block_id|latest|block_height>",
		Short: "Get block info",
	},
	Flags: &blockFlags,
	Run: func(
		cmd *cobra.Command,
		args []string,
		globalFlags command.GlobalFlags,
		services *services.Services,
	) (command.Result, error) {
		block, events, collections, err := services.Blocks.GetBlock(
			args[0], // block id
			blockFlags.Events,
			blockFlags.Verbose,
		)
		if err != nil {
			return nil, err
		}

		return &BlockResult{
			block:       block,
			events:      events,
			verbose:     blockFlags.Verbose,
			collections: collections,
		}, nil
	},
}
