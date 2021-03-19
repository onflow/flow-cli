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

package sign

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"

	"github.com/onflow/flow-go-sdk"
	"github.com/psiemens/sconfig"
	"github.com/spf13/cobra"

	cli "github.com/onflow/flow-cli/flow"
)

type Config struct {
	Args                  string   `default:"" flag:"args" info:"arguments in JSON-Cadence format"`
	Signer                string   `default:"service" flag:"signer,s"`
	Role                  string   `default:"authorizer" flag:"role"`
	AdditionalAuthorizers []string `flag:"additional-authorizers" info:"Additional authorizer addresses to add to the transaction"`
	PayerAddress          string   `flag:"payer" info:"Specify payer of the transaction. Defaults to current signer."`
	Code                  string   `flag:"code,c" info:"path to Cadence file"`
	Host                  string   `flag:"host" info:"Flow Access API host address"`
	Encoding              string   `default:"hexrlp" flag:"encoding" info:"Encoding to use for transactio (rlp)"`
	Output                string   `default:"" flag:"output,o" info:"Output location for transaction file"`
}

var conf Config

var Cmd = &cobra.Command{
	Use:   "sign",
	Short: "Sign a transaction",
	Run: func(cmd *cobra.Command, args []string) {
		projectConf := cli.LoadConfig()

		signerAccount := projectConf.Accounts[conf.Signer]
		validateKeyPreReq(signerAccount)
		var (
			code  []byte
			payer flow.Address
			err   error
		)

		if conf.Code != "" {
			code, err = ioutil.ReadFile(conf.Code)
			if err != nil {
				cli.Exitf(1, "Failed to read transaction script from %s", conf.Code)
			}
		}

		if conf.PayerAddress != "" {
			payer = flow.HexToAddress(conf.PayerAddress)
		} else {
			payer = signerAccount.Address
		}

		tx := flow.NewTransaction().
			SetScript(code)

		// Arguments
		if conf.Args != "" {
			transactionArguments, err := cli.ParseArguments(conf.Args)
			if err != nil {
				cli.Exitf(1, "Invalid arguments passed: %s", conf.Args)
			}

			for _, arg := range transactionArguments {
				err := tx.AddArgument(arg)

				if err != nil {
					cli.Exitf(1, "Failed to add %s argument to a transaction ", conf.Code)
				}
			}
		}

		signerRole := cli.SignerRole(conf.Role)
		switch signerRole {
		case cli.SignerRoleAuthorizer:
			tx.AddAuthorizer(signerAccount.Address)
		case cli.SignerRolePayer:
			if payer != signerAccount.Address {
				cli.Exitf(1, "Role specified as Payer, but Payer address also provided, and different: %s !=", payer, signerAccount.Address)
			}
		case cli.SignerRoleProposer:
			cli.Exitf(1, "Proposer role not yet supported: %s", conf.Role)
		default:
			cli.Exitf(1, "unknown role %s", conf.Role)
		}

		for _, authorizerString := range conf.AdditionalAuthorizers {
			authorizerAddress := flow.HexToAddress(authorizerString)
			tx.AddAuthorizer(authorizerAddress)
		}

		cli.PrepareTransaction(projectConf.HostWithOverride(conf.Host), signerAccount, tx, payer)

		tx = cli.SignTransaction(projectConf.HostWithOverride(conf.Host), signerAccount, signerRole, tx)

		fmt.Printf("%s encoded transaction written to %s\n", conf.Encoding, conf.Output)

		output := fmt.Sprintf("%x", tx.Encode())
		if len(strings.TrimSpace(conf.Output)) == 0 {
			fmt.Println(output)
			return
		}
		err = ioutil.WriteFile(conf.Output, []byte(output), os.ModePerm)
		if err != nil {
			cli.Exitf(1, "Failed to save encoded transaction to file %s", conf.Output)
		}
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
