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
	Host string `flag:"host" info:"Flow Access API host address"`
}

var flags Flags

var Cmd = &cobra.Command{
	Use:   "get <collection_id>",
	Short: "Get collection info",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		host, err := cli.LoadHostForNetwork(flags.Host, "")
		if err != nil {
			cli.Exitf(1, err.Error())
		}

		collectionID := flow.HexToID(args[0])
		collection := cli.GetCollectionByID(host, collectionID)
		printCollection(collection)
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

func printCollection(collection *flow.Collection) {
	fmt.Println()
	fmt.Println("Collection ID: ", collection.ID())
	for i, transaction := range collection.TransactionIDs {
		fmt.Printf("  Transaction %d: %s\n", i, transaction)
	}
	fmt.Println()
}
