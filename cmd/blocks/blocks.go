package blocks

import (
	"bytes"
	"fmt"
	"text/tabwriter"

	"github.com/onflow/flow-go-sdk"
	"github.com/onflow/flow-go-sdk/client"
	"github.com/spf13/cobra"
)

var Cmd = &cobra.Command{
	Use:              "blocks",
	Short:            "Utilities to read blocks",
	TraverseChildren: true,
}

type BlockResult struct {
	block   *flow.Block
	events  []client.BlockEvents
	verbose bool
}

// JSON convert result to JSON
func (r *BlockResult) JSON() interface{} {
	return r
}

// String convert result to string
func (r *BlockResult) String() string {
	var b bytes.Buffer
	writer := tabwriter.NewWriter(&b, 0, 8, 1, '\t', tabwriter.AlignRight)

	fmt.Fprintf(writer, "Block ID\t%s\n", r.block.ID)
	fmt.Fprintf(writer, "Prent ID\t%s\n", r.block.ParentID)
	fmt.Fprintf(writer, "Timestamp\t%s\n", r.block.Timestamp)
	fmt.Fprintf(writer, "Height\t%v\n", r.block.Height)

	fmt.Fprintf(writer, "Total Seals\t%v\n", len(r.block.Seals))

	fmt.Fprintf(writer, "Total Collections\t%v\n", len(r.block.CollectionGuarantees))

	for i, guarantee := range r.block.CollectionGuarantees {
		fmt.Fprintf(writer, "    Collection %d:\t%s\n", i, guarantee.CollectionID)

		/* todo:
		if r.verbose {
			collection := cli.GetCollectionByID(conf.Host, guarantee.CollectionID)
			for i, transaction := range collection.TransactionIDs {
				fmt.Printf("    Transaction %d: %s\n", i, transaction)
			}
		}
		*/
	}

	writer.Flush()
	return b.String()
}

// Oneliner show result as one liner grep friendly
func (r *BlockResult) Oneliner() string {
	return fmt.Sprintf("%s", r.block.ID)
}
