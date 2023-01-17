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

package transactions

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/onflow/flow-cli/internal/command"
	"github.com/onflow/flow-cli/internal/events"
	"github.com/onflow/flow-cli/pkg/flowkit/output"
	"github.com/onflow/flow-cli/pkg/flowkit/util"

	"github.com/onflow/flow-go-sdk"

	"github.com/spf13/cobra"
)

var Cmd = &cobra.Command{
	Use:              "transactions",
	Short:            "Build, sign, send and retrieve transactions",
	TraverseChildren: true,
	GroupID:          "interactions",
}

func init() {
	GetCommand.AddToParent(Cmd)
	SendCommand.AddToParent(Cmd)
	SignCommand.AddToParent(Cmd)
	BuildCommand.AddToParent(Cmd)
	SendSignedCommand.AddToParent(Cmd)
	DecodeCommand.AddToParent(Cmd)
}

type TransactionResult struct {
	Result  *flow.TransactionResult
	Tx      *flow.Transaction
	include []string
	exclude []string
}

func (r *TransactionResult) JSON() interface{} {
	result := make(map[string]interface{})
	result["id"] = r.Tx.ID().String()
	result["payload"] = fmt.Sprintf("%x", r.Tx.Encode())
	result["authorizers"] = fmt.Sprintf("%s", r.Tx.Authorizers)
	result["payer"] = r.Tx.Payer.String()

	if r.Result != nil {
		result["status"] = r.Result.Status.String()

		txEvents := make([]interface{}, 0, len(r.Result.Events))
		for _, event := range r.Result.Events {
			txEvents = append(txEvents, map[string]interface{}{
				"index": event.EventIndex,
				"type":  event.Type,
				"values": json.RawMessage(
					event.Payload,
				),
			})
		}
		result["events"] = txEvents

		if r.Result.Error != nil {
			result["error"] = r.Result.Error.Error()
		}
	}

	return result
}

func (r *TransactionResult) String() string {
	var b bytes.Buffer
	writer := util.CreateTabWriter(&b)

	if r.Result != nil {
		if r.Result.Error != nil {
			_, _ = fmt.Fprintf(writer, "%s Transaction Error \n%s\n\n\n", output.ErrorEmoji(), r.Result.Error.Error())
		}

		statusBadge := ""
		if r.Result.Status == flow.TransactionStatusSealed {
			statusBadge = output.OkEmoji()
		}
		_, _ = fmt.Fprintf(writer, "Status\t%s %s\n", statusBadge, r.Result.Status)
	}

	_, _ = fmt.Fprintf(writer, "ID\t%s\n", r.Tx.ID())
	_, _ = fmt.Fprintf(writer, "Payer\t%s\n", r.Tx.Payer.Hex())
	_, _ = fmt.Fprintf(writer, "Authorizers\t%s\n", r.Tx.Authorizers)

	_, _ = fmt.Fprintf(writer,
		"\nProposal Key:\t\n    Address\t%s\n    Index\t%v\n    Sequence\t%v\n",
		r.Tx.ProposalKey.Address, r.Tx.ProposalKey.KeyIndex, r.Tx.ProposalKey.SequenceNumber,
	)

	if len(r.Tx.PayloadSignatures) == 0 {
		_, _ = fmt.Fprintf(writer, "\nNo Payload Signatures\n")
	}

	if len(r.Tx.EnvelopeSignatures) == 0 {
		_, _ = fmt.Fprintf(writer, "\nNo Envelope Signatures\n")
	}

	for i, e := range r.Tx.PayloadSignatures {
		if command.ContainsFlag(r.include, "signatures") {
			_, _ = fmt.Fprintf(writer, "\nPayload Signature %v:\n", i)
			_, _ = fmt.Fprintf(writer, "    Address\t%s\n", e.Address)
			_, _ = fmt.Fprintf(writer, "    Signature\t%x\n", e.Signature)
			_, _ = fmt.Fprintf(writer, "    Key Index\t%d\n", e.KeyIndex)
		} else {
			_, _ = fmt.Fprintf(writer, "\nPayload Signature %v: %s", i, e.Address)
		}
	}

	for i, e := range r.Tx.EnvelopeSignatures {
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

	if r.Result != nil && !command.ContainsFlag(r.exclude, "events") {
		e := events.EventResult{
			Events: r.Result.Events,
		}

		eventsOutput := e.String()
		if eventsOutput == "" {
			eventsOutput = "None"
		}

		_, _ = fmt.Fprintf(writer, "\n\nEvents:\t %s\n", eventsOutput)
	}

	if r.Tx.Script != nil {
		if command.ContainsFlag(r.include, "code") {
			if len(r.Tx.Arguments) == 0 {
				_, _ = fmt.Fprintf(writer, "\n\nArguments\tNo arguments\n")
			} else {
				_, _ = fmt.Fprintf(writer, "\n\nArguments (%d):\n", len(r.Tx.Arguments))
				for i, argument := range r.Tx.Arguments {
					_, _ = fmt.Fprintf(writer, "    - Argument %d: %s\n", i, argument)
				}
			}

			_, _ = fmt.Fprintf(writer, "\nCode\n\n%s\n", r.Tx.Script)
		} else {
			_, _ = fmt.Fprint(writer, "\n\nCode (hidden, use --include code)")
		}
	}

	if command.ContainsFlag(r.include, "payload") {
		_, _ = fmt.Fprintf(writer, "\n\nPayload:\n%x", r.Tx.Encode())
	} else {
		_, _ = fmt.Fprint(writer, "\n\nPayload (hidden, use --include payload)")
	}

	_ = writer.Flush()
	return b.String()
}

func (r *TransactionResult) Oneliner() string {
	result := fmt.Sprintf(
		"ID: %s, Payer: %s, Authorizer: %s",
		r.Tx.ID(), r.Tx.Payer, r.Tx.Authorizers)

	if r.Result != nil {
		result += fmt.Sprintf(", Status: %s, Events: %s", r.Result.Status, r.Result.Events)
	}

	return result
}
