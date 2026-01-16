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

package transactions

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/onflow/flow-go-sdk"
	"github.com/spf13/cobra"

	"github.com/onflow/flowkit/v2/output"

	"github.com/onflow/flow-cli/common/branding"
	"github.com/onflow/flow-cli/internal/command"
	"github.com/onflow/flow-cli/internal/events"
	"github.com/onflow/flow-cli/internal/util"

	jsoncdc "github.com/onflow/cadence/encoding/json"
)

var Cmd = &cobra.Command{
	Use:              "transactions",
	Aliases:          []string{"tx"},
	Short:            "Build, sign, send and retrieve transactions",
	TraverseChildren: true,
	GroupID:          "interactions",
}

func init() {
	getCommand.AddToParent(Cmd)
	sendCommand.AddToParent(Cmd)
	signCommand.AddToParent(Cmd)
	buildCommand.AddToParent(Cmd)
	sendSignedCommand.AddToParent(Cmd)
	getSystemCommand.AddToParent(Cmd)
	decodeCommand.AddToParent(Cmd)
	profileCommand.AddToParent(Cmd)
}

type transactionResult struct {
	result  *flow.TransactionResult
	tx      *flow.Transaction
	include []string
	exclude []string
	network string
}

func NewTransactionResult(tx *flow.Transaction, result *flow.TransactionResult) *transactionResult {
	return &transactionResult{
		result:  result,
		tx:      tx,
		network: "", // Default to empty, should be set by caller
	}
}

// getBlockExplorerLink returns the block explorer link for the transaction if it's on mainnet or testnet
func (r *transactionResult) getBlockExplorerLink() string {
	if r.network == "" {
		return ""
	}

	// Only show block explorer links for mainnet and testnet
	if r.network != "mainnet" && r.network != "testnet" {
		return ""
	}

	txID := r.tx.ID().String()

	if r.network == "mainnet" {
		return fmt.Sprintf("https://www.flowscan.io/tx/%s", txID)
	} else if r.network == "testnet" {
		return fmt.Sprintf("https://testnet.flowscan.io/tx/%s", txID)
	}

	return ""
}

func (r *transactionResult) JSON() any {
	result := make(map[string]any)
	result["id"] = r.tx.ID().String()
	result["payload"] = fmt.Sprintf("%x", r.tx.Encode())
	result["authorizers"] = fmt.Sprintf("%s", r.tx.Authorizers)
	result["payer"] = r.tx.Payer.String()

	if blockExplorerLink := r.getBlockExplorerLink(); blockExplorerLink != "" {
		result["view_on_block_explorer"] = blockExplorerLink
	}

	if r.result != nil {
		result["block_id"] = r.result.BlockID.String()
		result["block_height"] = r.result.BlockHeight
		result["status"] = r.result.Status.String()

		txEvents := make([]any, 0, len(r.result.Events))
		for _, event := range r.result.Events {
			txEvents = append(txEvents, map[string]any{
				"index": event.EventIndex,
				"type":  event.Type,
				"values": json.RawMessage(
					jsoncdc.MustEncode(event.Value),
				),
			})
		}
		result["events"] = txEvents

		if r.result.Error != nil {
			result["error"] = r.result.Error.Error()
		}
	}

	return result
}

func (r *transactionResult) String() string {
	var b bytes.Buffer
	writer := util.CreateTabWriter(&b)
	const feeEventsCountAppended = 5
	const feeDeductedEvent = "FeesDeducted"

	if r.result != nil {
		_, _ = fmt.Fprintf(writer, "%s\t%s\n", branding.GrayStyle.Render("Block ID"), branding.PurpleStyle.Render(r.result.BlockID.String()))
		_, _ = fmt.Fprintf(writer, "%s\t%d\n", branding.GrayStyle.Render("Block Height"), r.result.BlockHeight)
		if r.result.Error != nil {
			_, _ = fmt.Fprintf(writer, "%s %s\n%s\n\n\n", output.ErrorEmoji(), branding.GrayStyle.Render("Transaction Error"), branding.ErrorStyle.Render(r.result.Error.Error()))
		}

		statusBadge := ""
		statusText := r.result.Status.String()
		if r.result.Status == flow.TransactionStatusSealed {
			statusBadge = output.OkEmoji()
			statusText = branding.GreenStyle.Render(statusText)
		}
		// leave uncolored for non-sealed statuses
		_, _ = fmt.Fprintf(writer, "%s\t%s %s\n", branding.GrayStyle.Render("Status"), statusBadge, statusText)
	}

	_, _ = fmt.Fprintf(writer, "%s\t%s\n", branding.GrayStyle.Render("ID"), branding.PurpleStyle.Render(r.tx.ID().String()))

	_, _ = fmt.Fprintf(writer, "%s\t%s\n", branding.GrayStyle.Render("Payer"), branding.PurpleStyle.Render(r.tx.Payer.Hex()))
	_, _ = fmt.Fprintf(writer, "%s\t%s\n", branding.GrayStyle.Render("Authorizers"), branding.PurpleStyle.Render(fmt.Sprintf("%s", r.tx.Authorizers)))

	_, _ = fmt.Fprintf(writer,
		"\n%s\t\n    %s\t%s\n    %s\t%v\n    %s\t%v\n",
		branding.GrayStyle.Render("Proposal Key:"),
		branding.GrayStyle.Render("Address"), branding.PurpleStyle.Render(r.tx.ProposalKey.Address.String()),
		branding.GrayStyle.Render("Index"), r.tx.ProposalKey.KeyIndex,
		branding.GrayStyle.Render("Sequence"), r.tx.ProposalKey.SequenceNumber,
	)

	if len(r.tx.PayloadSignatures) == 0 {
		_, _ = fmt.Fprintf(writer, "\n%s\n", branding.GrayStyle.Render("No Payload Signatures"))
	}

	if len(r.tx.EnvelopeSignatures) == 0 {
		_, _ = fmt.Fprintf(writer, "\n%s\n", branding.GrayStyle.Render("No Envelope Signatures"))
	}

	for i, e := range r.tx.PayloadSignatures {
		if command.ContainsFlag(r.include, "signatures") {
			_, _ = fmt.Fprintf(writer, "\n%s %d:\n", branding.GrayStyle.Render("Payload Signature"), i)
			_, _ = fmt.Fprintf(writer, "    %s\t%s\n", branding.GrayStyle.Render("Address"), branding.PurpleStyle.Render(e.Address.String()))
			_, _ = fmt.Fprintf(writer, "    %s\t%x\n", branding.GrayStyle.Render("Signature"), e.Signature)
			_, _ = fmt.Fprintf(writer, "    %s\t%d\n", branding.GrayStyle.Render("Key Index"), e.KeyIndex)
		} else {
			_, _ = fmt.Fprintf(writer, "\n%s %d: %s", branding.GrayStyle.Render("Payload Signature"), i, branding.PurpleStyle.Render(e.Address.String()))
		}
	}

	for i, e := range r.tx.EnvelopeSignatures {
		if command.ContainsFlag(r.include, "signatures") {
			_, _ = fmt.Fprintf(writer, "\n%s %d:\n", branding.GrayStyle.Render("Envelope Signature"), i)
			_, _ = fmt.Fprintf(writer, "    %s\t%s\n", branding.GrayStyle.Render("Address"), branding.PurpleStyle.Render(e.Address.String()))
			_, _ = fmt.Fprintf(writer, "    %s\t%x\n", branding.GrayStyle.Render("Signature"), e.Signature)
			_, _ = fmt.Fprintf(writer, "    %s\t%d\n", branding.GrayStyle.Render("Key Index"), e.KeyIndex)
		} else {
			_, _ = fmt.Fprintf(writer, "\n%s %d: %s", branding.GrayStyle.Render("Envelope Signature"), i, branding.PurpleStyle.Render(e.Address.String()))
		}
	}

	if !command.ContainsFlag(r.include, "signatures") {
		_, _ = fmt.Fprintf(writer, "\n%s", branding.GrayStyle.Render("Signatures (minimized, use --include signatures)"))
	}

	if r.result != nil && !command.ContainsFlag(r.exclude, "events") {
		e := events.EventResult{
			Events: r.result.Events,
		}

		if r.result != nil && e.Events != nil && !command.ContainsFlag(r.include, "fee-events") {
			for _, event := range e.Events {
				if strings.Contains(event.Type, feeDeductedEvent) {
					// if fee event are present remove them
					e.Events = e.Events[:len(e.Events)-feeEventsCountAppended]
					break
				}
			}
		}

		eventsOutput := e.String()
		if eventsOutput == "" {
			eventsOutput = "None"
		}

		_, _ = fmt.Fprintf(writer, "\n\n%s\t %s\n", branding.GrayStyle.Render("Events:"), eventsOutput)
	}

	if r.tx.Script != nil {
		if command.ContainsFlag(r.include, "code") {
			if len(r.tx.Arguments) == 0 {
				_, _ = fmt.Fprintf(writer, "\n\n%s\tNo arguments\n", branding.GrayStyle.Render("Arguments"))
			} else {
				_, _ = fmt.Fprintf(writer, "\n\n%s (%d)\n", branding.GrayStyle.Render("Arguments"), len(r.tx.Arguments))
				for i, argument := range r.tx.Arguments {
					_, _ = fmt.Fprintf(writer, "    - %s %d: %s\n", branding.GrayStyle.Render("Argument"), i, string(argument))
				}
			}

			_, _ = fmt.Fprintf(writer, "\n%s\n\n%s\n", branding.GrayStyle.Render("Code"), string(r.tx.Script))
		} else {
			_, _ = fmt.Fprint(writer, "\n\n"+branding.GrayStyle.Render("Code (hidden, use --include code)"))
		}
	}

	if command.ContainsFlag(r.include, "payload") {
		_, _ = fmt.Fprintf(writer, "\n\n%s\n%x", branding.GrayStyle.Render("Payload"), r.tx.Encode())
	} else {
		_, _ = fmt.Fprint(writer, "\n\n"+branding.GrayStyle.Render("Payload (hidden, use --include payload)"))
	}

	if !command.ContainsFlag(r.include, "fee-events") && !command.ContainsFlag(r.exclude, "events") {
		_, _ = fmt.Fprint(writer, "\n\n"+branding.GrayStyle.Render("Fee Events (hidden, use --include fee-events)"))
	}

	if blockExplorerLink := r.getBlockExplorerLink(); blockExplorerLink != "" {
		_, _ = fmt.Fprintf(writer, "\n\n%s\n%s", branding.GrayStyle.Render("🔗 View on Block Explorer:"), branding.PurpleStyle.Render(blockExplorerLink))
	}

	_ = writer.Flush()
	return b.String()
}

func (r *transactionResult) Oneliner() string {
	result := fmt.Sprintf(
		"ID: %s, Payer: %s, Authorizer: %s",
		r.tx.ID(), r.tx.Payer, r.tx.Authorizers)

	if r.result != nil {
		result += fmt.Sprintf(", Status: %s, Events: %s", r.result.Status, r.result.Events)
	}

	return result
}
