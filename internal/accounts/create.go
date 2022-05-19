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
	SigAlgo   []string `default:"ECDSA_P256" flag:"sig-algo" info:"Signature algorithm used to generate the keys"`
	HashAlgo  []string `default:"SHA3_256" flag:"hash-algo" info:"Hash used for the digest"`
	Contracts []string `flag:"contract" info:"Contract to be deployed during account creation. <name:filename>"`
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
	signer, err := state.Accounts().ByName(createFlags.Signer)
	if err != nil {
		return nil, err
	}

	if len(createFlags.SigAlgo) == 1 && len(createFlags.HashAlgo) == 1 {
		// Fill up depending on size of key input
		if len(createFlags.Keys) > 1 {
			for i := 1; i < len(createFlags.Keys); i++ {
				createFlags.SigAlgo = append(createFlags.SigAlgo, createFlags.SigAlgo[0])
				createFlags.HashAlgo = append(createFlags.HashAlgo, createFlags.HashAlgo[0])
			}
			// Deprecated usage message?
		}

	} else
	// double check matching array lengths on inputs
	if len(createFlags.Keys) != len(createFlags.SigAlgo) || len(createFlags.SigAlgo) != len(createFlags.HashAlgo) {
		return nil, fmt.Errorf("must provide a signature and hash algorithm for every key provided to --key: %d keys, %d signature algo, %d hash algo", len(createFlags.Keys), len(createFlags.SigAlgo), len(createFlags.HashAlgo))
	}

	// read all signature algorithms
	sigAlgos := make([]crypto.SignatureAlgorithm, 0, len(createFlags.SigAlgo))
	for _, sigAlgoStr := range createFlags.SigAlgo {
		sigAlgo := crypto.StringToSignatureAlgorithm(sigAlgoStr)
		if sigAlgo == crypto.UnknownSignatureAlgorithm {
			return nil, fmt.Errorf("invalid signature algorithm: %s", createFlags.SigAlgo)
		}
		sigAlgos = append(sigAlgos, sigAlgo)
	}

	// read all hash algorithms
	hashAlgos := make([]crypto.HashAlgorithm, 0, len(createFlags.HashAlgo))
	for _, hashAlgoStr := range createFlags.HashAlgo {

		hashAlgo := crypto.StringToHashAlgorithm(hashAlgoStr)
		if hashAlgo == crypto.UnknownHashAlgorithm {
			return nil, fmt.Errorf("invalid hash algorithm: %s", createFlags.HashAlgo)
		}
		hashAlgos = append(hashAlgos, hashAlgo)
	}

	keyWeights := createFlags.Weights

	// decode public keys
	pubKeys := make([]crypto.PublicKey, 0, len(createFlags.Keys))
	for i, k := range createFlags.Keys {
		k = strings.TrimPrefix(k, "0x") // clear possible prefix
		key, err := crypto.DecodePublicKeyHex(sigAlgos[i], k)
		if err != nil {
			return nil, fmt.Errorf("failed decoding public key: %s with error: %w", key, err)
		}
		pubKeys = append(pubKeys, key)
	}

	account, err := services.Accounts.Create(
		signer,
		pubKeys,
		keyWeights,
		sigAlgos,
		hashAlgos,
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
