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

package signatures

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/onflow/flow-cli/pkg/flowkit/util"

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
		Use:     "verify <message> <signature> <public key>",
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
	message := []byte(args[0])

	sig, err := hex.DecodeString(strings.ReplaceAll(args[1], "0x", ""))
	if err != nil {
		return nil, fmt.Errorf("invalid message signature: %w", err)
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

	valid, err := pkey.Verify(sig, message, hasher)
	if err != nil {
		return nil, err
	}

	return &VerificationResult{
		valid:     valid,
		message:   message,
		signature: sig,
		hashAlgo:  hashAlgo,
		sigAlgo:   sigAlgo,
		pubKey:    key,
	}, nil
}

type VerificationResult struct {
	valid     bool
	message   []byte
	signature []byte
	pubKey    []byte
	sigAlgo   crypto.SignatureAlgorithm
	hashAlgo  crypto.HashAlgorithm
}

func (s *VerificationResult) JSON() interface{} {
	return map[string]string{
		"valid":     fmt.Sprintf("%v", s.valid),
		"message":   string(s.message),
		"signature": string(s.signature),
		"hashAlgo":  s.hashAlgo.String(),
		"sigAlgo":   s.sigAlgo.String(),
		"pubKey":    fmt.Sprintf("%x", s.pubKey),
	}
}
func (s *VerificationResult) String() string {
	var b bytes.Buffer
	writer := util.CreateTabWriter(&b)

	_, _ = fmt.Fprintf(writer, "Valid \t %v\n", s.valid)
	_, _ = fmt.Fprintf(writer, "Message \t %s\n", s.message)
	_, _ = fmt.Fprintf(writer, "Signature \t %x\n", s.signature)
	_, _ = fmt.Fprintf(writer, "Public Key \t %x\n", s.pubKey)
	_, _ = fmt.Fprintf(writer, "Hash Algorithm \t %s\n", s.hashAlgo)
	_, _ = fmt.Fprintf(writer, "Signature Algorithm \t %s\n", s.sigAlgo)

	_ = writer.Flush()
	return b.String()
}

func (s *VerificationResult) Oneliner() string {
	return fmt.Sprintf(
		"valid: %v, message: %s, signature: %x, sigAlgo: %s, hashAlgo: %s, pubKey: %x",
		s.valid, s.message, s.signature, s.sigAlgo, s.hashAlgo, s.pubKey,
	)
}
