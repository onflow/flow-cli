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

package decode

import (
	"encoding/hex"
	"fmt"

	flowsdk "github.com/onflow/flow-go-sdk"
	"github.com/spf13/cobra"

	cli "github.com/onflow/flow-cli/flow"
)

var Cmd = &cobra.Command{
	Use:   "decode <public key>",
	Short: "Decode a public account key hex string",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {

		publicKey := args[0]

		publicKeyBytes, err := hex.DecodeString(publicKey)
		if err != nil {
			cli.Exitf(1, "Failed to decode public key: %v", err)
		}

		accountKey, err := flowsdk.DecodeAccountKey(publicKeyBytes)
		if err != nil {
			cli.Exitf(1, "Failed to decode private key bytes: %v", err)
		}

		fmt.Printf("  Public key: %x\n", accountKey.PublicKey.Encode())
		fmt.Printf("  Signature algorithm: %s\n", accountKey.SigAlgo)
		fmt.Printf("  Hash algorithm: %s\n", accountKey.HashAlgo)
		fmt.Printf("  Weight: %d\n", accountKey.Weight)
	},
}
