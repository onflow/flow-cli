/*
 * Flow CLI
 *
 * Copyright 2019-2020 Dapper Labs, Inc.
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

package project

import (
	"bytes"
	"fmt"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"github.com/onflow/flow-cli/flow/project/commands/deploy_contracts"
)

var Cmd = &cobra.Command{
	Use:              "project",
	Short:            "Manage your Cadence project",
	TraverseChildren: true,
}

func init() {
	Cmd.AddCommand(deploy_contracts.Cmd)
}

type ProjectResult struct {
}

// JSON convert result to JSON
func (r *ProjectResult) JSON() interface{} {
	return r
}

// String convert result to string
func (r *ProjectResult) String() string {
	var b bytes.Buffer
	writer := tabwriter.NewWriter(&b, 0, 8, 1, '\t', tabwriter.AlignRight)
	writer.Flush()
	return b.String()
}

// Oneliner show result as one liner grep friendly
func (r *ProjectResult) Oneliner() string {
	return fmt.Sprintf("%s", "")
}
