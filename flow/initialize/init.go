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

package initialize

import (
	"fmt"
	"log"

	"github.com/onflow/flow-go-sdk/crypto"
	"github.com/psiemens/sconfig"
	"github.com/spf13/cobra"

	cli "github.com/onflow/flow-cli/flow"
)

type Config struct {
	ServicePrivateKey  string `flag:"service-priv-key" info:"Service account private key"`
	ServiceKeySigAlgo  string `default:"ECDSA_P256" flag:"service-sig-algo" info:"Service account key signature algorithm"`
	ServiceKeyHashAlgo string `default:"SHA3_256" flag:"service-hash-algo" info:"Service account key hash algorithm"`
	Reset              bool   `default:"false" flag:"reset" info:"Reset flow.json config file"`
}

var (
	conf Config
)

var Cmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize a new account profile",
	Run: func(cmd *cobra.Command, args []string) {
		if !cli.ConfigExists() || conf.Reset {
			var pconf *cli.Config
			serviceKeySigAlgo := crypto.StringToSignatureAlgorithm(conf.ServiceKeySigAlgo)
			serviceKeyHashAlgo := crypto.StringToHashAlgorithm(conf.ServiceKeyHashAlgo)
			if len(conf.ServicePrivateKey) > 0 {
				serviceKey := cli.MustDecodePrivateKeyHex(serviceKeySigAlgo, conf.ServicePrivateKey)
				pconf = InitProjectWithServiceKey(serviceKey, serviceKeyHashAlgo)
			} else {
				pconf = InitProject(serviceKeySigAlgo, serviceKeyHashAlgo)
			}
			serviceAcct := pconf.ServiceAccount()

			fmt.Printf("⚙️   Flow client initialized with service account:\n\n")
			fmt.Printf("👤  Address: 0x%x\n", serviceAcct.Address.Bytes())
			fmt.Printf("ℹ️   Start the emulator with this service account by running: flow emulator start\n")
		} else {
			fmt.Printf("⚠️   Flow configuration file already exists! Begin by running: flow emulator start\n")
		}
	},
}

// InitProject generates a new service key and saves project config.
func InitProject(sigAlgo crypto.SignatureAlgorithm, hashAlgo crypto.HashAlgorithm) *cli.Config {
	seed := cli.RandomSeed(crypto.MinSeedLength)

	serviceKey, err := crypto.GeneratePrivateKey(sigAlgo, seed)
	if err != nil {
		cli.Exitf(1, "Failed to generate private key: %v", err)
	}

	return InitProjectWithServiceKey(serviceKey, hashAlgo)
}

// InitProjectWithServiceKey creates and saves a new project config
// using the specified service key.
func InitProjectWithServiceKey(privateKey crypto.PrivateKey, hashAlgo crypto.HashAlgorithm) *cli.Config {
	pconf := cli.NewConfig()
	pconf.SetServiceAccountKey(privateKey, hashAlgo)
	cli.MustSaveConfig(pconf)
	return pconf
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
