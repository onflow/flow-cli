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

package util

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"strings"
	"text/tabwriter"

	"github.com/onflow/flow-go-sdk"
	"github.com/onflow/flow-go-sdk/crypto"
)

// ConvertSigAndHashAlgo parses and validates a signature and hash algorithm pair.
func ConvertSigAndHashAlgo(
	signatureAlgorithm string,
	hashingAlgorithm string,
) (crypto.SignatureAlgorithm, crypto.HashAlgorithm, error) {
	sigAlgo := crypto.StringToSignatureAlgorithm(signatureAlgorithm)
	if sigAlgo == crypto.UnknownSignatureAlgorithm {
		return crypto.UnknownSignatureAlgorithm,
			crypto.UnknownHashAlgorithm,
			fmt.Errorf("failed to determine signature algorithm from %s", sigAlgo)
	}

	hashAlgo := crypto.StringToHashAlgorithm(hashingAlgorithm)
	if hashAlgo == crypto.UnknownHashAlgorithm {
		return crypto.UnknownSignatureAlgorithm,
			crypto.UnknownHashAlgorithm,
			fmt.Errorf("failed to determine hash algorithm from %s", hashAlgo)
	}

	return sigAlgo, hashAlgo, nil
}

// ContainsString returns true if the slice contains the given string.
func ContainsString(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}

	return false
}

// GetAddressNetwork returns the chain ID for an address.
func GetAddressNetwork(address flow.Address) (flow.ChainID, error) {
	networks := []flow.ChainID{
		flow.Mainnet,
		flow.Testnet,
		flow.Emulator,
	}
	for _, net := range networks {
		if address.IsValid(net) {
			return net, nil
		}
	}

	return "", fmt.Errorf("address not valid for any known chain: %s", address)
}

func CreateTabWriter(b *bytes.Buffer) *tabwriter.Writer {
	return tabwriter.NewWriter(b, 0, 8, 1, '\t', tabwriter.AlignRight)
}

func ParseAddress(value string) (flow.Address, bool) {
	address := flow.HexToAddress(value)

	// valid on any chain
	return address, address.IsValid(flow.Mainnet) ||
		address.IsValid(flow.Testnet) ||
		address.IsValid(flow.Emulator)
}

func RemoveFromStringArray(s []string, el string) []string {
	for i, v := range s {
		if v == el {
			s = append(s[:i], s[i+1:]...)
			break
		}
	}

	return s
}

// ValidateECDSAP256Pub attempt to decode the hex string representation of a ECDSA P256 public key
func ValidateECDSAP256Pub(key string) error {
	b, err := hex.DecodeString(strings.TrimPrefix(key, "0x"))
	if err != nil {
		return fmt.Errorf("failed to decode public key hex string: %w", err)
	}

	_, err = crypto.DecodePublicKey(crypto.ECDSA_P256, b)
	if err != nil {
		return fmt.Errorf("failed to decode public key: %w", err)
	}

	return nil
}
