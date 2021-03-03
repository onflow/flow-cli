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
	"fmt"
	"log"

	"github.com/onflow/flow-cli/flow/cli"
	"github.com/onflow/flow-go-sdk"
	"github.com/psiemens/sconfig"
	"github.com/spf13/cobra"
)

type Flags struct {
	Host        string `flag:"host" info:"Flow Access API host address"`
	Latest      bool   `default:"false" flag:"latest" info:"Display latest block"`
	BlockID     string `default:"" flag:"id" info:"Display block by id"`
	BlockHeight uint64 `default:"0" flag:"height" info:"Display block by height"`
	Events      string `default:"" flag:"events" info:"List events of this type for the block"`
	Verbose     bool   `default:"false" flag:"verbose" info:"Display transactions in block"`
}

var flags Flags

var Cmd = &cobra.Command{
	Use:   "get <block_id>",
	Short: "Get block info",
	Run: func(cmd *cobra.Command, args []string) {
		var block *flow.Block
		project, _ := cli.LoadProject(cli.ConfigPath)

		host := flags.Host
		if host == "" {
			host = project.DefaultHost("")
		}

		if flags.Latest {
			block = cli.GetLatestBlock(host)
		} else if len(flags.BlockID) > 0 {
			blockID := flow.HexToID(flags.BlockID)
			block = cli.GetBlockByID(host, blockID)
		} else if len(args) > 0 && len(args[0]) > 0 {
			blockID := flow.HexToID(args[0])
			block = cli.GetBlockByID(host, blockID)
		} else {
			block = cli.GetBlockByHeight(host, flags.BlockHeight)
		}
		printBlock(block, flags.Verbose)
		if flags.Events != "" {
			cli.GetBlockEvents(host, block.Height, block.Height, flags.Events, true)
		}
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

func printBlock(block *flow.Block, verbose bool) {
	fmt.Println()
	fmt.Println("Block ID: ", block.ID)
	fmt.Println("Parent ID: ", block.ParentID)
	fmt.Println("Height: ", block.Height)
	fmt.Println("Timestamp: ", block.Timestamp)
	fmt.Println("Total Collections: ", len(block.CollectionGuarantees))
	for i, guarantee := range block.CollectionGuarantees {
		fmt.Printf("  Collection %d: %s\n", i, guarantee.CollectionID)
		if verbose {
			collection := cli.GetCollectionByID(flags.Host, guarantee.CollectionID)
			for i, transaction := range collection.TransactionIDs {
				fmt.Printf("    Transaction %d: %s\n", i, transaction)
			}
		}
	}
	fmt.Println("Total Seals: ", len(block.Seals))
	fmt.Println()
}
