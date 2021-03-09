package scripts

import (
	"bytes"
	"fmt"
	"text/tabwriter"

	"github.com/onflow/cadence"
	"github.com/spf13/cobra"
)

var Cmd = &cobra.Command{
	Use:              "scripts",
	Short:            "Utilities to execute scripts",
	TraverseChildren: true,
}

type ScriptResult struct {
	cadence.Value
}

// JSON convert result to JSON
func (r *ScriptResult) JSON() interface{} {
	return r
}

// String convert result to string
func (r *ScriptResult) String() string {
	var b bytes.Buffer
	writer := tabwriter.NewWriter(&b, 0, 8, 1, '\t', tabwriter.AlignRight)

	fmt.Fprintf(writer, "Result: %s\n", r.Value)

	writer.Flush()

	return b.String()
}

// Oneliner show result as one liner grep friendly
func (r *ScriptResult) Oneliner() string {
	return fmt.Sprintf("%s", r.Value)
}
