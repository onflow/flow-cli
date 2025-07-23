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

	"github.com/onflow/flow-cli/internal/command"
	"github.com/onflow/flow-cli/internal/events"
	"github.com/onflow/flow-cli/internal/util"

	jsoncdc "github.com/onflow/cadence/encoding/json"
)

var Cmd = &cobra.Command{
	Use:              "transactions",
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

	// Add block explorer link for mainnet and testnet
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
		_, _ = fmt.Fprintf(writer, "Block ID\t%s\n", r.result.BlockID)
		_, _ = fmt.Fprintf(writer, "Block Height\t%d\n", r.result.BlockHeight)
		if r.result.Error != nil {
			_, _ = fmt.Fprintf(writer, "%s Transaction Error \n%s\n\n\n", output.ErrorEmoji(), r.result.Error.Error())
		}

		statusBadge := ""
		if r.result.Status == flow.TransactionStatusSealed {
			statusBadge = output.OkEmoji()
		}
		_, _ = fmt.Fprintf(writer, "Status\t%s %s\n", statusBadge, r.result.Status)
	}

	_, _ = fmt.Fprintf(writer, "ID\t%s\n", r.tx.ID())

	// Add block explorer link for mainnet and testnet
	if blockExplorerLink := r.getBlockExplorerLink(); blockExplorerLink != "" {
		_, _ = fmt.Fprintf(writer, "🔗 View on Block Explorer\t%s\n", blockExplorerLink)
	}

	_, _ = fmt.Fprintf(writer, "Payer\t%s\n", r.tx.Payer.Hex())
	_, _ = fmt.Fprintf(writer, "Authorizers\t%s\n", r.tx.Authorizers)

	_, _ = fmt.Fprintf(writer,
		"\nProposal Key:\t\n    Address\t%s\n    Index\t%v\n    Sequence\t%v\n",
		r.tx.ProposalKey.Address, r.tx.ProposalKey.KeyIndex, r.tx.ProposalKey.SequenceNumber,
	)

	if len(r.tx.PayloadSignatures) == 0 {
		_, _ = fmt.Fprintf(writer, "\nNo Payload Signatures\n")
	}

	if len(r.tx.EnvelopeSignatures) == 0 {
		_, _ = fmt.Fprintf(writer, "\nNo Envelope Signatures\n")
	}

	for i, e := range r.tx.PayloadSignatures {
		if command.ContainsFlag(r.include, "signatures") {
			_, _ = fmt.Fprintf(writer, "\nPayload Signature %v:\n", i)
			_, _ = fmt.Fprintf(writer, "    Address\t%s\n", e.Address)
			_, _ = fmt.Fprintf(writer, "    Signature\t%x\n", e.Signature)
			_, _ = fmt.Fprintf(writer, "    Key Index\t%d\n", e.KeyIndex)
		} else {
			_, _ = fmt.Fprintf(writer, "\nPayload Signature %v: %s", i, e.Address)
		}
	}

	for i, e := range r.tx.EnvelopeSignatures {
		if command.ContainsFlag(r.include, "signatures") {
			_, _ = fmt.Fprintf(writer, "\nEnvelope Signature %v:\n", i)
			_, _ = fmt.Fprintf(writer, "    Address\t%s\n", e.Address)
			_, _ = fmt.Fprintf(writer, "    Signature\t%x\n", e.Signature)
			_, _ = fmt.Fprintf(writer, "    Key Index\t%d\n", e.KeyIndex)
		} else {
			_, _ = fmt.Fprintf(writer, "\nEnvelope Signature %v: %s", i, e.Address)
		}
	}

	if !command.ContainsFlag(r.include, "signatures") {
		_, _ = fmt.Fprintf(writer, "\nSignatures (minimized, use --include signatures)")
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

		_, _ = fmt.Fprintf(writer, "\n\nEvents:\t %s\n", eventsOutput)
	}

	if r.tx.Script != nil {
		if command.ContainsFlag(r.include, "code") {
			if len(r.tx.Arguments) == 0 {
				_, _ = fmt.Fprintf(writer, "\n\nArguments\tNo arguments\n")
			} else {
				_, _ = fmt.Fprintf(writer, "\n\nArguments (%d):\n", len(r.tx.Arguments))
				for i, argument := range r.tx.Arguments {
					_, _ = fmt.Fprintf(writer, "    - Argument %d: %s\n", i, argument)
				}
			}

			_, _ = fmt.Fprintf(writer, "\nCode\n\n%s\n", r.tx.Script)
		} else {
			_, _ = fmt.Fprint(writer, "\n\nCode (hidden, use --include code)")
		}
	}

	if command.ContainsFlag(r.include, "payload") {
		_, _ = fmt.Fprintf(writer, "\n\nPayload:\n%x", r.tx.Encode())
	} else {
		_, _ = fmt.Fprint(writer, "\n\nPayload (hidden, use --include payload)")
	}

	if !command.ContainsFlag(r.include, "fee-events") && !command.ContainsFlag(r.exclude, "events") {
		_, _ = fmt.Fprint(writer, "\n\nFee Events (hidden, use --include fee-events)")
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
