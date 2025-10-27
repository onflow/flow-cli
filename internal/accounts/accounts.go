/*
 * Flow CLI
 *
 * Copyright Flow Foundation
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

	"github.com/onflow/cadence"
	"github.com/onflow/flow-go-sdk"
	"github.com/spf13/cobra"

	"github.com/onflow/flow-cli/common/branding"
	"github.com/onflow/flow-cli/internal/command"
	"github.com/onflow/flow-cli/internal/util"
)

var Cmd = &cobra.Command{
	Use:              "accounts",
	Short:            "Create and retrieve accounts and deploy contracts",
	TraverseChildren: true,
	GroupID:          "resources",
}

func testnetFaucetURL(address flow.Address) string {
	return fmt.Sprintf("https://testnet-faucet.onflow.org/fund-account?address=%s", address)
}

func init() {
	addContractCommand.AddToParent(Cmd)
	removeCommand.AddToParent(Cmd)
	updateCommand.AddToParent(Cmd)
	createCommand.AddToParent(Cmd)
	stakingCommand.AddToParent(Cmd)
	getCommand.AddToParent(Cmd)
	fundCommand.AddToParent(Cmd)
	listCommand.AddToParent(Cmd)
}

// accountResult represent result from all account commands.
type accountResult struct {
	*flow.Account
	include []string
}

func (r *accountResult) JSON() any {
	result := make(map[string]any)
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

func (r *accountResult) String() string {
	var b bytes.Buffer
	writer := util.CreateTabWriter(&b)

	if r.Address.IsValid(flow.Testnet) {
		_, _ = fmt.Fprintf(
			writer,
			"%s %s\n\n",
			branding.GrayStyle.Render("If you would like to fund the account with 1000 FLOW tokens for testing, visit"),
			branding.PurpleStyle.Render(testnetFaucetURL(r.Address)),
		)
	}

	_, _ = fmt.Fprintf(writer, "%s\t %s\n", branding.GrayStyle.Render("Address"), branding.PurpleStyle.Render("0x"+r.Address.String()))
	_, _ = fmt.Fprintf(writer, "%s\t %s\n", branding.GrayStyle.Render("Balance"), branding.GreenStyle.Render(cadence.UFix64(r.Balance).String()))

	_, _ = fmt.Fprintf(writer, "%s\t %d\n", branding.GrayStyle.Render("Keys"), len(r.Keys))

	for i, key := range r.Keys {
		_, _ = fmt.Fprintf(writer, "\n%s %d\t%s\t %x\n", branding.GrayStyle.Render("Key"), i, branding.GrayStyle.Render("Public Key"), key.PublicKey.Encode())
		_, _ = fmt.Fprintf(writer, "\t%s\t %d\n", branding.GrayStyle.Render("Weight"), key.Weight)
		_, _ = fmt.Fprintf(writer, "\t%s\t %s\n", branding.GrayStyle.Render("Signature Algorithm"), key.SigAlgo)
		_, _ = fmt.Fprintf(writer, "\t%s\t %s\n", branding.GrayStyle.Render("Hash Algorithm"), key.HashAlgo)
		_, _ = fmt.Fprintf(writer, "\t%s \t %t\n", branding.GrayStyle.Render("Revoked"), key.Revoked)
		_, _ = fmt.Fprintf(writer, "\t%s \t %d\n", branding.GrayStyle.Render("Sequence Number"), key.SequenceNumber)
		_, _ = fmt.Fprintf(writer, "\t%s \t %d\n", branding.GrayStyle.Render("Index"), key.Index)
		_, _ = fmt.Fprintf(writer, "\n")

		// only show up to 3 keys and then show label to expand more info
		if i == 3 && !command.ContainsFlag(r.include, "keys") {
			_, _ = fmt.Fprint(writer, branding.GrayStyle.Render("...keys minimized, use --include keys flag if you want to view all\n\n"))
			break
		}
	}

	_, _ = fmt.Fprintf(writer, "%s %d\n", branding.GrayStyle.Render("Contracts Deployed:"), len(r.Contracts))
	for name := range r.Contracts {
		_, _ = fmt.Fprintf(writer, "%s %s\n", branding.GrayStyle.Render("Contract:"), branding.PurpleStyle.Render("'"+name+"'"))
	}

	if command.ContainsFlag(r.include, "contracts") {
		for name, code := range r.Contracts {
			_, _ = fmt.Fprintf(writer, "%s %s:\n", branding.GrayStyle.Render("Contracts"), branding.PurpleStyle.Render("'"+name+"'"))
			_, _ = fmt.Fprintln(writer, string(code))
		}
	} else {
		_, _ = fmt.Fprint(writer, "\n\n"+branding.GrayStyle.Render("Contracts (hidden, use --include contracts)"))
	}

	_ = writer.Flush()

	return b.String()
}

func (r *accountResult) Oneliner() string {
	keys := make([]string, 0, len(r.Keys))
	for _, key := range r.Keys {
		keys = append(keys, key.PublicKey.String())
	}

	return fmt.Sprintf("Address: 0x%s, Balance: %s, Public Keys: %s", r.Address, cadence.UFix64(r.Balance), keys)
}
