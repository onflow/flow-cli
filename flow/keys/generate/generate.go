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

package generate

import (
	"encoding/hex"
	"fmt"
	"log"

	"github.com/onflow/flow-go-sdk/crypto"
	"github.com/psiemens/sconfig"
	"github.com/spf13/cobra"

	cli "github.com/onflow/flow-cli/flow"
)

type Config struct {
	Seed    string `flag:"seed,s" info:"Deterministic seed phrase"`
	SigAlgo string `default:"ECDSA_P256" flag:"algo,a" info:"Signature algorithm"`
}

var conf Config

var Cmd = &cobra.Command{
	Use:   "generate",
	Short: "Generate a new key-pair",
	Run: func(cmd *cobra.Command, args []string) {
		var seed []byte
		if conf.Seed == "" {
			seed = cli.RandomSeed(crypto.MinSeedLength)
		} else {
			seed = []byte(conf.Seed)
		}

		sigAlgo := crypto.StringToSignatureAlgorithm(conf.SigAlgo)
		if sigAlgo == crypto.UnknownSignatureAlgorithm {
			cli.Exitf(1, "Invalid signature algorithm: %s", conf.SigAlgo)
		}

		fmt.Printf(
			"Generating key pair with signature algorithm:                 %s\n...\n",
			sigAlgo,
		)

		privateKey, err := crypto.GeneratePrivateKey(sigAlgo, seed)
		if err != nil {
			cli.Exitf(1, "Failed to generate private key: %v", err)
		}

		fmt.Printf(
			"\U0001F510 Private key (\u26A0\uFE0F\u202F store safely and don't share with anyone): %s\n",
			hex.EncodeToString(privateKey.Encode()),
		)
		fmt.Printf(
			"\U0001F54A️\uFE0F\u202F Encoded public key (share freely):                         %x\n",
			privateKey.PublicKey().Encode(),
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
