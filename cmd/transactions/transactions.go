package transactions

import (
	"bytes"
	"fmt"
	"text/tabwriter"

	"github.com/onflow/flow-cli/cmd/events"

	"github.com/onflow/flow-go-sdk"
	"github.com/spf13/cobra"
)

var Cmd = &cobra.Command{
	Use:              "transactions",
	Short:            "Utilities to send transactions",
	TraverseChildren: true,
}

// TransactionResult represent result from all account commands
type TransactionResult struct {
	result *flow.TransactionResult
	tx     *flow.Transaction
	code   bool
}

// JSON convert result to JSON
func (r *TransactionResult) JSON() interface{} {
	result := make(map[string]string, 0)
	result["Hash"] = r.tx.ID().String()
	result["Status"] = r.result.Status.String()
	result["Events"] = fmt.Sprintf("%s", r.result.Events)

	return result
}

// String convert result to string
func (r *TransactionResult) String() string {
	var b bytes.Buffer
	writer := tabwriter.NewWriter(&b, 0, 8, 1, '\t', tabwriter.AlignRight)

	fmt.Fprintf(writer, "Hash\t %s\n", r.tx.ID())
	fmt.Fprintf(writer, "Status\t %s\n", r.result.Status)
	fmt.Fprintf(writer, "Payer\t %s\n", r.tx.Payer.Hex())

	events := events.EventResult{
		Events: r.result.Events,
	}
	fmt.Fprintf(writer, "Events\t %s\n", events.String())

	if r.code {
		fmt.Fprintf(writer, "Code\n\n%s\n", r.tx.Script)
	}

	writer.Flush()
	return b.String()
}

// Oneliner show result as one liner grep friendly
func (r *TransactionResult) Oneliner() string {
	return fmt.Sprintf("Hash: %s, Status: %s, Events: %s", r.tx.ID(), r.result.Status, r.result.Events)
}
