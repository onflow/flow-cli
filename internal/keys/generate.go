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
	"github.com/onflow/flow-cli/pkg/flowkit/services"
)

type flagsGenerate struct {
	Seed       string `flag:"seed" info:"Deterministic seed phrase"`
	KeySigAlgo string `default:"ECDSA_P256" flag:"sig-algo" info:"Signature algorithm"`
}

var generateFlags = flagsGenerate{}

var GenerateCommand = &command.Command{
	Cmd: &cobra.Command{
		Use:     "generate",
		Short:   "Generate a new key-pair",
		Example: "flow keys generate",
	},
	Flags: &generateFlags,
	Run:   generate,
}

func generate(
	_ []string,
	_ flowkit.ReaderWriter,
	_ command.GlobalFlags,
	services *services.Services,
) (command.Result, error) {
	sigAlgo := crypto.StringToSignatureAlgorithm(generateFlags.KeySigAlgo)
	if sigAlgo == crypto.UnknownSignatureAlgorithm {
		return nil, fmt.Errorf("invalid signature algorithm: %s", generateFlags.KeySigAlgo)
	}

	privateKey, err := services.Keys.Generate(generateFlags.Seed, sigAlgo)
	if err != nil {
		return nil, err
	}

	pubKey := privateKey.PublicKey()
	return &KeyResult{privateKey: privateKey, publicKey: pubKey}, nil
}
