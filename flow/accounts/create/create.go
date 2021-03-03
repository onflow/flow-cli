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

package create

import (
	"io/ioutil"
	"log"
	"strings"

	"github.com/onflow/flow-go-sdk"
	"github.com/onflow/flow-go-sdk/crypto"
	"github.com/onflow/flow-go-sdk/templates"
	"github.com/psiemens/sconfig"
	"github.com/spf13/cobra"

	"github.com/onflow/flow-cli/flow/cli"
)

type Flags struct {
	Signer    string   `default:"service" flag:"signer,s"`
	Keys      []string `flag:"key,k" info:"Public keys to attach to account"`
	SigAlgo   string   `default:"ECDSA_P256" flag:"sig-algo" info:"Signature algorithm used to generate the keys"`
	HashAlgo  string   `default:"SHA3_256" flag:"hash-algo" info:"Hash used for the digest"`
	Host      string   `flag:"host" info:"Flow Access API host address"`
	Results   bool     `default:"false" flag:"results" info:"Display the results of the transaction"`
	Contracts []string `flag:"contract,c" info:"Contract to be deployed during account creation. <name:path>"`
}

var flags Flags

var Cmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new account",
	Run: func(cmd *cobra.Command, args []string) {
		project, _ := cli.LoadProject(cli.ConfigPath)

		host := flags.Host
		if host == "" {
			host = project.DefaultHost("")
		}

		signerAccount := project.GetAccountByName(flags.Signer)

		accountKeys := make([]*flow.AccountKey, len(flags.Keys))

		sigAlgo := crypto.StringToSignatureAlgorithm(flags.SigAlgo)
		if sigAlgo == crypto.UnknownSignatureAlgorithm {
			cli.Exitf(1, "Failed to determine signature algorithm from %s", flags.SigAlgo)
		}
		hashAlgo := crypto.StringToHashAlgorithm(flags.HashAlgo)
		if hashAlgo == crypto.UnknownHashAlgorithm {
			cli.Exitf(1, "Failed to determine hash algorithm from %s", flags.HashAlgo)
		}

		for i, publicKeyHex := range flags.Keys {
			publicKey := cli.MustDecodePublicKeyHex(cli.DefaultSigAlgo, publicKeyHex)
			accountKeys[i] = &flow.AccountKey{
				PublicKey: publicKey,
				SigAlgo:   sigAlgo,
				HashAlgo:  hashAlgo,
				Weight:    flow.AccountKeyWeightThreshold,
			}
		}

		contracts := []templates.Contract{}

		for _, contract := range flags.Contracts {
			contractFlagContent := strings.SplitN(contract, ":", 2)
			if len(contractFlagContent) != 2 {
				cli.Exitf(1, "Failed to read contract name and path from flag. Ensure you're providing a contract name and a file path. %s", contract)
			}
			contractName := contractFlagContent[0]
			contractPath := contractFlagContent[1]
			contractSource, err := ioutil.ReadFile(contractPath)
			if err != nil {
				cli.Exitf(1, "Failed to read contract from source file %s", contractPath)
			}
			contracts = append(contracts,
				templates.Contract{
					Name:   contractName,
					Source: string(contractSource),
				},
			)
		}

		tx := templates.CreateAccount(accountKeys, contracts, signerAccount.Address())

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
