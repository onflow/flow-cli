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
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/onflow/flow-cli/internal/command"
	"github.com/onflow/flow-cli/pkg/flowkit"
	"github.com/onflow/flow-cli/pkg/flowkit/services"
	"github.com/onflow/flow-cli/pkg/flowkit/util"
)

type flagsGenerate struct {
	Signer string `default:"emulator-account" flag:"signer" info:"name of the account used to sign"`
}

var generateFlags = flagsGenerate{}

var GenerateCommand = &command.Command{
	Cmd: &cobra.Command{
		Use:     "generate <message>",
		Short:   "Generate the message signature",
		Example: "flow signatures generate 'The quick brown fox jumps over the lazy dog' --signer alice",
		Args:    cobra.ExactArgs(1),
	},
	Flags: &generateFlags,
	RunS:  sign,
}

func sign(
	args []string,
	_ flowkit.ReaderWriter,
	_ command.GlobalFlags,
	_ *services.Services,
	state *flowkit.State,
) (command.Result, error) {
	message := []byte(args[0])
	accountName := generateFlags.Signer
	acc, err := state.Accounts().ByName(accountName)
	if err != nil {
		return nil, err
	}

	s, err := acc.Key().Signer(context.Background())
	if err != nil {
		return nil, err
	}

	signed, err := s.Sign(message)
	if err != nil {
		return nil, err
	}

	return &SignatureResult{
		result:  string(signed),
		message: string(message),
		key:     acc.Key(),
	}, nil
}

type SignatureResult struct {
	result  string
	message string
	key     flowkit.AccountKey
}

func (s *SignatureResult) pubKey() string {
	pkey, err := s.key.PrivateKey()
	if err == nil {
		return (*pkey).PublicKey().String()
	}

	return "ERR"
}

func (s *SignatureResult) JSON() interface{} {
	return map[string]string{
		"signature": fmt.Sprintf("%x", s.result),
		"message":   s.message,
		"hashAlgo":  s.key.HashAlgo().String(),
		"sigAlgo":   s.key.SigAlgo().String(),
		"pubKey":    s.pubKey(),
	}
}
func (s *SignatureResult) String() string {
	var b bytes.Buffer
	writer := util.CreateTabWriter(&b)

	_, _ = fmt.Fprintf(writer, "Signature \t %x\n", s.result)
	_, _ = fmt.Fprintf(writer, "Message \t %s\n", s.message)
	_, _ = fmt.Fprintf(writer, "Public Key \t %s\n", s.pubKey())
	_, _ = fmt.Fprintf(writer, "Hash Algorithm \t %s\n", s.key.HashAlgo())
	_, _ = fmt.Fprintf(writer, "Signature Algorithm \t %s\n", s.key.SigAlgo())

	_ = writer.Flush()
	return b.String()
}

func (s *SignatureResult) Oneliner() string {

	return fmt.Sprintf(
		"signature: %x, message: %s, hashAlgo: %s, sigAlgo: %s, pubKey: %s",
		s.result, s.message, s.key.HashAlgo(), s.key.SigAlgo(), s.pubKey(),
	)
}
