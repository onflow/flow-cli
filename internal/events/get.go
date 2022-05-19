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

package events

import (
	"fmt"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/onflow/flow-cli/internal/command"
	"github.com/onflow/flow-cli/pkg/flowkit"
	"github.com/onflow/flow-cli/pkg/flowkit/services"
)

type flagsEvents struct {
	Start   uint64 `flag:"start" info:"Start block height"`
	End     uint64 `flag:"end" info:"End block height"`
	Last    uint64 `default:"10" flag:"last" info:"Fetch number of blocks relative to the last block. Ignored if the start flag is set. Used as a default if no flags are provided"`
	Workers int    `default:"10" flag:"workers" info:"Number of workers to use when fetching events in parallel"`
	Batch   uint64 `default:"25" flag:"batch" info:"Number of blocks each worker will fetch"`
}

var eventsFlags = flagsEvents{}

var GetCommand = &command.Command{
	Cmd: &cobra.Command{
		Use:   "get <event_name>",
		Short: "Get events in a block range",
		Args:  cobra.MinimumNArgs(1),
		Example: `#fetch events from the latest 10 blocks is the default behavior
flow events get A.1654653399040a61.FlowToken.TokensDeposited

#specify manual start and stop blocks
flow events get A.1654653399040a61.FlowToken.TokensDeposited --start 11559500 --end 11559600

#in order to get and event from the 20 latest blocks on a network run
flow events get A.1654653399040a61.FlowToken.TokensDeposited --last 20 --network mainnet

#if you want to fetch multiple event types that is done by sending in more events. Even fetching will be done in parallel.
flow events get A.1654653399040a61.FlowToken.TokensDeposited A.1654653399040a61.FlowToken.TokensWithdrawn
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
	var err error
	start := eventsFlags.Start
	end := eventsFlags.End
	last := eventsFlags.Last

	// backward compatibility check
	if len(args) >= 2 {
		startV25, err := strconv.ParseUint(args[1], 10, 64)
		// only do backward compatibility if second argument is a number
		// since we can actually provide multiple arguments with the new format.
		if err == nil {
			fmt.Println("\n⚠️  DEPRECATION WARNING: use the new command format, see the get event command for help")
			start = startV25
			end = startV25

			if len(args) == 3 {
				endV25 := args[2]
				if endV25 == "latest" {
					latest, err := services.Blocks.GetLatestBlockHeight()
					if err != nil {
						return nil, err
					}

					end = latest
				} else {
					end, err = strconv.ParseUint(endV25, 10, 64)
					if err != nil {
						return nil, fmt.Errorf("failed to parse end height of block range: %s", endV25)
					}
				}
			}
		}
	}

	// handle if not passing start and end
	if start == 0 && end == 0 {
		//cannot use := here as it will not overrwrite end in the outer scope if you do
		end, err = services.Blocks.GetLatestBlockHeight()
		if err != nil {
			return nil, err
		}

		if last > end {
			start = 0
		} else {
			start = end - last
		}
	} else if start == 0 || end == 0 {
		return nil, fmt.Errorf("please provide either both start and end for range or only last flag")
	}

	events, err := services.Events.Get(args, start, end, eventsFlags.Batch, eventsFlags.Workers)
	if err != nil {
		return nil, err
	}

	return &EventResult{BlockEvents: events}, nil
}
