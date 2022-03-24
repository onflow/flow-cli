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

package accounts

import (
	"bytes"
	"fmt"

	"github.com/onflow/flow-cli/internal/command"

	"github.com/onflow/cadence"
	"github.com/onflow/flow-go-sdk"
	"github.com/spf13/cobra"

	"github.com/onflow/flow-cli/pkg/flowkit/util"
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

// AccountResult represent result from all account commands.
type AccountResult struct {
	*flow.Account
	include []string
}

func (r *AccountResult) JSON() interface{} {
	result := make(map[string]interface{})
	result["address"] = r.Address
	result["balance"] = cadence.UFix64(r.Balance).String()

	keys := make([]string, 0)
	for _, key := range r.Keys {
		keys = append(keys, fmt.Sprintf("%x", key.PublicKey.Encode()))
	}

	result["keys"] = keys

	contracts := make([]string, 0, len(r.Contracts))
	for name := range r.Contracts {
		contracts = append(contracts, name)
	}

	result["contracts"] = contracts

	if command.ContainsFlag(r.include, "contracts") {
		c := make(map[string]string)
		for name, code := range r.Contracts {
			c[name] = string(code)
		}
		result["code"] = c
	}

	return result
}

func (r *AccountResult) String() string {
	var b bytes.Buffer
	writer := util.CreateTabWriter(&b)

	_, _ = fmt.Fprintf(writer, "Address\t 0x%s\n", r.Address)
	_, _ = fmt.Fprintf(writer, "Balance\t %s\n", cadence.UFix64(r.Balance))

	_, _ = fmt.Fprintf(writer, "Keys\t %d\n", len(r.Keys))

	for i, key := range r.Keys {
		_, _ = fmt.Fprintf(writer, "\nKey %d\tPublic Key\t %x\n", i, key.PublicKey.Encode())
		_, _ = fmt.Fprintf(writer, "\tWeight\t %d\n", key.Weight)
		_, _ = fmt.Fprintf(writer, "\tSignature Algorithm\t %s\n", key.SigAlgo)
		_, _ = fmt.Fprintf(writer, "\tHash Algorithm\t %s\n", key.HashAlgo)
		_, _ = fmt.Fprintf(writer, "\tRevoked \t %t\n", key.Revoked)
		_, _ = fmt.Fprintf(writer, "\tSequence Number \t %d\n", key.SequenceNumber)
		_, _ = fmt.Fprintf(writer, "\tIndex \t %d\n", key.Index)
		_, _ = fmt.Fprintf(writer, "\n")

		// only show up to 3 keys and then show label to expand more info
		if i == 3 && !command.ContainsFlag(r.include, "keys") {
			_, _ = fmt.Fprint(writer, "...keys minimized, use --include keys flag if you want to view all\n\n")
			break
		}
	}

	_, _ = fmt.Fprintf(writer, "Contracts Deployed: %d\n", len(r.Contracts))
	for name := range r.Contracts {
		_, _ = fmt.Fprintf(writer, "Contract: '%s'\n", name)
	}

	if command.ContainsFlag(r.include, "contracts") {
		for name, code := range r.Contracts {
			_, _ = fmt.Fprintf(writer, "Contracts '%s':\n", name)
			_, _ = fmt.Fprintln(writer, string(code))
		}
	} else {
		_, _ = fmt.Fprint(writer, "\n\nContracts (hidden, use --include contracts)")
	}

	_ = writer.Flush()

	return b.String()
}

func (r *AccountResult) Oneliner() string {
	keys := make([]string, 0, len(r.Keys))
	for _, key := range r.Keys {
		keys = append(keys, key.PublicKey.String())
	}

	return fmt.Sprintf("Address: 0x%s, Balance: %s, Public Keys: %s", r.Address, cadence.UFix64(r.Balance), keys)
}
