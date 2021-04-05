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

package emulator

import (
	"strings"

	emulator "github.com/onflow/flow-emulator"

	"github.com/onflow/flow-emulator/cmd/emulator/start"
	"github.com/onflow/flow-go-sdk/crypto"
	"github.com/spf13/cobra"

	"github.com/onflow/flow-cli/pkg/flowcli/config"
	"github.com/onflow/flow-cli/pkg/flowcli/project"
	"github.com/onflow/flow-cli/pkg/flowcli/util"
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
	if init {
		if sigAlgo == crypto.UnknownSignatureAlgorithm {
			sigAlgo = emulator.DefaultServiceKeySigAlgo
		}

		if hashAlgo == crypto.UnknownHashAlgorithm {
			hashAlgo = emulator.DefaultServiceKeyHashAlgo
		}

		p, err := project.Init(sigAlgo, hashAlgo)
		if err != nil {
			util.Exitf(1, err.Error())
		} else {
			err = p.Save(project.DefaultConfigPath)
			if err != nil {
				util.Exitf(1, err.Error())
			}
		}
	}

	proj, err := project.Load(util.ConfigPath)
	if err != nil {
		if strings.Contains(err.Error(), "project config file does not exist") {
			util.Exitf(1, "🙏 Configuration is missing, initialize it with: 'flow project init' and then rerun this command.")
		} else {
			util.Exitf(1, err.Error())
		}
	}

	serviceAccount, _ := proj.EmulatorServiceAccount()

	serviceKeyHex, ok := serviceAccount.DefaultKey().(*project.HexAccountKey)
	if !ok {
		util.Exit(1, "Only hexadecimal keys can be used as the emulator service account key.")
	}

	privateKey, err := crypto.DecodePrivateKeyHex(serviceKeyHex.SigAlgo(), serviceKeyHex.PrivateKeyHex())
	if err != nil {
		util.Exitf(
			1,
			"Invalid private key in \"%s\" emulator configuration",
			config.DefaultEmulatorConfigName,
		)
	}

	return privateKey, serviceKeyHex.SigAlgo(), serviceKeyHex.HashAlgo()
}

func init() {
	Cmd = start.Cmd(configuredServiceKey)
	Cmd.Use = "emulator"
}
