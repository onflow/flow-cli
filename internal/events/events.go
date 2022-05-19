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

package events

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"

	"github.com/onflow/flow-cli/pkg/flowkit/util"

	"github.com/onflow/cadence"
	jsoncdc "github.com/onflow/cadence/encoding/json"
	"github.com/onflow/flow-go-sdk"
	"github.com/onflow/flow-go-sdk/client"
	"github.com/spf13/cobra"
)

var Cmd = &cobra.Command{
	Use:              "events",
	Short:            "Utilities to read events",
	TraverseChildren: true,
}

func init() {
	GetCommand.AddToParent(Cmd)
}

type EventResult struct {
	BlockEvents []client.BlockEvents
	Events      []flow.Event
}

func (e *EventResult) JSON() interface{} {
	result := make([]interface{}, 0)

	for _, blockEvent := range e.BlockEvents {
		if len(blockEvent.Events) > 0 {
			for _, event := range blockEvent.Events {
				result = append(result, map[string]interface{}{
					"blockID":       blockEvent.Height,
					"index":         event.EventIndex,
					"type":          event.Type,
					"transactionId": event.TransactionID.String(),
					"values": json.RawMessage(
						jsoncdc.MustEncode(event.Value),
					),
				})
			}
		}
	}

	return result
}

func (e *EventResult) String() string {
	var b bytes.Buffer
	writer := util.CreateTabWriter(&b)

	for _, blockEvent := range e.BlockEvents {
		if len(blockEvent.Events) > 0 {
			_, _ = fmt.Fprintf(writer, "Events Block #%v:", blockEvent.Height)
			eventsString(writer, blockEvent.Events)
			_, _ = fmt.Fprintf(writer, "\n")
		}
	}

	// if we have events passed directly and not in relation to block
	eventsString(writer, e.Events)

	_ = writer.Flush()
	return b.String()
}

func (e *EventResult) Oneliner() string {
	result := ""
	for _, blockEvent := range e.BlockEvents {
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
	_, _ = fmt.Fprintf(writer, "\n    Index\t%d\n", event.EventIndex)
	_, _ = fmt.Fprintf(writer, "    Type\t%s\n", event.Type)
	_, _ = fmt.Fprintf(writer, "    Tx ID\t%s\n", event.TransactionID)
	_, _ = fmt.Fprintf(writer, "    Values\n")

	for i, field := range event.Value.EventType.Fields {
		value := event.Value.Fields[i]
		printField(writer, field, value)
	}
}

func printField(writer io.Writer, field cadence.Field, value cadence.Value) {
	var typeId string
	if field.Type != nil {
		typeId = field.Type.ID()
	}

	v := value.String()
	if typeId == "" { // exception for not known typeId workaround for cadence arrays
		v = fmt.Sprintf("%s\n\t\thex: %x", v, v)
	}

	_, _ = fmt.Fprintf(writer, "\t\t- %s (%s): %s \n", field.Identifier, typeId, v)
}
