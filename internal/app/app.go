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

package app

import (
	"bytes"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/onflow/flow-cli/pkg/flowkit/util"
)

var Cmd = &cobra.Command{
	Use:              "app",
	Short:            "Utilities to create Flow app",
	TraverseChildren: true,
}

func init() {
	CreateCommand.AddToParent(Cmd)
}

type CreateResult struct {
	created  string
	example  string
	template string
}

func (r *CreateResult) JSON() interface{} {
	result := make(map[string]interface{})
	result["created"] = r.created

	if r.example != "" {
		result["example"] = r.example
	} else {
		result["template"] = r.template
	}

	return result
}

func (r *CreateResult) String() string {
	var b bytes.Buffer
	writer := util.CreateTabWriter(&b)

	_, _ = fmt.Fprintf(writer, "Created\t %s\n", r.created)

	if r.example != "" {
		_, _ = fmt.Fprintf(writer, "Example\t %s\n", r.example)
	} else {
		_, _ = fmt.Fprintf(writer, "Template\t %s\n", r.template)
	}

	_ = writer.Flush()

	return b.String()
}

func (r *CreateResult) Oneliner() string {
	if r.example != "" {
		return fmt.Sprintf("Created: %s, Example: %s", r.created, r.example)
	} else {
		return fmt.Sprintf("Created: %s Template: %s", r.created, r.template)
	}
}
