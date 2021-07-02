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

	"github.com/onflow/flow-cli/internal/command"
	"github.com/onflow/flow-cli/pkg/flowkit"
	"github.com/onflow/flow-cli/pkg/flowkit/services"
)

type flagsEvents struct {
	Start   uint64 `flag:"start" info:"Block height start"`
	End     uint64 `flag:"end" info:"Block height end"`
	Last    uint64 `default:"1" flag:"last" info:"Fetch number of blocks relative to the last block. Ignored if the start flag is set. Used as a default if no flags are provided"`
	Workers int    `default:"10" flag:"workers" info:"Number of workers to use when fetching events in parallel"`
	Batch   uint64 `default:"250" flag:"batch" info:"Number of blocks to batch together when fetching events in parallel"`
}

var eventsFlags = flagsEvents{}

var GetCommand = &command.Command{
	Cmd: &cobra.Command{
		Use:   "get <event_name> (<event_name> <event_name>)",
		Short: "Get events in a block range",
		Args:  cobra.MinimumNArgs(1),
		Example: `#specify manual start and stop blocks
flow events get A.1654653399040a61.FlowToken.TokensDeposited --start 11559500 --end 11559600

#in order to get and event from the 10 latest blocks on a network run
flow events get A.1654653399040a61.FlowToken.TokensDeposited --last 10 --network mainnet
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
	start, end, err := services.Events.CalculateStartEnd(eventsFlags.Start, eventsFlags.End, eventsFlags.Last)
	if err != nil {
		return nil, err
	}

	events, err := services.Events.Get(args, start, end, eventsFlags.Batch, eventsFlags.Workers)
	if err != nil {
		return nil, err
	}

	return &EventResult{BlockEvents: events}, nil
}
