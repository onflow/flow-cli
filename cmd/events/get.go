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

package events

import (
	"github.com/onflow/flow-cli/cmd"
	"github.com/onflow/flow-cli/flow/lib"
	"github.com/onflow/flow-cli/flow/services"
	"github.com/psiemens/sconfig"
	"github.com/spf13/cobra"
)

type flagsGenerate struct {
}

type cmdGet struct {
	cmd   *cobra.Command
	flags flagsGenerate
}

// NewGetCmd return new command
func NewGetCmd() cmd.Command {
	return &cmdGet{
		cmd: &cobra.Command{
			Use:     "get <event_name> <block_height_range_start> <optional:block_height_range_end|latest>",
			Short:   "Get events in a block range",
			Args:    cobra.RangeArgs(2, 3),
			Example: "flow events get A.1654653399040a61.FlowToken.TokensDeposited 11559500 11559600",
		},
	}
}

// Run get event command
func (a *cmdGet) Run(
	cmd *cobra.Command,
	args []string,
	project *lib.Project,
	services *services.Services,
) (cmd.Result, error) {
	end := ""
	if len(args) == 3 {
		end = args[2]
	}

	events, err := services.Events.Get(args[0], args[1], end)
	return &EventResult{BlockEvents: events}, err
}

// GetFlags for the event
func (a *cmdGet) GetFlags() *sconfig.Config {
	return sconfig.New(&a.flags)
}

// GetCmd gets event command
func (a *cmdGet) GetCmd() *cobra.Command {
	return a.cmd
}
