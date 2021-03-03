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

package get

import (
	"fmt"
	"log"
	"strings"

	"github.com/onflow/flow-cli/flow/cli"
	"github.com/onflow/flow-go-sdk"
	"github.com/psiemens/sconfig"
	"github.com/spf13/cobra"

	"github.com/onflow/cadence"
)

type Flags struct {
	Host string `flag:"host" info:"Flow Access API host address"`
	Code bool   `default:"false" flag:"code" info:"Display code deployed to the account"`
}

var flag Flags

var Cmd = &cobra.Command{
	Use:   "get <address>",
	Short: "Get account info",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		project := cli.LoadProject()
		if project == nil {
			return
		}

		address := flow.HexToAddress(
			strings.ReplaceAll(args[0], "0x", ""),
		)

		account := cli.GetAccount(
			project.HostWithOverride(flag.Host),
			address,
		)

		printAccount(account, flag.Code)
	},
}

func init() {
	initConfig()
}

func initConfig() {
	err := sconfig.New(&flag).
		FromEnvironment(cli.EnvPrefix).
		BindFlags(Cmd.PersistentFlags()).
		Parse()
	if err != nil {
		log.Fatal(err)
	}
}

func printAccount(account *flow.Account, printCode bool) {
	fmt.Println()
	fmt.Println("Address: " + account.Address.Hex())
	fmt.Println("Balance: ", cadence.UFix64(account.Balance))
	fmt.Println("Total Keys: ", len(account.Keys))

	for _, key := range account.Keys {
		fmt.Println("  ---")
		fmt.Println("  Key Index: ", key.Index)
		fmt.Printf("  PublicKey: %x\n", key.PublicKey.Encode())
		fmt.Println("  SigAlgo: ", key.SigAlgo)
		fmt.Println("  HashAlgo: ", key.HashAlgo)
		fmt.Println("  Weight: ", key.Weight)
		fmt.Println("  SequenceNumber: ", key.SequenceNumber)
		fmt.Println("  Revoked: ", key.Revoked)
	}

	fmt.Println("  ---")

	if printCode {
		for name, code := range account.Contracts {
			fmt.Printf("Code '%s':\n", name)
			fmt.Println(string(code))
		}
	}

	fmt.Println()
}
