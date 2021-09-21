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

package signatures

import (
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/onflow/flow-go-sdk/crypto"
	"github.com/spf13/cobra"

	"github.com/onflow/flow-cli/internal/command"
	"github.com/onflow/flow-cli/pkg/flowkit"
	"github.com/onflow/flow-cli/pkg/flowkit/services"
)

type flagsVerify struct {
	SigAlgo  string `flag:"sig-algo" default:"ECDSA_P256" info:"Signature algorithm used to create the public key"`
	HashAlgo string `flag:"hash-algo" default:"SHA3_256" info:"Hashing algorithm used to create signature"`
}

var verifyFlags = flagsVerify{}

var VerifyCommand = &command.Command{
	Cmd: &cobra.Command{
		Use:     "verify <payload> <signature> <public key>",
		Short:   "Verify the signature",
		Example: "flow signatures verify 'The quick brown fox jumps over the lazy dog' 99fa...25b af3...52d",
		Args:    cobra.ExactArgs(3),
	},
	Flags: &verifyFlags,
	RunS:  verify,
}

func verify(
	args []string,
	_ flowkit.ReaderWriter,
	_ command.GlobalFlags,
	_ *services.Services,
	_ *flowkit.State,
) (command.Result, error) {
	payload := []byte(args[0])

	sig, err := hex.DecodeString(strings.ReplaceAll(args[1], "0x", ""))
	if err != nil {
		return nil, fmt.Errorf("invalid payload signature: %w", err)
	}

	key, err := hex.DecodeString(strings.ReplaceAll(args[2], "0x", ""))
	if err != nil {
		return nil, fmt.Errorf("invalid public key: %w", err)
	}

	sigAlgo := crypto.StringToSignatureAlgorithm(verifyFlags.SigAlgo)
	hashAlgo := crypto.StringToHashAlgorithm(verifyFlags.HashAlgo)

	pkey, err := crypto.DecodePublicKey(sigAlgo, key)
	if err != nil {
		return nil, fmt.Errorf("invalid public key: %w", err)
	}

	hasher, err := crypto.NewHasher(hashAlgo)
	if err != nil {
		return nil, err
	}

	valid, err := pkey.Verify(sig, payload, hasher)
	if err != nil {
		return nil, err
	}

	return &VerificationResult{
		valid:     valid,
		payload:   payload,
		signature: sig,
	}, nil
}

type VerificationResult struct {
	valid     bool
	payload   []byte
	signature []byte
}

func (s *VerificationResult) JSON() interface{} {
	return map[string]string{
		"valid":     fmt.Sprintf("%v", s.valid),
		"payload":   fmt.Sprintf("%s", s.payload),
		"signature": fmt.Sprintf("%s", s.signature),
	}
}
func (s *VerificationResult) String() string {
	return fmt.Sprintf(
		"valid: \t\t%v\npayload: \t%s\nsignature: \t%x\n",
		s.valid, s.payload, s.signature,
	)
}

func (s *VerificationResult) Oneliner() string {
	return fmt.Sprintf(
		"valid: %v, payload: %s, signature: %x",
		s.valid, s.payload, s.signature,
	)
}
