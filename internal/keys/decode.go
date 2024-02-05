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
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/onflow/flow-go-sdk"
	"github.com/onflow/flow-go-sdk/crypto"
	"github.com/spf13/cobra"

	"github.com/onflow/flowkit"
	"github.com/onflow/flowkit/output"

	"github.com/onflow/flow-cli/internal/command"
)

type flagsDecode struct {
	SigAlgo  string `default:"ECDSA_P256" flag:"sig-algo" info:"Signature algorithm"`
	FromFile string `default:"" flag:"from-file" info:"Load key from file"`
}

var decodeFlags = flagsDecode{}

var decodeCommand = &command.Command{
	Cmd: &cobra.Command{
		Use:       "decode <rlp|pem> <encoded public key>",
		Short:     "Decode an encoded public key",
		Args:      cobra.RangeArgs(1, 2),
		ValidArgs: []string{"rlp", "pem"},
		Example:   "flow keys decode rlp f847b8408...2402038203e8",
	},
	Flags: &decodeFlags,
	Run:   decode,
}

func decode(
	args []string,
	_ command.GlobalFlags,
	_ output.Logger,
	reader flowkit.ReaderWriter,
	_ flowkit.Services,
) (command.Result, error) {
	encoding := args[0]
	fromFile := decodeFlags.FromFile

	var encoded string
	if len(args) > 1 {
		encoded = args[1]
	}

	/* TODO(sideninja) from file flag should be remove and should be replaced with $(echo file)
	   but cobra has an issue with parsing pem content as it recognize it as flag due to ---- characters */
	if encoded != "" && fromFile != "" {
		return nil, fmt.Errorf("can not pass both command argument and from file flag")
	}
	if encoded == "" && fromFile == "" {
		return nil, fmt.Errorf("provide argument for encoded key or use from file flag")
	}

	if fromFile != "" {
		e, err := reader.ReadFile(fromFile)
		if err != nil {
			return nil, err
		}
		encoded = strings.TrimSpace(string(e))
	}

	var accountKey *flow.AccountKey
	var err error
	switch strings.ToLower(encoding) {
	case "pem":
		sigAlgo := crypto.StringToSignatureAlgorithm(decodeFlags.SigAlgo)
		if sigAlgo == crypto.UnknownSignatureAlgorithm {
			return nil, fmt.Errorf("invalid signature algorithm: %s", decodeFlags.SigAlgo)
		}

		accountKey, err = decodePEM(encoded, sigAlgo)
	case "rlp":
		accountKey, err = decodeRLP(encoded)
	default:
		return nil, fmt.Errorf("encoding type not supported. Valid encoding: RLP and PEM")
	}

	if err != nil {
		return nil, err
	}

	return &keyResult{
		publicKey: accountKey.PublicKey,
		sigAlgo:   accountKey.SigAlgo,
		hashAlgo:  accountKey.HashAlgo,
		weight:    accountKey.Weight,
	}, err
}

func decodePEM(pubKey string, sigAlgo crypto.SignatureAlgorithm) (*flow.AccountKey, error) {
	pk, err := crypto.DecodePublicKeyPEM(sigAlgo, pubKey)
	if err != nil {
		return nil, err
	}

	return &flow.AccountKey{
		PublicKey: pk,
		SigAlgo:   sigAlgo,
		Weight:    -1,
	}, nil
}

func decodeRLP(pubKey string) (*flow.AccountKey, error) {
	publicKeyBytes, err := hex.DecodeString(pubKey)
	if err != nil {
		return nil, fmt.Errorf("failed to decode public key: %w", err)
	}

	accountKey, err := flow.DecodeAccountKey(publicKeyBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to decode: %w", err)
	}

	return accountKey, nil
}
