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

package collections

import (
	"bytes"
	"fmt"
	"strings"
	"text/tabwriter"

	"github.com/onflow/flow-go-sdk"
	"github.com/spf13/cobra"
)

var Cmd = &cobra.Command{
	Use:              "collections",
	Short:            "Utilities to read collections",
	TraverseChildren: true,
}

func init() {
	GetCommand.AddToParent(Cmd)
}

// CollectionResult
type CollectionResult struct {
	*flow.Collection
}

// JSON convert result to JSON
func (c *CollectionResult) JSON() interface{} {
	txIDs := make([]string, 0)

	for _, tx := range c.Collection.TransactionIDs {
		txIDs = append(txIDs, tx.String())
	}

	return txIDs
}

// String convert result to string
func (c *CollectionResult) String() string {
	var b bytes.Buffer
	writer := tabwriter.NewWriter(&b, 0, 8, 1, '\t', tabwriter.AlignRight)

	fmt.Fprintf(writer, "Collection ID %s:\n", c.Collection.ID())

	for _, tx := range c.Collection.TransactionIDs {
		fmt.Fprintf(writer, "%s\n", tx.String())
	}

	writer.Flush()

	return b.String()
}

// Oneliner show result as one liner grep friendly
func (c *CollectionResult) Oneliner() string {
	return strings.Join(c.JSON().([]string), ",")
}
