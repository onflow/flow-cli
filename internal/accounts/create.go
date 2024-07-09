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
	"context"
	"fmt"
	"strings"

	"github.com/onflow/flowkit/v2/accounts"

	"github.com/onflow/flow-go-sdk/crypto"
	"github.com/spf13/cobra"

	"github.com/onflow/flowkit/v2"
	"github.com/onflow/flowkit/v2/output"

	"github.com/onflow/flow-cli/internal/command"
)

type flagsCreate struct {
	Signer   string   `default:"emulator-account" flag:"signer" info:"Account name from configuration used to sign the transaction"`
	Keys     []string `flag:"key" info:"Public keys to attach to account"`
	Weights  []int    `default:"1000" flag:"key-weight" info:"Weight for the key"`
	SigAlgo  []string `default:"ECDSA_P256" flag:"sig-algo" info:"Signature algorithm used to generate the keys"`
	HashAlgo []string `default:"SHA3_256" flag:"hash-algo" info:"Hash used for the digest"`
	Include  []string `default:"" flag:"include" info:"Fields to include in the output"`
}

var createFlags = flagsCreate{}

var createCommand = &command.Command{
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
	globalFlags command.GlobalFlags,
	logger output.Logger,
	flow flowkit.Services,
	state *flowkit.State,
) (command.Result, error) {
	if len(createFlags.Keys) == 0 { // if user doesn't provide any flags go into interactive mode

		if len(createFlags.SigAlgo) > 0 || globalFlags.Network != "" {
			return nil, fmt.Errorf("You provided flags, but no key. A key is required to create an account in manual mode. Remove flags if you'd like to use interactive mode.")
		} else {
			return createInteractive(state)
		}

	} else {
		return createManual(state, flow)
	}
}

func createManual(
	state *flowkit.State,
	flow flowkit.Services,
) (*accountResult, error) {
	sigsFlag := createFlags.SigAlgo
	hashFlag := createFlags.HashAlgo
	keysFlag := createFlags.Keys
	weightFlag := createFlags.Weights

	signer, err := state.Accounts().ByName(createFlags.Signer)
	if err != nil {
		return nil, err
	}

	if len(sigsFlag) == 1 && len(hashFlag) == 1 {
		// Fill up depending on size of key input
		if len(createFlags.Keys) > 1 {
			for i := 1; i < len(createFlags.Keys); i++ {
				sigsFlag = append(sigsFlag, sigsFlag[0])
				hashFlag = append(hashFlag, hashFlag[0])
			}
		}
	} else if len(keysFlag) != len(sigsFlag) || len(sigsFlag) != len(hashFlag) { // double check matching array lengths on inputs
		return nil, fmt.Errorf("must provide a signature and hash algorithm for every key provided to --key: %d keys, %d signature algo, %d hash algo", len(keysFlag), len(sigsFlag), len(hashFlag))
	}

	if len(keysFlag) != len(weightFlag) {
		return nil, fmt.Errorf("must provide a key weight for each key provided, keys provided: %d, weights provided: %d", len(keysFlag), len(weightFlag))
	}

	sigAlgos, err := parseSignatureAlgorithms(sigsFlag)
	if err != nil {
		return nil, err
	}

	hashAlgos, err := parseHashingAlgorithms(hashFlag)
	if err != nil {
		return nil, err
	}

	pubKeys, err := parsePublicKeys(keysFlag, sigAlgos)
	if err != nil {
		return nil, err
	}

	keys := make([]accounts.PublicKey, len(pubKeys))
	for i, key := range pubKeys {
		keys[i] = accounts.PublicKey{
			Public: key, Weight: weightFlag[i], SigAlgo: sigAlgos[i], HashAlgo: hashAlgos[i],
		}
	}

	account, _, err := flow.CreateAccount(
		context.Background(),
		signer,
		keys,
	)
	if err != nil {
		return nil, err
	}

	return &accountResult{
		Account: account,
		include: createFlags.Include,
	}, nil
}

func parseHashingAlgorithms(algorithms []string) ([]crypto.HashAlgorithm, error) {
	hashAlgos := make([]crypto.HashAlgorithm, 0, len(algorithms))
	for _, hashAlgoStr := range algorithms {
		hashAlgo := crypto.StringToHashAlgorithm(hashAlgoStr)
		if hashAlgo == crypto.UnknownHashAlgorithm {
			return nil, fmt.Errorf("invalid hash algorithm: %s", hashAlgoStr)
		}
		hashAlgos = append(hashAlgos, hashAlgo)
	}
	return hashAlgos, nil
}

func parseSignatureAlgorithms(algorithms []string) ([]crypto.SignatureAlgorithm, error) {
	sigAlgos := make([]crypto.SignatureAlgorithm, 0, len(algorithms))
	for _, sigAlgoStr := range algorithms {
		sigAlgo := crypto.StringToSignatureAlgorithm(sigAlgoStr)
		if sigAlgo == crypto.UnknownSignatureAlgorithm {
			return nil, fmt.Errorf("invalid signature algorithm: %s", sigAlgoStr)
		}
		sigAlgos = append(sigAlgos, sigAlgo)
	}
	return sigAlgos, nil
}

func parsePublicKeys(publicKeys []string, sigAlgorithms []crypto.SignatureAlgorithm) ([]crypto.PublicKey, error) {
	pubKeys := make([]crypto.PublicKey, 0, len(createFlags.Keys))
	for i, k := range publicKeys {
		k = strings.TrimPrefix(k, "0x") // clear possible prefix
		key, err := crypto.DecodePublicKeyHex(sigAlgorithms[i], k)
		if err != nil {
			return nil, fmt.Errorf("failed decoding public key: %s with error: %w", k, err)
		}
		pubKeys = append(pubKeys, key)
	}
	return pubKeys, nil
}
