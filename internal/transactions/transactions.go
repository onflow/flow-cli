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

package transactions

import (
	"bytes"
	"fmt"
	"text/tabwriter"

	"github.com/onflow/flow-cli/internal/events"

	"github.com/onflow/flow-go-sdk"
	"github.com/spf13/cobra"
)

var Cmd = &cobra.Command{
	Use:              "transactions",
	Short:            "Utilities to send transactions",
	TraverseChildren: true,
}

func init() {
	GetCommand.AddToParent(Cmd)
	SendCommand.AddToParent(Cmd)
}

// TransactionResult represent result from all account commands
type TransactionResult struct {
	result *flow.TransactionResult
	tx     *flow.Transaction
	code   bool
}

// JSON convert result to JSON
func (r *TransactionResult) JSON() interface{} {
	result := make(map[string]string)
	result["Hash"] = r.tx.ID().String()
	result["Status"] = r.result.Status.String()

	if r.result != nil {
		result["Events"] = fmt.Sprintf("%s", r.result.Events)
	}

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
