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

package send

import (
	"io/ioutil"
	"log"
	"os"

	"github.com/onflow/flow-cli/flow/cli"
	"github.com/onflow/flow-cli/flow/config"
	"github.com/onflow/flow-go-sdk"
	"github.com/psiemens/sconfig"
	"github.com/spf13/cobra"
)

type Flags struct {
	Args    string `default:"" flag:"args" info:"arguments in JSON-Cadence format"`
	Code    string `flag:"code,c" info:"path to Cadence file"`
	Host    string `flag:"host" info:"Flow Access API host address"`
	Signer  string `default:"emulator-account" flag:"signer,s"`
	Results bool   `default:"false" flag:"results" info:"Display the results of the transaction"`
}

var flags Flags

var Cmd = &cobra.Command{
	Use:     "send",
	Short:   "Send a transaction",
	Example: `flow transactions send --code=tx.cdc --args="[{\"type\": \"String\", \"value\": \"Hello, Cadence\"}]"`,
	Run: func(cmd *cobra.Command, args []string) {
		project, err := cli.LoadProject(cli.ConfigPath)
		if err != nil {
			cli.Exitf(1, err.Error())
		}

		host := flags.Host
		if host == "" {
			host = project.DefaultHost("")
		}

		signerAccount := project.GetAccountByName(flags.Signer)
		if signerAccount == nil {
			cli.Exitf(1, "Account %s not found. Check that account name matches an account defined in configuration.", flags.Signer)
		}

		validateKeyPreReq(signerAccount)

		var code []byte
		if flags.Code != "" {
			code, err = ioutil.ReadFile(flags.Code)
			if err != nil {
				cli.Exitf(1, "Failed to read transaction script from %s", flags.Code)
			}
		}

		tx := flow.NewTransaction().
			SetScript(code).
			AddAuthorizer(signerAccount.Address())

		// Arguments
		if flags.Args != "" {
			transactionArguments, err := cli.ParseArguments(flags.Args)
			if err != nil {
				cli.Exitf(1, "Invalid arguments passed: %s", flags.Args)
			}

			for _, arg := range transactionArguments {
				err := tx.AddArgument(arg)

				if err != nil {
					cli.Exitf(1, "Failed to add %s argument to a transaction ", flags.Code)
				}
			}
		}

		cli.SendTransaction(host, signerAccount, tx, flags.Results)
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

func validateKeyPreReq(account *cli.Account) {
	if account.DefaultKey().Type() == config.KeyTypeHex {
		// Always Valid
		return // TODO: check difference between googleKMS and KMS
	} else if account.DefaultKey().Type() == config.KeyTypeGoogleKMS {
		// Check GOOGLE_APPLICATION_CREDENTIALS
		googleAppCreds := os.Getenv("GOOGLE_APPLICATION_CREDENTIALS")
		if len(googleAppCreds) == 0 {
			if len(account.DefaultKey().ToConfig().Context["projectId"]) == 0 {
				cli.Exitf(1, "Could not get GOOGLE_APPLICATION_CREDENTIALS, no google service account json provided but private key type is KMS", account.Address)
			}
			cli.GcloudApplicationSignin(account)
		}
		return
	}

	cli.Exitf(1, "Failed to validate %s key for %s", account.DefaultKey().Type(), account.Address)
}
