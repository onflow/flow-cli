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

package hex

import (
	"log"

	"github.com/onflow/flow-cli/flow/cli"
	"github.com/psiemens/sconfig"
	"github.com/spf13/cobra"
)

type Config struct {
	Name      string `flag:"name" info:"name of the key"`
	Address   string `flag:"address" info:"flow address of the account"`
	SigAlgo   string `flag:"sigalgo" info:"signature algorithm for the key"`
	HashAlgo  string `flag:"hashalgo" info:"hash algorithm for the key"`
	KeyIndex  int    `flag:"index" info:"index of the key on the account"`
	KeyHex    string `flag:"privatekey" info:"private key in hex format"`
	Overwrite bool   `flag:"overwrite" info:"bool indicating if we should overwrite an existing config with the same name in the config file"`
}

var conf Config

var Cmd = &cobra.Command{
	Use:     "hex",
	Short:   "Save a hex key to the config file",
	Example: "flow keys save hex --name test --address 8c5303eaa26202d6 --sigalgo ECDSA_secp256k1 --hashalgo SHA2_256 --index 0 --privatekey <HEX_PRIVATEKEY>",
	Run: func(cmd *cobra.Command, args []string) {
		project := cli.LoadProject()
		if project == nil {
			return
		}

		if conf.Name == "" {
			cli.Exitf(1, "missing name")
		}
		// TODO: implement
		// Populate account
		/*

			accountExists := project.GetAccountByName(conf.Name)
			if accountExists && !conf.Overwrite {
				cli.Exitf(1, "%s already exists in the config, and overwrite is false", conf.Name)
			}

			// Parse address
			decodedAddress, err := hex.DecodeString(conf.Address)
			if err != nil {
				cli.Exitf(1, "invalid address: %s", err.Error())
			}
			address := flow.BytesToAddress(decodedAddress)

			// Parse signature algorithm
			if conf.SigAlgo == "" {
				cli.Exitf(1, "missing signature algorithm")
			}

			algorithm := crypto.StringToSignatureAlgorithm(conf.SigAlgo)
			if algorithm == crypto.UnknownSignatureAlgorithm {
				cli.Exitf(1, "invalid signature algorithm")
			}

			// Parse hash algorithm

			if conf.HashAlgo == "" {
				cli.Exitf(1, "missing hash algorithm")
			}

			hashAlgorithm := crypto.StringToHashAlgorithm(conf.HashAlgo)
			if hashAlgorithm == crypto.UnknownHashAlgorithm {
				cli.Exitf(1, "invalid hash algorithm")
			}


					account := &cli.Account{
						KeyType:    cli.KeyTypeHex,
						Address:    address,
						SigAlgo:    algorithm,
						HashAlgo:   hashAlgorithm,
						KeyIndex:   conf.KeyIndex,
						KeyContext: map[string]string{"privateKey": conf.KeyHex},
					}
					privateKey, err := crypto.DecodePrivateKeyHex(account.SigAlgo, conf.KeyHex)
					if err != nil {
						cli.Exitf(1, "key hex could not be parsed")
					}

					account.PrivateKey = privateKey

					// Validate account
					err = account.LoadSigner()
					if err != nil {
						cli.Exitf(1, "provide key could not be loaded as a valid signer %s", conf.KeyHex)
					}

					project.AddAccountByName(conf.Name, account)

				accountKey := &config.AccountKey{
					Type:     config.KeyTypeHex,
					SigAlgo:  algorithm,
					HashAlgo: hashAlgorithm,
					Index:    conf.KeyIndex,
					Context:  map[string]string{"privateKey": conf.KeyHex},
				}

				account := &config.Account{
					Address: address,
					Keys:    []config.AccountKey{*accountKey},
				}

				newKey, err := keys.NewAccountKey(*accountKey)
				account.Keys[0] = newKey
		*/

		project.Save() // TODO: handle error
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
