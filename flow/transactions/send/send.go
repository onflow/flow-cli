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

	"github.com/onflow/flow-cli/flow/cli"
	"github.com/onflow/flow-go-sdk"
	"github.com/psiemens/sconfig"
	"github.com/spf13/cobra"
)

type Flags struct {
	Signer  string `default:"service" flag:"signer,s"`
	Code    string `flag:"code,c" info:"path to Cadence file"`
	Host    string `flag:"host" info:"Flow Access API host address"`
	Results bool   `default:"false" flag:"results" info:"Display the results of the transaction"`
}

var flags Flags

var Cmd = &cobra.Command{
	Use:   "send",
	Short: "Send a transaction",
	Run: func(cmd *cobra.Command, args []string) {
		project := cli.LoadProject()
		if project == nil {
			return
		}

		signerAccount := project.GetAccountByName(flags.Signer)
		//validateKeyPreReq(signerAccount)
		var (
			code []byte
			err  error
		)

		if flags.Code != "" {
			code, err = ioutil.ReadFile(flags.Code)
			if err != nil {
				cli.Exitf(1, "Failed to read transaction script from %s", flags.Code)
			}
		}

		tx := flow.NewTransaction().
			SetScript(code).
			AddAuthorizer(signerAccount.Address())

		cli.SendTransaction(project.HostWithOverride(flags.Host), signerAccount, tx, flags.Results)
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

// TODO:
/*
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
*/
