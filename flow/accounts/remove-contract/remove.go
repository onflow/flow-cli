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

package remove_contract

import (
	"log"

	"github.com/onflow/flow-go-sdk/templates"
	"github.com/psiemens/sconfig"
	"github.com/spf13/cobra"

	cli "github.com/onflow/flow-cli/flow"
)

type Config struct {
	Signer  string `default:"service" flag:"signer,s"`
	Host    string `flag:"host" info:"Flow Access API host address"`
	Results bool   `default:"false" flag:"results" info:"Display the results of the transaction"`
}

var conf Config

var Cmd = &cobra.Command{
	Use:   "remove-contract <name>",
	Short: "Remove a contract deployed to an account",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		projectConf := cli.LoadConfig()

		contractName := args[0]

		signerAccount := projectConf.Accounts[conf.Signer]

		tx := templates.RemoveAccountContract(signerAccount.Address, contractName)

		cli.SendTransaction(
			projectConf.HostWithOverride(conf.Host),
			signerAccount,
			tx,
			conf.Results,
		)
	},
}

func init() {
	initConfig()
}

func initConfig() {
	err := sconfig.New(&conf).
		FromEnvironment(cli.EnvPrefix).
		BindFlags(Cmd.PersistentFlags()).
		Parse()
	if err != nil {
		log.Fatal(err)
	}
}
