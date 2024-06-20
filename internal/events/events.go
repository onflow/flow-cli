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

	"github.com/onflow/cadence"
	jsoncdc "github.com/onflow/cadence/encoding/json"
	"github.com/onflow/flow-go-sdk"
	"github.com/spf13/cobra"

	"github.com/onflow/flow-cli/internal/util"
)

var Cmd = &cobra.Command{
	Use:              "events",
	Short:            "Retrieve events",
	TraverseChildren: true,
	GroupID:          "resources",
}

func init() {
	getCommand.AddToParent(Cmd)
}

type EventResult struct {
	BlockEvents []flow.BlockEvents
	Events      []flow.Event
}

func (e *EventResult) JSON() any {
	result := make([]any, 0)

	for _, blockEvent := range e.BlockEvents {
		if len(blockEvent.Events) > 0 {
			for _, event := range blockEvent.Events {
				result = append(result, map[string]any{
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

	fields := cadence.FieldsMappedByName(event.Value)

	for _, field := range event.Value.EventType.Fields {
		value := fields[field.Identifier]
		printField(writer, field, value)
	}
}

func printValues(writer io.Writer, fieldIdentifier, typedId, valueString string) {
	_, _ = fmt.Fprintf(writer, "\t\t- %s (%s): %s \n", fieldIdentifier, typedId, valueString)
}

func printField(writer io.Writer, field cadence.Field, value cadence.Value) {
	v := value.String()
	var typeId string

	defer func() {
		if err := recover(); err != nil {
			printValues(writer, field.Identifier, "?", v)
		}
	}()

	if field.Type != nil {
		//TODO: onflow/cadence issue #1672
		//currently getting ID for cadence array will cause panic
		typeId = field.Type.ID()
	}

	if typeId == "" { // exception for not known typeId workaround for cadence arrays
		v = fmt.Sprintf("%s\n\t\thex: %x", v, v)
		typeId = "?"
	}
	printValues(writer, field.Identifier, typeId, v)
}
