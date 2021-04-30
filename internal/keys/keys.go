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

package keys

import (
	"bytes"
	"encoding/hex"
	"fmt"

	"github.com/onflow/flow-go-sdk"
	"github.com/onflow/flow-go-sdk/crypto"
	"github.com/spf13/cobra"

	"github.com/onflow/flow-cli/pkg/flowcli/util"
)

var Cmd = &cobra.Command{
	Use:              "keys",
	Short:            "Utilities to manage keys",
	TraverseChildren: true,
}

func init() {
	GenerateCommand.AddToParent(Cmd)
}

// KeyResult represent result from all account commands
type KeyResult struct {
	privateKey *crypto.PrivateKey
	publicKey  *crypto.PublicKey
	accountKey *flow.AccountKey
}

// JSON convert result to JSON
func (k *KeyResult) JSON() interface{} {
	result := make(map[string]string)
	result["public"] = hex.EncodeToString(k.privateKey.PublicKey().Encode())

	if k.privateKey != nil {
		result["private"] = hex.EncodeToString(k.privateKey.Encode())
	}

	return result
}

// String convert result to string
func (k *KeyResult) String() string {
	var b bytes.Buffer
	writer := util.CreateTabWriter(&b)

	if k.privateKey != nil {
		fmt.Fprintf(writer, "🔴️ Store private key safely and don't share with anyone! \n")
		fmt.Fprintf(writer, "Private Key \t %x \n", k.privateKey.Encode())
	}

	fmt.Fprintf(writer, "Public Key \t %x \n", k.publicKey.Encode())

	if k.accountKey != nil {
		fmt.Fprintf(writer, "Signature algorithm \t %s\n", k.accountKey.SigAlgo)
		fmt.Fprintf(writer, "Hash algorithm \t %s\n", k.accountKey.HashAlgo)
		fmt.Fprintf(writer, "Weight \t %d\n", k.accountKey.Weight)
		fmt.Fprintf(writer, "Revoked \t %t\n", k.accountKey.Revoked)
	}

	writer.Flush()

	return b.String()
}

// Oneliner show result as one liner grep friendly
func (k *KeyResult) Oneliner() string {
	result := fmt.Sprintf("Public Key: %x, ", k.publicKey.Encode())

	if k.privateKey != nil {
		result += fmt.Sprintf("Private Key: %x", k.privateKey.Encode())
	}

	return result
}
