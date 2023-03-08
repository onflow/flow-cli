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

package test

import (
	"bytes"
	"context"
	"fmt"

	cdcTests "github.com/onflow/cadence-tools/test"
	"github.com/spf13/cobra"

	"github.com/onflow/flow-cli/internal/command"
	"github.com/onflow/flow-cli/pkg/flowkit"
	"github.com/onflow/flow-cli/pkg/flowkit/output"
	"github.com/onflow/flow-cli/pkg/flowkit/util"
)

type flagsTests struct {
	// Nothing for now
}

var testFlags = flagsTests{}

var TestCommand = &command.Command{
	Cmd: &cobra.Command{
		Use:     "test <filename>",
		Short:   "Run Cadence tests",
		Example: `flow test script.cdc`,
		Args:    cobra.MinimumNArgs(1),
		GroupID: "tools",
	},
	Flags: &testFlags,
	Run:   run,
}

func run(
	args []string,
	_ command.GlobalFlags,
	logger output.Logger,
	readerWriter flowkit.ReaderWriter,
	services flowkit.Services,
) (command.Result, error) {
	filename := args[0]

	code, err := readerWriter.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("error loading script file: %w", err)
	}

	logger.StartProgress("Running tests...")
	defer logger.StopProgress()

	result, err := services.Test(
		context.Background(),
		code,
		filename,
	)
	if err != nil {
		return nil, err
	}

	return &TestResult{
		Results: result,
	}, nil
}

type TestResult struct {
	cdcTests.Results
}

var _ command.Result = &TestResult{}

func (r *TestResult) JSON() any {
	results := make([]map[string]string, 0, len(r.Results))

	for _, result := range r.Results {
		results = append(results, map[string]string{
			"testName": result.TestName,
			"error":    result.Error.Error(),
		})
	}

	return results
}

func (r *TestResult) String() string {
	var b bytes.Buffer
	writer := util.CreateTabWriter(&b)

	_, _ = fmt.Fprintf(writer, cdcTests.PrettyPrintResults(r.Results))

	_ = writer.Flush()

	return b.String()
}

func (r *TestResult) Oneliner() string {
	return cdcTests.PrettyPrintResults(r.Results)
}
