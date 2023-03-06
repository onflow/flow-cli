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

package keys

import (
	"fmt"

	"github.com/onflow/flow-go-sdk/crypto"
	"github.com/spf13/cobra"

	"github.com/onflow/flow-cli/internal/command"
	"github.com/onflow/flow-cli/pkg/flowkit"
	"github.com/onflow/flow-cli/pkg/flowkit/output"
)

type flagsDerive struct {
	KeySigAlgo string `default:"ECDSA_P256" flag:"sig-algo" info:"Signature algorithm"`
}

var deriveFlags = flagsDerive{}

var DeriveCommand = &command.Command{
	Cmd: &cobra.Command{
		Use:     "derive <encoded private key>",
		Short:   "Derive public key from a private key",
		Args:    cobra.ExactArgs(1),
		Example: "flow keys derive 4247b8408...2402038203e8",
	},
	Flags: &deriveFlags,
	Run:   derive,
}

func derive(
	args []string,
	_ command.GlobalFlags,
	_ output.Logger,
	_ flowkit.ReaderWriter,
	_ flowkit.Services,
) (command.Result, error) {

	sigAlgo := crypto.StringToSignatureAlgorithm(deriveFlags.KeySigAlgo)
	if sigAlgo == crypto.UnknownSignatureAlgorithm {
		return nil, fmt.Errorf("invalid signature algorithm: %s", deriveFlags.KeySigAlgo)
	}

	parsedPrivateKey, err := crypto.DecodePrivateKeyHex(sigAlgo, args[0])
	if err != nil {
		return nil, fmt.Errorf("failed to decode private key: %w", err)
	}

	return &KeyResult{privateKey: parsedPrivateKey, publicKey: parsedPrivateKey.PublicKey()}, nil
}
