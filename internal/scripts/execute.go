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

package scripts

import (
	"context"
	"fmt"

	"github.com/onflow/cadence"
	flowsdk "github.com/onflow/flow-go-sdk"
	"github.com/spf13/cobra"

	"github.com/onflow/flowkit"
	"github.com/onflow/flowkit/arguments"
	"github.com/onflow/flowkit/output"

	"github.com/onflow/flow-cli/internal/command"
)

type Flags struct {
	ArgsJSON    string `default:"" flag:"args-json" info:"arguments in JSON-Cadence format"`
	BlockID     string `default:"" flag:"block-id" info:"block ID to execute the script at"`
	BlockHeight uint64 `default:"" flag:"block-height" info:"block height to execute the script at"`
}

var flags = Flags{}

var executeCommand = &command.Command{
	Cmd: &cobra.Command{
		Use:     "execute <filename> [<argument> <argument> ...]",
		Short:   "Execute a script",
		Example: `flow scripts execute script.cdc "Meow" "Woof"`,
		Args:    cobra.MinimumNArgs(1),
	},
	Flags: &flags,
	Run:   execute,
}

func execute(
	args []string,
	_ command.GlobalFlags,
	_ output.Logger,
	readerWriter flowkit.ReaderWriter,
	flow flowkit.Services,
) (command.Result, error) {
	filename := args[0]

	code, err := readerWriter.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("error loading script file: %w", err)
	}

	return SendScript(code, args[1:], filename, flow, flags)
}

func SendScript(code []byte, argsArr []string, location string, flow flowkit.Services, scriptFlags Flags) (command.Result, error) {
	var cadenceArgs []cadence.Value
	var err error
	if scriptFlags.ArgsJSON != "" {
		cadenceArgs, err = arguments.ParseJSON(scriptFlags.ArgsJSON)
	} else {
		cadenceArgs, err = arguments.ParseWithoutType(argsArr, code, location)
	}

	if err != nil {
		return nil, fmt.Errorf("error parsing script arguments: %w", err)
	}

	query := flowkit.ScriptQuery{}
	if scriptFlags.BlockHeight != 0 {
		query.Height = scriptFlags.BlockHeight
	} else if scriptFlags.BlockID != "" {
		query.ID = flowsdk.HexToID(scriptFlags.BlockID)
	} else {
		query.Latest = true
	}

	value, err := flow.ExecuteScript(
		context.Background(),
		flowkit.Script{
			Code:     code,
			Args:     cadenceArgs,
			Location: location,
		},
		query,
	)
	if err != nil {
		return nil, err
	}

	return &scriptResult{value}, nil
}
