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

package accounts

import (
	"bytes"
	"fmt"
	"text/tabwriter"

	flowsdk "github.com/onflow/flow-go-sdk"
	"github.com/spf13/cobra"
)

var Cmd = &cobra.Command{
	Use:              "accounts",
	Short:            "Utilities to manage accounts",
	TraverseChildren: true,
}

// AccountResult represent result from all account commands
type AccountResult struct {
	*flowsdk.Account
	showCode bool
}

// JSON convert result to JSON
func (r *AccountResult) JSON() interface{} {
	return r
}

// String convert result to string
func (r *AccountResult) String() string {
	var b bytes.Buffer
	writer := tabwriter.NewWriter(&b, 0, 8, 1, '\t', tabwriter.AlignRight)

	fmt.Fprintf(writer, "Address\t 0x%s\n", r.Address)
	fmt.Fprintf(writer, "Balance\t %d\n", r.Balance)

	fmt.Fprintf(writer, "Keys\t %d\n", len(r.Keys))

	for i, key := range r.Keys {
		fmt.Fprintf(writer, "\nKey %d\tPublic Key\t %x\n", i, key.PublicKey.Encode())
		fmt.Fprintf(writer, "\tWeight\t %d\n", key.Weight)
		fmt.Fprintf(writer, "\tSignature Algorithm\t %s\n", key.SigAlgo)
		fmt.Fprintf(writer, "\tHash Algorithm\t %s\n", key.HashAlgo)
		fmt.Fprintf(writer, "\n")
	}

	fmt.Fprintf(writer, "Contracts Deployed: %d\n", len(r.Contracts))
	for name, _ := range r.Contracts {
		fmt.Fprintf(writer, "Contract: '%s'\n", name)
	}

	if r.showCode {
		for name, code := range r.Contracts {
			fmt.Fprintf(writer, "Code '%s':\n", name)
			fmt.Fprintln(writer, string(code))
		}
	}

	writer.Flush()

	return b.String()
}

// Oneliner show result as one liner grep friendly
func (r *AccountResult) Oneliner() string {
	return fmt.Sprintf("Address: 0x%s, Balance: %v, Keys: %s", r.Address, r.Balance, r.Keys[0].PublicKey)
}

func (r *AccountResult) ToConfig() string {
	// TODO: it would be good to have a --save-config flag and it would be added to config
	return ""
}
