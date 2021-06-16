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

package accounts

import (
	"fmt"
	"strings"

	"github.com/onflow/flow-cli/pkg/flowkit"

	"github.com/onflow/flow-go-sdk/crypto"

	"github.com/spf13/cobra"

	"github.com/onflow/flow-cli/internal/command"
	"github.com/onflow/flow-cli/pkg/flowkit/services"
)

type flagsCreate struct {
	Signer    string   `default:"emulator-account" flag:"signer" info:"Account name from configuration used to sign the transaction"`
	Keys      []string `flag:"key" info:"Public keys to attach to account"`
	Weights   []int    `flag:"key-weight" info:"Weight for the key"`
	SigAlgo   string   `default:"ECDSA_P256" flag:"sig-algo" info:"Signature algorithm used to generate the keys"`
	HashAlgo  string   `default:"SHA3_256" flag:"hash-algo" info:"Hash used for the digest"`
	Contracts []string `flag:"contract" info:"Contract to be deployed during account creation. <name:filename>"`
	Results   bool     `default:"false" flag:"results" info:"⚠️  Deprecated: results are provided by default"`
	Include   []string `default:"" flag:"include" info:"Fields to include in the output"`
}

var createFlags = flagsCreate{}

var CreateCommand = &command.Command{
	Cmd: &cobra.Command{
		Use:     "create",
		Short:   "Create a new account on network",
		Example: `flow accounts create --key d651f1931a2...8745`,
	},
	Flags: &createFlags,
	RunS:  create,
}

func create(
	_ []string,
	_ flowkit.ReaderWriter,
	_ command.GlobalFlags,
	services *services.Services,
	state *flowkit.State,
) (command.Result, error) {
	if createFlags.Results {
		fmt.Println("⚠️ DEPRECATION WARNING: results flag is deprecated, results are by default included in all executions")
	}

	signer := state.Accounts().ByName(createFlags.Signer)

	sigAlgo := crypto.StringToSignatureAlgorithm(createFlags.SigAlgo)
	if sigAlgo == crypto.UnknownSignatureAlgorithm {
		return nil, fmt.Errorf("invalid signature algorithm: %s", createFlags.SigAlgo)
	}

	hashAlgo := crypto.StringToHashAlgorithm(createFlags.HashAlgo)
	if hashAlgo == crypto.UnknownHashAlgorithm {
		return nil, fmt.Errorf("invalid hash algorithm: %s", createFlags.HashAlgo)
	}

	keyWeights := createFlags.Weights

	// decode public keys
	var pubKeys []crypto.PublicKey
	for _, k := range createFlags.Keys {
		k = strings.ReplaceAll(k, "0x", "") // clear possible prefix
		key, err := crypto.DecodePublicKeyHex(sigAlgo, k)
		if err != nil {
			return nil, fmt.Errorf("failed decoding public key: %s with error: %w", key, err)
		}
		pubKeys = append(pubKeys, key)
	}

	account, err := services.Accounts.Create(
		signer,
		pubKeys,
		keyWeights,
		sigAlgo,
		hashAlgo,
		createFlags.Contracts,
	)

	if err != nil {
		return nil, err
	}

	return &AccountResult{
		Account: account,
		include: createFlags.Include,
	}, nil
}
