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

package kms

import (
	"log"

	"github.com/onflow/flow-go-sdk"
	"github.com/onflow/flow-go-sdk/crypto"
	"github.com/psiemens/sconfig"
	"github.com/spf13/cobra"

	cli "github.com/onflow/flow-cli/flow"
)

type Config struct {
	Name       string `flag:"name" info:"name of the key"`
	Address    string `flag:"address" info:"flow address of the account"`
	SigAlgo    string `flag:"sigalgo" info:"signature algorithm for the key"`
	HashAlgo   string `flag:"hashalgo" info:"hash algorithm for the key"`
	KeyIndex   int    `flag:"index" info:"index of the key on the account"`
	KeyContext string `flag:"context" info:"projects/<PROJECTID>/locations/<LOCATION>/keyRings/<KEYRINGID>/cryptoKeys/<KEYID>/cryptoKeyVersions/<KEYVERSION>"`
	Overwrite  bool   `flag:"overwrite" info:"bool indicating if we should overwrite an existing config with the same name in the config file"`
}

var conf Config

var Cmd = &cobra.Command{
	Use:     "kms",
	Short:   "Save a KMS key to the config file",
	Example: "flow keys save kms --name test --address 8c5303eaa26202d6 --sigalgo ECDSA_secp256k1 --hashalgo SHA2_256 --index 0 --context 'KMS_RESOURCE_ID'",
	Run: func(cmd *cobra.Command, args []string) {
		projectConf := cli.LoadConfig()

		_, accountExists := projectConf.Accounts[conf.Name]
		if accountExists && !conf.Overwrite {
			cli.Exitf(1, "%s already exists in the config, and overwrite is false", conf.Name)
		}

		keyContext, err := cli.KeyContextFromKMSResourceID(conf.KeyContext)
		if err != nil {
			cli.Exitf(1, "key context could not be parsed %s", conf.KeyContext)
		}
		// Populate account
		account := &cli.Account{
			KeyType:    cli.KeyTypeKMS,
			Address:    flow.HexToAddress(conf.Address),
			SigAlgo:    crypto.StringToSignatureAlgorithm(conf.SigAlgo),
			HashAlgo:   crypto.StringToHashAlgorithm(conf.HashAlgo),
			KeyIndex:   conf.KeyIndex,
			KeyContext: keyContext,
		}

		// Validate account
		err = account.LoadSigner()
		if err != nil {
			cli.Exitf(1, "provide key context could not be loaded as a valid signer %s", conf.KeyContext)
		}

		projectConf.Accounts[conf.Name] = account
		err = cli.SaveConfig(projectConf)
		if err != nil {
			cli.Exitf(1, "could not save config file %s", cli.ConfigPath)
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
