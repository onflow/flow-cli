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

	"github.com/spf13/cobra"

	"github.com/onflow/flow-cli/internal/command"
	"github.com/onflow/flow-cli/pkg/flowkit"
	"github.com/onflow/flow-cli/pkg/flowkit/services"
)

type flagsGenerate struct {
	Verbose bool `flag:"verbose" info:"⚠️  Deprecated"`
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
	Run:   get,
}

func get(
	args []string,
	readerWriter flowkit.ReaderWriter,
	globalFlags command.GlobalFlags,
	services *services.Services,
) (command.Result, error) {
	if generateFlag.Verbose {
		fmt.Println("⚠️  DEPRECATION WARNING: verbose flag is deprecated")
	}
	name := args[0]
	start := args[1] // block height range start
	end := ""        // block height range end

	if len(args) == 3 {
		end = args[2]
	}

	events, err := services.Events.Get(name, start, end)
	if err != nil {
		return nil, err
	}

	return &EventResult{BlockEvents: events}, nil
}
