/*
 * Flow CLI
 *
 * Copyright 2022 Dapper Labs, Inc.
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

package tests

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/onflow/flow-cli/internal/command"

	"github.com/spf13/cobra"

	"github.com/onflow/flow-cli/pkg/flowkit/util"

	"github.com/onflow/cadence/test-framework"
)

var Cmd = &cobra.Command{
	Use:              "tests",
	Short:            "Utilities to run tests",
	TraverseChildren: true,
}

func init() {
	ExecuteCommand.AddToParent(Cmd)
}

var _ command.Result = &TestResult{}

type TestResult struct {
	test_framework.Results
}

func (r *TestResult) JSON() interface{} {
	return json.RawMessage(
		r.Oneliner(),
	)
}

func (r *TestResult) String() string {
	var b bytes.Buffer
	writer := util.CreateTabWriter(&b)

	_, _ = fmt.Fprintf(writer, test_framework.PrettyPrintResults(r.Results))

	_ = writer.Flush()

	return b.String()
}

func (r *TestResult) Oneliner() string {
	return test_framework.PrettyPrintResults(r.Results)
}
