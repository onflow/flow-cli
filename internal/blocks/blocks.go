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

package blocks

import (
	"bytes"
	"fmt"
	"text/tabwriter"

	"github.com/onflow/flow-cli/internal/events"

	"github.com/onflow/flow-go-sdk"
	"github.com/onflow/flow-go-sdk/client"
	"github.com/spf13/cobra"
)

var Cmd = &cobra.Command{
	Use:              "blocks",
	Short:            "Utilities to read blocks",
	TraverseChildren: true,
}

func init() {
	GetCommand.AddToParent(Cmd)
}

// BlockResult
type BlockResult struct {
	block       *flow.Block
	events      []client.BlockEvents
	collections []*flow.Collection
	verbose     bool
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
	fmt.Fprintf(writer, "Parent ID\t%s\n", r.block.ParentID)
	fmt.Fprintf(writer, "Timestamp\t%s\n", r.block.Timestamp)
	fmt.Fprintf(writer, "Height\t%v\n", r.block.Height)

	fmt.Fprintf(writer, "Total Seals\t%v\n", len(r.block.Seals))

	fmt.Fprintf(writer, "Total Collections\t%v\n", len(r.block.CollectionGuarantees))

	for i, guarantee := range r.block.CollectionGuarantees {
		fmt.Fprintf(writer, "    Collection %d:\t%s\n", i, guarantee.CollectionID)

		if r.verbose {
			for x, tx := range r.collections[i].TransactionIDs {
				fmt.Fprintf(writer, "         Transaction %d: %s\n", x, tx)
			}
		}
	}

	if len(r.events) > 0 {
		fmt.Fprintf(writer, "\n")

		e := events.EventResult{BlockEvents: r.events}
		fmt.Fprintf(writer, "%s", e.String())
	}

	writer.Flush()
	return b.String()
}

// Oneliner show result as one liner grep friendly
func (r *BlockResult) Oneliner() string {
	return r.block.ID.String()
}
