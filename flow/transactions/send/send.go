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

package send

import (
	"log"
	"os"

	"github.com/onflow/flow-go-sdk"
	"github.com/psiemens/sconfig"
	"github.com/spf13/cobra"

	cli "github.com/onflow/flow-cli/flow"
	"github.com/onflow/flow-cli/flow/transactions/utils"
)

type Config struct {
	Args    string `default:"" flag:"args" info:"arguments in JSON-Cadence format"`
	Code    string `flag:"code,c" info:"path to Cadence file"`
	Payload string `flag:"payload" info:"path to Transaction Payload file"`
	Host    string `flag:"host" info:"Flow Access API host address"`
	Signer  string `default:"service" flag:"signer,s"`
	Results bool   `default:"false" flag:"results" info:"Display the results of the transaction"`
}

var conf Config

var Cmd = &cobra.Command{
	Use:     "send",
	Short:   "Send a transaction",
	Example: `flow transactions send --code=tx.cdc --args="[{\"type\": \"String\", \"value\": \"Hello, Cadence\"}]"`,
	Run: func(cmd *cobra.Command, args []string) {
		projectConf := cli.LoadConfig()

		signerAccount := projectConf.Accounts[conf.Signer]
		// TODO: Remove once new configuration is migrated
		if signerAccount == nil && conf.Signer == "service" {
			signerAccount = projectConf.Accounts["emulator-account"]
		}

		validateKeyPreReq(signerAccount)
		var tx *flow.Transaction

		if conf.Payload != "" && conf.Code != "" {
			cli.Exitf(1, "Both a partial transaction and Cadence code file provided, but cannot use both")
		} else if conf.Payload != "" {
			utils.AssertEmptyOnLoadingPayload(conf.Args, "arguments")

			tx = utils.LoadTransactionPayloadFromFile(conf.Payload)
		} else {
			tx = utils.NewTransactionWithCodeArgsAuthorizers(conf.Code, conf.Args, []string{signerAccount.Address.String()})
			tx = cli.PrepareTransaction(projectConf.HostWithOverride(conf.Host), signerAccount, tx, signerAccount.Address)
		}

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

func validateKeyPreReq(account *cli.Account) {
	if account.KeyType == cli.KeyTypeHex {
		// Always Valid
		return
	} else if account.KeyType == cli.KeyTypeKMS {
		// Check GOOGLE_APPLICATION_CREDENTIALS
		googleAppCreds := os.Getenv("GOOGLE_APPLICATION_CREDENTIALS")
		if len(googleAppCreds) == 0 {
			if len(account.KeyContext["projectId"]) == 0 {
				cli.Exitf(1, "Could not get GOOGLE_APPLICATION_CREDENTIALS, no google service account json provided but private key type is KMS", account.Address)
			}
			cli.GcloudApplicationSignin(account.KeyContext["projectId"])
		}
		return
	}
	cli.Exitf(1, "Failed to validate %s key for %s", account.KeyType, account.Address)

}
