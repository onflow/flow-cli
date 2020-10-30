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

package emulator

import (
	"fmt"

	emulator "github.com/dapperlabs/flow-emulator"
	"github.com/dapperlabs/flow-emulator/cmd/emulator/start"
	"github.com/onflow/flow-go-sdk/crypto"
	"github.com/spf13/cobra"

	cli "github.com/dapperlabs/flow-cli/flow"
	"github.com/dapperlabs/flow-cli/flow/initialize"
)

var Cmd = &cobra.Command{
	Use:              "emulator",
	Short:            "Flow emulator server",
	TraverseChildren: true,
}

func configuredServiceKey(
	init bool,
	sigAlgo crypto.SignatureAlgorithm,
	hashAlgo crypto.HashAlgorithm,
) (
	crypto.PrivateKey,
	crypto.SignatureAlgorithm,
	crypto.HashAlgorithm,
) {
	if sigAlgo == crypto.UnknownSignatureAlgorithm {
		sigAlgo = emulator.DefaultServiceKeySigAlgo
	}

	if hashAlgo == crypto.UnknownHashAlgorithm {
		hashAlgo = emulator.DefaultServiceKeyHashAlgo
	}

	var serviceAcct *cli.Account

	if init {
		pconf := initialize.InitProject(sigAlgo, hashAlgo)
		serviceAcct = pconf.ServiceAccount()

		fmt.Printf("⚙️   Flow client initialized with service account:\n\n")
		fmt.Printf("👤  Address: 0x%s\n", serviceAcct.Address)
	} else {
		serviceAcct = cli.LoadConfig().ServiceAccount()
	}

	return serviceAcct.PrivateKey, serviceAcct.SigAlgo, serviceAcct.HashAlgo
}

func init() {
	Cmd.AddCommand(start.Cmd(configuredServiceKey))
}
