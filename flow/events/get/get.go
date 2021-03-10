/*
 * Flow CLI
 *
 * Copyright 2019-2020 Dapper Labs, Inc.
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

package get

import (
	"log"
	"strconv"
	"strings"

	"github.com/onflow/flow-cli/flow/cli"

	"github.com/psiemens/sconfig"
	"github.com/spf13/cobra"
)

type Flags struct {
	Host    string `flag:"host" info:"Flow Access API host address"`
	Verbose bool   `flag:"verbose" info:"Verbose output"`
}

var flags Flags

var Cmd = &cobra.Command{
	Use:     "get <event_name> <block_height_range_start> <optional:block_height_range_end|latest>",
	Short:   "Get events in a block range",
	Args:    cobra.RangeArgs(2, 3),
	Example: "flow events get A.1654653399040a61.FlowToken.TokensDeposited 11559500 11559600",
	Run: func(cmd *cobra.Command, args []string) {
		host, err := cli.LoadHostForNetwork(flags.Host, "")
		if err != nil {
			cli.Exitf(1, err.Error())
		}

		eventName, startHeight, endHeight := validateArguments(host, args)

		cli.GetBlockEvents(host, startHeight, endHeight, eventName, flags.Verbose)
	},
}

func init() {
	initConfig()
}

func initConfig() {
	err := sconfig.New(&flags).
		FromEnvironment(cli.EnvPrefix).
		BindFlags(Cmd.PersistentFlags()).
		Parse()
	if err != nil {
		log.Fatal(err)
	}
}

func validateArguments(host string, args []string) (eventName string, startHeight, endHeight uint64) {
	var err error
	eventName = args[0]
	if len(eventName) == 0 {
		cli.Exitf(1, "Cannot use empty string as event name")
	}

	startHeight, err = strconv.ParseUint(args[1], 10, 64)
	if err != nil {
		cli.Exitf(1, "Failed to parse start height of block range: %s", args[1])
	}
	if len(args) == 2 {
		endHeight = startHeight
		return
	}
	if strings.EqualFold(strings.TrimSpace(args[2]), "latest") {
		latestBlock := cli.GetLatestBlock(host)
		endHeight = latestBlock.Height
		return
	}
	endHeight, err = strconv.ParseUint(args[2], 10, 64)
	if err != nil {
		cli.Exitf(1, "Failed to parse end height of block range: %s", args[2])
	}
	if endHeight < startHeight {
		cli.Exitf(1, "Cannot have end height (%d) of block range less that start height (%d)", endHeight, startHeight)
	}
	return
}
