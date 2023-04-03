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
	"github.com/onflow/flow-go-sdk"
	"github.com/spf13/cobra"

	"github.com/onflow/flow-cli/internal/command"
	"github.com/onflow/flow-cli/pkg/flowkit"
	"github.com/onflow/flow-cli/pkg/flowkit/output"
	"github.com/onflow/flow-cli/pkg/flowkit/util"
)

type flagsScripts struct {
	ArgsJSON    string `default:"" flag:"args-json" info:"arguments in JSON-Cadence format"`
	BlockID     string `default:"" flag:"block-id" info:"block ID to execute the script at"`
	BlockHeight uint64 `default:"" flag:"block-height" info:"block height to execute the script at"`
}

var scriptFlags = flagsScripts{}

var ExecuteCommand = &command.Command{
	Cmd: &cobra.Command{
		Use:     "execute <filename> [<argument> <argument> ...]",
		Short:   "Execute a script",
		Example: `flow scripts execute script.cdc "Meow" "Woof"`,
		Args:    cobra.MinimumNArgs(1),
	},
	Flags: &scriptFlags,
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

	var scriptArgs []cadence.Value
	if scriptFlags.ArgsJSON != "" {
		scriptArgs, err = flowkit.ParseArgumentsJSON(scriptFlags.ArgsJSON)
	} else {
		scriptArgs, err = flowkit.ParseArgumentsWithoutType(filename, code, args[1:])
	}

	if err != nil {
		return nil, fmt.Errorf("error parsing script arguments: %w", err)
	}

	value, err := flow.ExecuteScript(
		context.Background(),
		flowkit.NewScript(code, scriptArgs, filename),
		&util.ScriptQuery{
			ID:     flow.HexToID(scriptFlags.BlockID),
			Height: scriptFlags.BlockHeight,
		},
	)
	if err != nil {
		return nil, err
	}

	return &ScriptResult{value}, nil
}
