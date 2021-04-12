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

	"github.com/onflow/cadence"
	"github.com/onflow/flow-go-sdk"
	"github.com/spf13/cobra"
)

var Cmd = &cobra.Command{
	Use:              "accounts",
	Short:            "Utilities to manage accounts",
	TraverseChildren: true,
}

func init() {
	AddContractCommand.AddToParent(Cmd)
	RemoveCommand.AddToParent(Cmd)
	UpdateCommand.AddToParent(Cmd)
	CreateCommand.AddToParent(Cmd)
	StakingCommand.AddToParent(Cmd)
	GetCommand.AddToParent(Cmd)
}

// AccountResult represent result from all account commands
type AccountResult struct {
	*flow.Account
	showCode bool
}

// JSON convert result to JSON
func (r *AccountResult) JSON() interface{} {
	result := make(map[string]interface{})
	result["address"] = r.Address
	result["balance"] = r.Balance

	keys := make([]string, 0)
	for _, key := range r.Keys {
		keys = append(keys, fmt.Sprintf("%x", key.PublicKey.Encode()))
	}

	result["keys"] = keys

	contracts := make([]string, 0)
	for name := range r.Contracts {
		contracts = append(contracts, name)
	}

	result["contracts"] = contracts

	if r.showCode {
		c := make(map[string]string)
		for name, code := range r.Contracts {
			c[name] = string(code)
		}
		result["code"] = c
	}

	return result
}

// String convert result to string
func (r *AccountResult) String() string {
	var b bytes.Buffer
	writer := tabwriter.NewWriter(&b, 0, 8, 1, '\t', tabwriter.AlignRight)

	fmt.Fprintf(writer, "Address\t 0x%s\n", r.Address)
	fmt.Fprintf(writer, "Balance\t %s\n", cadence.UFix64(r.Balance))

	fmt.Fprintf(writer, "Keys\t %d\n", len(r.Keys))

	for i, key := range r.Keys {
		fmt.Fprintf(writer, "\nKey %d\tPublic Key\t %x\n", i, key.PublicKey.Encode())
		fmt.Fprintf(writer, "\tWeight\t %d\n", key.Weight)
		fmt.Fprintf(writer, "\tSignature Algorithm\t %s\n", key.SigAlgo)
		fmt.Fprintf(writer, "\tHash Algorithm\t %s\n", key.HashAlgo)
		fmt.Fprintf(writer, "\tRevoked \t %t\n", key.Revoked)
		fmt.Fprintf(writer, "\n")
	}

	fmt.Fprintf(writer, "Contracts Deployed: %d\n", len(r.Contracts))
	for name := range r.Contracts {
		fmt.Fprintf(writer, "Contract: '%s'\n", name)
	}

	if r.showCode {
		for name, code := range r.Contracts {
			fmt.Fprintf(writer, "Contracts '%s':\n", name)
			fmt.Fprintln(writer, string(code))
		}
	}

	writer.Flush()

	return b.String()
}

// Oneliner show result as one liner grep friendly
func (r *AccountResult) Oneliner() string {
	keys := make([]string, 0)
	for _, key := range r.Keys {
		keys = append(keys, key.PublicKey.String())
	}

	return fmt.Sprintf("Address: 0x%s, Balance: %v, Public Keys: %s", r.Address, r.Balance, keys)
}
