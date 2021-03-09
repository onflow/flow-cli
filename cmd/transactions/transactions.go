package transactions

import (
	"bytes"
	"fmt"
	"text/tabwriter"

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
}

// JSON convert result to JSON
func (r *TransactionResult) JSON() interface{} {
	return r
}

// String convert result to string
func (r *TransactionResult) String() string {
	var b bytes.Buffer
	writer := tabwriter.NewWriter(&b, 0, 8, 1, '\t', tabwriter.AlignRight)

	fmt.Fprintf(writer, "Hash\t %s\n", r.tx.ID())
	fmt.Fprintf(writer, "Status\t %s\n", r.result.Status)
	fmt.Fprintf(writer, "Events\t %s\n", r.result.Events)

	writer.Flush()
	return b.String()
}

// Oneliner show result as one liner grep friendly
func (r *TransactionResult) Oneliner() string {
	return fmt.Sprintf("Hash: %s, Status: %s, Events: %s", r.tx.ID(), r.result.Status, r.result.Events)
}
