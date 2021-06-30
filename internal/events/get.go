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
	"github.com/spf13/cobra"
	"strings"

	"github.com/davecgh/go-spew/spew"
	"github.com/onflow/flow-cli/internal/command"
	"github.com/onflow/flow-cli/pkg/flowkit"
	"github.com/onflow/flow-cli/pkg/flowkit/services"
)

type flagsEvents struct{
	Start   uint64    `flag:"start" info:"Block height start"`
	End     uint64    `flag:"end" info:"Block height end"`
	Last    uint64   `default:"1" flag:"last" info:"Fetch last number of block relative to latestBlocks. Will be ignored if --start set"`
	Workers uint64   `default:"10" flag:"workers" info:"Number of workers to use when fetching events in parallel"`
	Batch   uint64   `default:"250" flag:"batch" info:"Number of blocks to batch together when fetching events in parallel"`
}

var eventsFlags = flagsEvents{}

var GetCommand = &command.Command{
	Cmd: &cobra.Command{
		Use:     "get <event_name> <block_height_range_start> <optional:block_height_range_end|latest>",
		Short:   "Get events in a block range",
		Args:    cobra.MinimumNArgs(1),
		Example: `flow events get A.1654653399040a61.FlowToken.TokensDeposited --start 11559500 --end 11559600
flow events get get A.1654653399040a61.FlowToken.TokensDeposited --last 10 --network mainnet
	`,
	},
	Flags: &eventsFlags,
	Run:   get,
}

func get(
	args []string,
	_ flowkit.ReaderWriter,
	_ command.GlobalFlags,
	services *services.Services,
) (command.Result, error) {
	name := args[0]

	start := args[1] // block height range start
	end := ""        // block height range end

	if len(args) == 3 {
		end = args[2]
	}

	spew.Dump(eventsFlags)

	eventNames := strings.Split(name, ",")
	events, err := services.Events.GetMany(eventNames, start, end, eventsFlags.Batch, 10)
	if err != nil {
		return nil, err
	}

	return &EventResult{BlockEvents: events}, nil
}
