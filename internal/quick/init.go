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

package quick

import (
	"fmt"

	"github.com/onflow/flow-go-sdk/crypto"
	"github.com/spf13/cobra"

	"github.com/onflow/flow-cli/internal/command"
	"github.com/onflow/flow-cli/internal/config"
	"github.com/onflow/flow-cli/pkg/flowkit"
	"github.com/onflow/flow-cli/pkg/flowkit/services"
)

// TODO(sideninja) workaround - init needed to be copied in order to work else there is flag duplicate error

var initFlag = config.FlagsInit{}

var InitCommand = &command.Command{
	Cmd: &cobra.Command{
		Use:     "init",
		Short:   "Initialize a new configuration",
		Example: "flow project init",
	},
	Flags: &initFlag,
	Run: func(
		cmd *cobra.Command,
		args []string,
		loader flowkit.Loader,
		globalFlags command.GlobalFlags,
		services *services.Services,
	) (command.Result, error) {
		sigAlgo := crypto.StringToSignatureAlgorithm(initFlag.ServiceKeySigAlgo)
		if sigAlgo == crypto.UnknownSignatureAlgorithm {
			return nil, fmt.Errorf("invalid signature algorithm: %s", initFlag.ServiceKeySigAlgo)
		}

		hashAlgo := crypto.StringToHashAlgorithm(initFlag.ServiceKeyHashAlgo)
		if hashAlgo == crypto.UnknownHashAlgorithm {
			return nil, fmt.Errorf("invalid hash algorithm: %s", initFlag.ServiceKeyHashAlgo)
		}

		privateKey, err := crypto.DecodePrivateKeyHex(sigAlgo, initFlag.ServicePrivateKey)
		if err != nil {
			return nil, fmt.Errorf("invalid private key: %w", err)
		}

		s, err := services.Project.Init(
			initFlag.Reset,
			initFlag.Global,
			sigAlgo,
			hashAlgo,
			privateKey,
		)
		if err != nil {
			return nil, err
		}

		return &config.InitResult{State: s}, nil
	},
}
