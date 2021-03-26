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
	"fmt"

	"github.com/onflow/flow-cli/internal/command"
	"github.com/onflow/flow-cli/pkg/flowcli/services"
	"github.com/spf13/cobra"
)

type flagsGenerate struct {
	Verbose bool `flag:"verbose" info:"⚠️  DEPRECATED"`
}

var generateFlag = flagsGenerate{}

var GetCommand = &command.Command{
	Cmd: &cobra.Command{
		Use:     "get <event_name> <block_height_range_start> <optional:block_height_range_end|latest>",
		Short:   "Get events in a block range",
		Args:    cobra.RangeArgs(2, 3),
		Example: "flow events get A.1654653399040a61.FlowToken.TokensDeposited 11559500 11559600",
	},
	Flags: &generateFlag,
	Run: func(
		cmd *cobra.Command,
		args []string,
		globalFlags command.GlobalFlags,
		services *services.Services,
	) (command.Result, error) {
		if generateFlag.Verbose {
			return nil, fmt.Errorf("⚠️  DEPRECATED: flag is deperacated.")
		}

		end := ""
		if len(args) == 3 {
			end = args[2] // block height range end
		}

		events, err := services.Events.Get(
			args[0], // event name
			args[1], // block height range start
			end,
		)
		if err != nil {
			return nil, err
		}

		return &EventResult{BlockEvents: events}, nil
	},
}
