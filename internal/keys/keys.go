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
	"bytes"
	"encoding/hex"
	"fmt"

	"github.com/onflow/flow-go-sdk/crypto"
	"github.com/spf13/cobra"

	"github.com/onflow/flowkit/v2/output"

	"github.com/onflow/flow-cli/internal/util"
)

var Cmd = &cobra.Command{
	Use:              "keys",
	Short:            "Generate and decode Flow keys",
	TraverseChildren: true,
	GroupID:          "security",
}

func init() {
	generateCommand.AddToParent(Cmd)
	decodeCommand.AddToParent(Cmd)
	deriveCommand.AddToParent(Cmd)
}

type keyResult struct {
	privateKey     crypto.PrivateKey
	publicKey      crypto.PublicKey
	sigAlgo        crypto.SignatureAlgorithm
	hashAlgo       crypto.HashAlgorithm
	weight         int
	mnemonic       string
	derivationPath string
}

func (k *keyResult) JSON() any {
	result := make(map[string]any)
	result["public"] = hex.EncodeToString(k.privateKey.PublicKey().Encode())

	if k.privateKey != nil {
		result["private"] = hex.EncodeToString(k.privateKey.Encode())
	}

	if k.mnemonic != "" {
		result["mnemonic"] = k.mnemonic
	}

	if k.derivationPath != "" {
		result["derivationPath"] = k.derivationPath
	}

	return result
}

func (k *keyResult) String() string {
	var b bytes.Buffer
	writer := util.CreateTabWriter(&b)

	if k.privateKey != nil {
		_, _ = fmt.Fprintf(writer, "%s Store private key safely and don't share with anyone! \n", output.StopEmoji())
		_, _ = fmt.Fprintf(writer, "Private Key \t %x \n", k.privateKey.Encode())
	}

	_, _ = fmt.Fprintf(writer, "Public Key \t %x \n", k.publicKey.Encode())

	if k.mnemonic != "" {
		_, _ = fmt.Fprintf(writer, "Mnemonic \t %s \n", k.mnemonic)
	}

	if k.derivationPath != "" {
		_, _ = fmt.Fprintf(writer, "Derivation Path \t %s \n", k.derivationPath)
	}

	if k.sigAlgo != crypto.UnknownSignatureAlgorithm {
		_, _ = fmt.Fprintf(writer, "Signature Algorithm \t %s\n", k.sigAlgo)
	}

	if k.hashAlgo != crypto.UnknownHashAlgorithm {
		_, _ = fmt.Fprintf(writer, "Hash Algorithm \t %s\n", k.hashAlgo)
	}

	if k.weight > 0 {
		_, _ = fmt.Fprintf(writer, "Weight \t %d\n", k.weight)
	}

	_ = writer.Flush()

	return b.String()
}

func (k *keyResult) Oneliner() string {
	result := fmt.Sprintf("Public Key: %x, ", k.publicKey.Encode())

	if k.privateKey != nil {
		result += fmt.Sprintf("Private Key: %x, ", k.privateKey.Encode())
	}

	if k.mnemonic != "" {
		result += fmt.Sprintf("Mnemonic: %s, ", k.mnemonic)
	}

	if k.derivationPath != "" {
		result += fmt.Sprintf("Derivation Path: %s", k.derivationPath)
	}

	return result
}
