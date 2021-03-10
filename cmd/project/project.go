package project

import (
	"bytes"
	"fmt"
	"text/tabwriter"

	"github.com/onflow/flow-go-sdk"
	"github.com/spf13/cobra"
)

var Cmd = &cobra.Command{
	Use:              "project",
	Short:            "Manage your Cadence project",
	TraverseChildren: true,
}

type ProjectResult struct {
	*flow.Collection
}

// JSON convert result to JSON
func (c *ProjectResult) JSON() interface{} {
	return c
}

// String convert result to string
func (c *ProjectResult) String() string {
	var b bytes.Buffer
	writer := tabwriter.NewWriter(&b, 0, 8, 1, '\t', tabwriter.AlignRight)
	fmt.Fprintf(writer, "%s\n", c.Collection)
	writer.Flush()

	return b.String()
}

// Oneliner show result as one liner grep friendly
func (c *ProjectResult) Oneliner() string {
	return fmt.Sprintf("%s", c.Collection)
}
