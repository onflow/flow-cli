/*
 * Flow CLI
 *
 * Copyright 2019 Dapper Labs, Inc.
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
	"errors"
	"fmt"
	"os"

	"github.com/spf13/afero"

	"github.com/onflow/flow-cli/pkg/flowkit"

	"github.com/onflow/flow-cli/internal/command"

	emulator "github.com/onflow/flow-emulator"

	"github.com/onflow/flow-emulator/cmd/emulator/start"
	"github.com/onflow/flow-go-sdk/crypto"
	"github.com/spf13/cobra"

	"github.com/onflow/flow-cli/pkg/flowkit/config"
	"github.com/onflow/flow-cli/pkg/flowkit/util"
)

var Cmd *cobra.Command

func ConfiguredServiceKey(
	init bool,
	sigAlgo crypto.SignatureAlgorithm,
	hashAlgo crypto.HashAlgorithm,
) (
	crypto.PrivateKey,
	crypto.SignatureAlgorithm,
	crypto.HashAlgorithm,
) {
	var state *flowkit.State
	var err error
	loader := &afero.Afero{Fs: afero.NewOsFs()}

	if init {
		if sigAlgo == crypto.UnknownSignatureAlgorithm {
			sigAlgo = emulator.DefaultServiceKeySigAlgo
		}

		if hashAlgo == crypto.UnknownHashAlgorithm {
			hashAlgo = emulator.DefaultServiceKeyHashAlgo
		}

		state, err = flowkit.Init(loader, sigAlgo, hashAlgo)
		if err != nil {
			Exitf(1, err.Error())
		} else {
			err = state.SaveDefault()
			if err != nil {
				Exitf(1, err.Error())
			}
		}
	} else {
		state, err = flowkit.Load(command.Flags.ConfigPaths, loader)
		if err != nil {
			if errors.Is(err, config.ErrDoesNotExist) {
				Exitf(1, "🙏 Configuration is missing, initialize it with: 'flow init' and then rerun this command.")
			} else {
				Exitf(1, err.Error())
			}
		}
	}

	serviceAccount, err := state.EmulatorServiceAccount()
	if err != nil {
		util.Exit(1, err.Error())
	}

	privateKey, err := serviceAccount.Key().PrivateKey()
	if err != nil {
		util.Exit(1, "Only hexadecimal keys can be used as the emulator service account key.")
	}

	err = serviceAccount.Key().Validate()
	if err != nil {
		util.Exit(
			1,
			fmt.Sprintf("invalid private key in %s emulator configuration, %s",
				serviceAccount.Name(),
				err.Error(),
			),
		)
	}

	return *privateKey, serviceAccount.Key().SigAlgo(), serviceAccount.Key().HashAlgo()
}

func init() {
	Cmd = start.Cmd(ConfiguredServiceKey)
	Cmd.Use = "emulator"
}

func Exitf(code int, msg string, args ...interface{}) {
	fmt.Printf(msg+"\n", args...)
	os.Exit(code)
}
