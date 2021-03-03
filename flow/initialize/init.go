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

package initialize

import (
	"fmt"
	"log"

	"github.com/onflow/flow-cli/flow/cli"
	"github.com/onflow/flow-go-sdk/crypto"
	"github.com/psiemens/sconfig"
	"github.com/spf13/cobra"
)

type Flags struct {
	ServicePrivateKey  string `flag:"service-priv-key" info:"Service account private key"`
	ServiceKeySigAlgo  string `default:"ECDSA_P256" flag:"service-sig-algo" info:"Service account key signature algorithm"`
	ServiceKeyHashAlgo string `default:"SHA3_256" flag:"service-hash-algo" info:"Service account key hash algorithm"`
	Reset              bool   `default:"false" flag:"reset" info:"Reset flow.json config file"`
}

var (
	flags Flags
)

var Cmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize a new account profile",
	Run: func(cmd *cobra.Command, args []string) {

		if !cli.ProjectExists() || flags.Reset {
			serviceKeySigAlgo := crypto.StringToSignatureAlgorithm(flags.ServiceKeySigAlgo)
			serviceKeyHashAlgo := crypto.StringToHashAlgorithm(flags.ServiceKeyHashAlgo)

			project := cli.InitProject(serviceKeySigAlgo, serviceKeyHashAlgo)

			if len(flags.ServicePrivateKey) > 0 {
				serviceKey := cli.MustDecodePrivateKeyHex(serviceKeySigAlgo, flags.ServicePrivateKey)
				project.SetEmulatorServiceKey(serviceKey)
			}

			project.Save()

			serviceAcct, _ := project.EmulatorServiceAccount()

			fmt.Printf("⚙️   Flow client initialized with service account:\n\n")
			fmt.Printf("👤  Address: 0x%s\n", serviceAcct.Address)
			fmt.Printf(
				"Start the Flow Emulator by running: %s\n",
				cli.Bold("flow project start-emulator"),
			)
		} else {
			fmt.Printf(
				cli.Red(fmt.Sprintf("A Flow project already exists in %s \n", cli.DefaultConfigPath)),
			)
			fmt.Printf(
				"Start the Flow Emulator by running: %s\n%s\n",
				cli.Bold("flow project start-emulator"),
				"Or reset config using --reset flag.",
			)
		}
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
