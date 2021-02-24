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

package kms

import (
	"log"

	"github.com/psiemens/sconfig"
	"github.com/spf13/cobra"

	"github.com/onflow/flow-cli/flow/cli"
	"github.com/onflow/flow-cli/flow/cli/keys"
	"github.com/onflow/flow-cli/flow/config"
	"github.com/onflow/flow-go-sdk"
	"github.com/onflow/flow-go-sdk/crypto"
)

type Flags struct {
	Name       string `flag:"name" info:"name of the key"`
	Address    string `flag:"address" info:"flow address of the account"`
	SigAlgo    string `flag:"sigalgo" info:"signature algorithm for the key"`
	HashAlgo   string `flag:"hashalgo" info:"hash algorithm for the key"`
	KeyIndex   int    `flag:"index" info:"index of the key on the account"`
	KeyContext string `flag:"context" info:"projects/<PROJECTID>/locations/<LOCATION>/keyRings/<KEYRINGID>/cryptoKeys/<KEYID>/cryptoKeyVersions/<KEYVERSION>"`
	Overwrite  bool   `flag:"overwrite" info:"bool indicating if we should overwrite an existing config with the same name in the config file"`
}

var flags Flags

var Cmd = &cobra.Command{
	Use:     "kms",
	Short:   "Save a KMS key to the config file",
	Example: "flow keys save kms --name test --address 8c5303eaa26202d6 --sigalgo ECDSA_secp256k1 --hashalgo SHA2_256 --index 0 --context 'KMS_RESOURCE_ID'",
	Run: func(cmd *cobra.Command, args []string) {
		project := cli.LoadProject()

		account := project.GetAccountByName(flags.Name)
		if account != nil && !flags.Overwrite {
			cli.Exitf(1, "%s already exists in the config, and overwrite is false", flags.Name)
		}

		keyContext, err := keys.KeyContextFromKMSResourceID(flags.KeyContext)
		if err != nil {
			cli.Exitf(1, "key context could not be parsed %s", flags.KeyContext)
		}

		address := flow.HexToAddress(flags.Address)
		if !address.IsValid(flow.Emulator) { // TODO: don't hardcode this
			cli.Exitf(1, "key address not valid")
		}

		keys := []config.AccountKey{{
			Type:     config.KeyTypeKMS,
			Index:    flags.KeyIndex,
			SigAlgo:  crypto.StringToSignatureAlgorithm(flags.SigAlgo),
			HashAlgo: crypto.StringToHashAlgorithm(flags.HashAlgo),
			Context:  keyContext,
		}}

		account, err := cli.AccountFromConfig(
			config.Account{
				Name:    flags.Name,
				Address: address,
				ChainID: flow.Emulator, // TODO: don't hardcode this
				Keys:    keys,
			},
		)

		/* TODO: discuss how this changed
		// Validate account
		err = account.LoadSigner()
		if err != nil {
			cli.Exitf(1, "provide key context could not be loaded as a valid signer %s", flags.KeyContext)
		}
		*/

		project.AddAccount(account)
		project.Save() // TODO: handle error
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
