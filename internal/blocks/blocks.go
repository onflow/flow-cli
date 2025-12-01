/*
 * Flow CLI
 *
 * Copyright Flow Foundation
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

	"github.com/onflow/flow-go-sdk"
	"github.com/spf13/cobra"

	"github.com/onflow/flow-cli/internal/command"
	"github.com/onflow/flow-cli/internal/events"
	"github.com/onflow/flow-cli/internal/util"
)

var Cmd = &cobra.Command{
	Use:              "blocks",
	Short:            "Retrieve blocks",
	TraverseChildren: true,
	GroupID:          "resources",
}

func init() {
	getCommand.AddToParent(Cmd)
}

type blockResult struct {
	block        *flow.Block
	events       []flow.BlockEvents
	transactions []*flow.Transaction
	results      []*flow.TransactionResult
	included     []string
}

func (r *blockResult) JSON() any {
	result := make(map[string]any)
	result["blockId"] = r.block.ID.String()
	result["parentId"] = r.block.ParentID.String()
	result["height"] = r.block.Height
	result["totalSeals"] = len(r.block.Seals)
	result["totalCollections"] = len(r.block.CollectionGuarantees)

	// Keep collection info for backwards compatibility
	collections := make([]any, 0, len(r.block.CollectionGuarantees))
	for _, guarantee := range r.block.CollectionGuarantees {
		collection := make(map[string]any)
		collection["id"] = guarantee.CollectionID.String()
		collections = append(collections, collection)
	}
	result["collections"] = collections

	// Add transaction details if requested
	if command.ContainsFlag(r.included, "transactions") && len(r.transactions) > 0 {
		txs := make([]map[string]any, 0, len(r.transactions))
		for i, tx := range r.transactions {
			txData := make(map[string]any)
			txData["id"] = tx.ID().String()
			txData["status"] = r.results[i].Status.String()

			// System transactions have empty collection ID
			if r.results[i].CollectionID == flow.EmptyID {
				txData["type"] = "system"
			} else {
				txData["type"] = "user"
				txData["collectionId"] = r.results[i].CollectionID.String()
			}

			txs = append(txs, txData)
		}
		result["transactions"] = txs
	}

	return result
}

func blockStatusToString(code flow.BlockStatus) string {
	switch code {
	case 1:
		return "Finalized"
	case 2:
		return "Sealed"
	default:
		return "Unknown"
	}
}

func (r *blockResult) String() string {
	var b bytes.Buffer
	writer := util.CreateTabWriter(&b)

	_, _ = fmt.Fprintf(writer, "Block ID\t%s\n", r.block.ID)
	_, _ = fmt.Fprintf(writer, "Parent ID\t%s\n", r.block.ParentID)
	_, _ = fmt.Fprintf(writer, "Proposal Timestamp\t%s\n", r.block.Timestamp)
	_, _ = fmt.Fprintf(writer, "Proposal Timestamp Unix\t%d\n", r.block.Timestamp.Unix())
	_, _ = fmt.Fprintf(writer, "Height\t%v\n", r.block.Height)
	_, _ = fmt.Fprintf(writer, "Status\t%s\n", blockStatusToString(r.block.Status))

	_, _ = fmt.Fprintf(writer, "Total Seals\t%v\n", len(r.block.Seals))
	_, _ = fmt.Fprintf(writer, "Total Collections\t%v\n", len(r.block.CollectionGuarantees))

	// Show collections
	for i, guarantee := range r.block.CollectionGuarantees {
		_, _ = fmt.Fprintf(writer, "    Collection %d:\t%s\n", i, guarantee.CollectionID)
	}

	// Show transactions if included
	if command.ContainsFlag(r.included, "transactions") && len(r.transactions) > 0 {
		_, _ = fmt.Fprintf(writer, "\nTransactions:\n")

		userCount := 0
		systemCount := 0

		for i, tx := range r.transactions {
			var txType string
			if r.results[i].CollectionID == flow.EmptyID {
				txType = "system"
				systemCount++
			} else {
				txType = "user"
				userCount++
			}

			_, _ = fmt.Fprintf(writer, "    [%d] %s\t%s (%s)\n",
				i,
				tx.ID().String(),
				r.results[i].Status.String(),
				txType,
			)
		}

		_, _ = fmt.Fprintf(writer, "\nTotal: %d transactions (%d user, %d system)\n",
			len(r.transactions), userCount, systemCount)
	}

	if len(r.events) > 0 {
		_, _ = fmt.Fprintf(writer, "\n")

		e := events.EventResult{BlockEvents: r.events}
		_, _ = fmt.Fprintf(writer, "%s", e.String())
	}

	_ = writer.Flush()
	return b.String()
}

func (r *blockResult) Oneliner() string {
	return r.block.ID.String()
}
