package events

import (
	"bytes"
	"fmt"
	"io"
	"text/tabwriter"

	"github.com/onflow/flow-cli/flow/util"

	"github.com/onflow/cadence"
	"github.com/onflow/flow-go-sdk"
	"github.com/onflow/flow-go-sdk/client"
	"github.com/spf13/cobra"
)

var Cmd = &cobra.Command{
	Use:              "events",
	Short:            "Utilities to get events",
	TraverseChildren: true,
}

// KeyResult represent result from all account commands
type EventResult struct {
	BlockEvents []client.BlockEvents
	Events      []flow.Event
}

// JSON convert result to JSON
func (k *EventResult) JSON() interface{} {
	result := make(map[string]string, 0)
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
	return fmt.Sprintf("")
}

func (k *EventResult) ToConfig() string {
	return ""
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
