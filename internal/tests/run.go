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
	"fmt"

	"github.com/spf13/cobra"

	"github.com/onflow/flow-cli/internal/command"
	"github.com/onflow/flow-cli/pkg/flowkit"
	"github.com/onflow/flow-cli/pkg/flowkit/services"
)

type flagsTests struct {
	// Nothing for now
}

var testFlags = flagsTests{}

var ExecuteCommand = &command.Command{
	Cmd: &cobra.Command{
		Use:     "run <filename>",
		Short:   "Run Cadence tests",
		Example: `flow test run script.cdc`,
		Args:    cobra.MinimumNArgs(1),
	},
	Flags: &testFlags,
	Run:   run,
}

func run(
	args []string,
	readerWriter flowkit.ReaderWriter,
	_ command.GlobalFlags,
	services *services.Services,
) (command.Result, error) {

	filename := args[0]

	code, err := readerWriter.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("error loading script file: %w", err)
	}

	result, err := services.Tests.Execute(
		code,
		filename,
		readerWriter,
	)

	if err != nil {
		return nil, err
	}

	return &TestResult{
		Results: result,
	}, nil
}
