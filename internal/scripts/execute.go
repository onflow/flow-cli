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
	"fmt"

	"github.com/onflow/cadence"
	"github.com/spf13/cobra"

	"github.com/onflow/flow-cli/internal/command"
	"github.com/onflow/flow-cli/pkg/flowkit"
	"github.com/onflow/flow-cli/pkg/flowkit/services"
)

type flagsScripts struct {
	ArgsJSON string   `default:"" flag:"args-json" info:"arguments in JSON-Cadence format"`
	Arg      []string `default:"" flag:"arg" info:"⚠️  Deprecated: use command arguments"`
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
	readerWriter flowkit.ReaderWriter,
	globalFlags command.GlobalFlags,
	services *services.Services,
) (command.Result, error) {
	filename := args[0]

	code, err := readerWriter.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("error loading script file: %w", err)
	}

	if len(scriptFlags.Arg) != 0 {
		fmt.Println("⚠️  DEPRECATION WARNING: use script arguments as command arguments: execute <filename> [<argument> <argument> ...]")
	}

	var scriptArgs []cadence.Value
	if scriptFlags.ArgsJSON != "" || len(scriptFlags.Arg) != 0 {
		scriptArgs, err = flowkit.ParseArguments(scriptFlags.Arg, scriptFlags.ArgsJSON)
	} else {
		scriptArgs, err = flowkit.ParseArgumentsWithoutType(filename, code, args[1:])
	}

	if err != nil {
		return nil, fmt.Errorf("error parsing script arguments: %w", err)
	}

	value, err := services.Scripts.Execute(
		code,
		scriptArgs,
		filename,
		globalFlags.Network,
	)
	if err != nil {
		return nil, err
	}

	return &ScriptResult{value}, nil
}
