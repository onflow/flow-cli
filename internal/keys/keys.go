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
	"text/tabwriter"

	"github.com/onflow/flow-go-sdk"

	"github.com/onflow/flow-go-sdk/crypto"

	"github.com/spf13/cobra"
)

var Cmd = &cobra.Command{
	Use:              "keys",
	Short:            "Utilities to manage keys",
	TraverseChildren: true,
}

func init() {
	GenerateCommand.Add(Cmd)
	DecodeCommand.Add(Cmd)
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
	result["Private"] = hex.EncodeToString(k.privateKey.PublicKey().Encode())
	result["Public"] = hex.EncodeToString(k.privateKey.Encode())

	return result
}

// String convert result to string
func (k *KeyResult) String() string {
	var b bytes.Buffer
	writer := tabwriter.NewWriter(&b, 0, 8, 1, '\t', tabwriter.AlignRight)

	if k.privateKey != nil {
		fmt.Fprintf(writer, "🔴️ Store Private Key safely and don't share with anyone! \n")
		fmt.Fprintf(writer, "Private Key \t %x \n", k.privateKey.Encode())
	}

	fmt.Fprintf(writer, "Public Key \t %x \n", k.publicKey.Encode())

	if k.accountKey != nil {
		fmt.Fprintf(writer, "Signature algorithm \t %s\n", k.accountKey.SigAlgo)
		fmt.Fprintf(writer, "Hash algorithm \t %s\n", k.accountKey.HashAlgo)
		fmt.Fprintf(writer, "Weight \t %d\n", k.accountKey.Weight)
	}

	writer.Flush()

	return b.String()
}

// Oneliner show result as one liner grep friendly
func (k *KeyResult) Oneliner() string {
	return fmt.Sprintf("Private Key: %x, Public Key: %x", k.privateKey.Encode(), k.publicKey.Encode())
}

func (k *KeyResult) ToConfig() string {
	return ""
}
