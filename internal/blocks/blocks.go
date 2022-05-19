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

package blocks

import (
	"bytes"
	"fmt"

	"github.com/onflow/flow-cli/internal/command"

	"github.com/onflow/flow-go-sdk"
	"github.com/onflow/flow-go-sdk/client"
	"github.com/spf13/cobra"

	"github.com/onflow/flow-cli/internal/events"
	"github.com/onflow/flow-cli/pkg/flowkit/util"
)

var Cmd = &cobra.Command{
	Use:              "blocks",
	Short:            "Utilities to read blocks",
	TraverseChildren: true,
}

func init() {
	GetCommand.AddToParent(Cmd)
}

type BlockResult struct {
	block       *flow.Block
	events      []client.BlockEvents
	collections []*flow.Collection
	included    []string
}

func (r *BlockResult) JSON() interface{} {
	result := make(map[string]interface{})
	result["blockId"] = r.block.ID.String()
	result["parentId"] = r.block.ParentID.String()
	result["height"] = r.block.Height
	result["totalSeals"] = len(r.block.Seals)
	result["totalCollections"] = len(r.block.CollectionGuarantees)

	collections := make([]interface{}, 0, len(r.block.CollectionGuarantees))
	for i, guarantee := range r.block.CollectionGuarantees {
		collection := make(map[string]interface{})
		collection["id"] = guarantee.CollectionID.String()

		if command.ContainsFlag(r.included, "transactions") {
			txs := make([]string, 0)
			for _, tx := range r.collections[i].TransactionIDs {
				txs = append(txs, tx.String())
			}
			collection["transactions"] = txs
		}

		collections = append(collections, collection)
	}

	result["collection"] = collections
	return result
}

func (r *BlockResult) String() string {
	var b bytes.Buffer
	writer := util.CreateTabWriter(&b)

	_, _ = fmt.Fprintf(writer, "Block ID\t%s\n", r.block.ID)
	_, _ = fmt.Fprintf(writer, "Parent ID\t%s\n", r.block.ParentID)
	_, _ = fmt.Fprintf(writer, "Timestamp\t%s\n", r.block.Timestamp)
	_, _ = fmt.Fprintf(writer, "Height\t%v\n", r.block.Height)

	_, _ = fmt.Fprintf(writer, "Total Seals\t%v\n", len(r.block.Seals))

	_, _ = fmt.Fprintf(writer, "Total Collections\t%v\n", len(r.block.CollectionGuarantees))

	for i, guarantee := range r.block.CollectionGuarantees {
		_, _ = fmt.Fprintf(writer, "    Collection %d:\t%s\n", i, guarantee.CollectionID)

		if command.ContainsFlag(r.included, "transactions") {
			for x, tx := range r.collections[i].TransactionIDs {
				_, _ = fmt.Fprintf(writer, "         Transaction %d: %s\n", x, tx)
			}
		}
	}

	if len(r.events) > 0 {
		_, _ = fmt.Fprintf(writer, "\n")

		e := events.EventResult{BlockEvents: r.events}
		_, _ = fmt.Fprintf(writer, "%s", e.String())
	}

	_ = writer.Flush()
	return b.String()
}

func (r *BlockResult) Oneliner() string {
	return r.block.ID.String()
}
