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

package services

import (
	"encoding/hex"
	"fmt"

	"github.com/onflow/flow-cli/pkg/flowkit"

	"github.com/onflow/flow-go-sdk"
	"github.com/onflow/flow-go-sdk/crypto"

	"github.com/onflow/flow-cli/pkg/flowkit/gateway"
	"github.com/onflow/flow-cli/pkg/flowkit/output"
	"github.com/onflow/flow-cli/pkg/flowkit/util"

	goeth "github.com/ethereum/go-ethereum/accounts"
	slip10 "github.com/lmars/go-slip10"
	bip39 "github.com/tyler-smith/go-bip39"
)

// Keys is a service that handles all key-related interactions.
type Keys struct {
	gateway gateway.Gateway
	state   *flowkit.State
	logger  output.Logger
}

// NewKeys returns a new keys service.
func NewKeys(
	gateway gateway.Gateway,
	state *flowkit.State,
	logger output.Logger,
) *Keys {
	return &Keys{
		gateway: gateway,
		state:   state,
		logger:  logger,
	}
}

func (k *Keys) GetMnemonic() (string, error) {
	entropy, err := bip39.NewEntropy(128)
	if err != nil {
		return "", err
	}
	mnemonic, err := bip39.NewMnemonic(entropy)
	if err != nil {
		return "", err
	}
	return mnemonic, nil
}

func (k *Keys) DerivePrivateKeyFromMnemonic(mnemonic string, sigAlgo crypto.SignatureAlgorithm, derivationPath string) (crypto.PrivateKey, error) {

	if derivationPath == "" {
		derivationPath = "m/44'/539'/0'/0/0"
	}

	if !bip39.IsMnemonicValid(mnemonic) {
		return nil, fmt.Errorf("invalid mnemonic")
	}

	path, err := goeth.ParseDerivationPath(derivationPath)
	if err != nil {
		return nil, fmt.Errorf("invalid derivation path")
	}

	seed := bip39.NewSeed(mnemonic, "")
	curve := slip10.CurveBitcoin
	if sigAlgo == crypto.ECDSA_P256 {
		curve = slip10.CurveP256
	}

	accountKey, err := slip10.NewMasterKeyWithCurve(seed, curve)
	if err != nil {
		return nil, err
	}

	for _, n := range path {
		accountKey, err = accountKey.NewChildKey(n)

		if err != nil {
			return nil, err
		}
	}
	privateKey, err := crypto.DecodePrivateKey(sigAlgo, accountKey.Key)
	if err != nil {
		return nil, err
	}
	return privateKey, nil
}

// Parses private key
func (k *Keys) ParsePrivateKey(inputPrivateKey string, sigAlgo crypto.SignatureAlgorithm) (crypto.PrivateKey, error) {
	privateKey, err := crypto.DecodePrivateKeyHex(sigAlgo, inputPrivateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to decode private key: %w", err)
	}
	return privateKey, nil
}

// Generate generates a new private key from the given seed and signature algorithm.
func (k *Keys) Generate(inputSeed string, sigAlgo crypto.SignatureAlgorithm) (crypto.PrivateKey, error) {
	var seed []byte
	var err error

	if inputSeed == "" {
		seed, err = util.RandomSeed(crypto.MinSeedLength)
		if err != nil {
			return nil, err
		}
	} else {
		seed = []byte(inputSeed)
	}

	privateKey, err := crypto.GeneratePrivateKey(sigAlgo, seed)
	if err != nil {
		return nil, fmt.Errorf("failed to generate private key: %w", err)
	}

	return privateKey, nil
}

// DecodeRLP decodes an RLP encoded public key
func (k *Keys) DecodeRLP(publicKey string) (*flow.AccountKey, error) {
	publicKeyBytes, err := hex.DecodeString(publicKey)
	if err != nil {
		return nil, fmt.Errorf("failed to decode public key: %w", err)
	}

	accountKey, err := flow.DecodeAccountKey(publicKeyBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to decode: %w", err)
	}

	return accountKey, nil
}

// DecodePEM decodes a PEM encoded public key with specified signature algorithm.
func (k *Keys) DecodePEM(key string, sigAlgo crypto.SignatureAlgorithm) (*flow.AccountKey, error) {
	pk, err := crypto.DecodePublicKeyPEM(sigAlgo, key)
	if err != nil {
		return nil, err
	}

	return &flow.AccountKey{
		PublicKey: pk,
		SigAlgo:   sigAlgo,
		Weight:    -1,
	}, nil
}
