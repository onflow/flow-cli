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

package events

import (
	"bytes"
	"fmt"
	"io"
	"text/tabwriter"

	"github.com/onflow/cadence"
	"github.com/onflow/flow-go-sdk"
	"github.com/onflow/flow-go-sdk/client"
	"github.com/spf13/cobra"

	"github.com/onflow/flow-cli/pkg/flowcli/util"
)

var Cmd = &cobra.Command{
	Use:              "events",
	Short:            "Utilities to read events",
	TraverseChildren: true,
}

func init() {
	GetCommand.AddToParent(Cmd)
}

// EventResult result structure
type EventResult struct {
	BlockEvents []client.BlockEvents
	Events      []flow.Event
}

// JSON convert result to JSON
func (k *EventResult) JSON() interface{} {
	result := make(map[string]map[uint64]map[string]interface{})
	for _, blockEvent := range k.BlockEvents {
		if len(blockEvent.Events) > 0 {
			for _, event := range blockEvent.Events {
				result["blockId"][blockEvent.Height]["index"] = event.EventIndex
				result["blockId"][blockEvent.Height]["type"] = event.Type
				result["blockId"][blockEvent.Height]["transactionId"] = event.TransactionID
				result["blockId"][blockEvent.Height]["values"] = event.Value
			}
		}
	}

	return result
}

// String convert result to string
func (k *EventResult) String() string {
	var b bytes.Buffer
	writer := tabwriter.NewWriter(&b, 0, 8, 1, '\t', tabwriter.AlignRight)

	for _, blockEvent := range k.BlockEvents {
		if len(blockEvent.Events) > 0 {
			fmt.Fprintf(writer, "Events Block #%v:", blockEvent.Height)
			eventsString(writer, blockEvent.Events)
			fmt.Fprintf(writer, "\n")
		}
	}

	// if we have events passed directly and not in relation to block
	eventsString(writer, k.Events)

	writer.Flush()
	return b.String()
}

// Oneliner show result as one liner grep friendly
func (k *EventResult) Oneliner() string {
	result := ""
	for _, blockEvent := range k.BlockEvents {
		if len(blockEvent.Events) > 0 {
			result += fmt.Sprintf("Events Block #%v: [", blockEvent.Height)
			for _, event := range blockEvent.Events {
				result += fmt.Sprintf(
					"Index: %v, Type: %v, TxID: %s, Value: %v",
					event.EventIndex, event.Type, event.TransactionID, event.Value,
				)
			}
			result += "] "
		}
	}

	return result
}

func eventsString(writer io.Writer, events []flow.Event) {
	for _, event := range events {
		eventString(writer, event)
	}
}

func eventString(writer io.Writer, event flow.Event) {
	fmt.Fprintf(writer, "\n\t Index\t %v\n", event.EventIndex)
	fmt.Fprintf(writer, "\t Type\t %s\n", event.Type)
	fmt.Fprintf(writer, "\t Tx ID\t %s\n", event.TransactionID)
	fmt.Fprintf(writer, "\t Values\n")

	for i, field := range event.Value.EventType.Fields {
		value := event.Value.Fields[i]
		printField(writer, field, value)
	}
}

func printField(writer io.Writer, field cadence.Field, value cadence.Value) {
	v := value.ToGoValue()
	typeInfo := "Unknown"

	if field.Type != nil {
		typeInfo = field.Type.ID()
	} else if _, isAddress := v.([8]byte); isAddress {
		typeInfo = "Address"
	}

	fmt.Fprintf(writer, "\t\t")
	fmt.Fprintf(writer, " %s (%s)\t", field.Identifier, typeInfo)
	// Try the two most obvious cases
	if address, ok := v.([8]byte); ok {
		fmt.Fprintf(writer, "%x", address)
	} else if util.IsByteSlice(v) || field.Identifier == "publicKey" {
		// make exception for public key, since it get's interpreted as []*big.Int
		for _, b := range v.([]interface{}) {
			fmt.Fprintf(writer, "%x", b)
		}
	} else if uintVal, ok := v.(uint64); typeInfo == "UFix64" && ok {
		fmt.Fprintf(writer, "%v", cadence.UFix64(uintVal))
	} else {
		fmt.Fprintf(writer, "%v", v)
	}
	fmt.Fprintf(writer, "\n")
}
