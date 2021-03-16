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
	"github.com/onflow/flow-cli/cmd"
	"github.com/onflow/flow-cli/flow/lib"
	"github.com/onflow/flow-cli/flow/services"
	"github.com/psiemens/sconfig"
	"github.com/spf13/cobra"
)

type flagsBlocks struct {
	Events  string `default:"" flag:"events" info:"List events of this type for the block"`
	Verbose bool   `default:"false" flag:"verbose" info:"Display transactions in block"`
}

type cmdGet struct {
	cmd   *cobra.Command
	flags flagsBlocks
}

// NewGetCmd creates new get command
func NewGetCmd() cmd.Command {
	return &cmdGet{
		cmd: &cobra.Command{
			Use:   "get <block_id|latest|block_height>",
			Short: "Get block info",
		},
	}
}

// Run block command
func (s *cmdGet) Run(
	cmd *cobra.Command,
	args []string,
	project *lib.Project,
	services *services.Services,
) (cmd.Result, error) {
	block, events, err := services.Blocks.GetBlock(args[0], s.flags.Events)

	return &BlockResult{
		block:   block,
		events:  events,
		verbose: false,
	}, err
}

// GetFlags for blocks
func (s *cmdGet) GetFlags() *sconfig.Config {
	return sconfig.New(&s.flags)
}

// GetCmd get command
func (s *cmdGet) GetCmd() *cobra.Command {
	return s.cmd
}
