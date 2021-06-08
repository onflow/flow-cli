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

package scripts

import (
	"fmt"

	"github.com/onflow/flow-cli/pkg/flowcli"

	"github.com/onflow/flow-cli/pkg/flowcli/project"
	"github.com/onflow/flow-cli/pkg/flowcli/util"

	"github.com/spf13/cobra"

	"github.com/onflow/flow-cli/internal/command"
	"github.com/onflow/flow-cli/pkg/flowcli/services"
)

type flagsScripts struct {
	ArgsJSON string   `default:"" flag:"args-json" info:"arguments in JSON-Cadence format"`
	Arg      []string `default:"" flag:"arg" info:"argument in Type:Value format"`
	Code     string   `default:"" flag:"code" info:"⚠️  Deprecated: use filename argument"`
	Args     string   `default:"" flag:"args" info:"⚠️  Deprecated: use arg or args-json flag"`
}

var scriptFlags = flagsScripts{}

var ExecuteCommand = &command.Command{
	Cmd: &cobra.Command{
		Use:     "execute <filename>",
		Short:   "Execute a script",
		Example: `flow scripts execute script.cdc --arg String:"Meow" --arg String:"Woof"`,
		Args:    cobra.MaximumNArgs(1),
	},
	Flags: &scriptFlags,
	Run: func(
		cmd *cobra.Command,
		args []string,
		globalFlags command.GlobalFlags,
		services *services.Services,
		project *project.Project,
	) (command.Result, error) {
		filename := ""
		if len(args) == 1 {
			filename = args[0]
		} else if scriptFlags.Code != "" {
			fmt.Println("⚠️  DEPRECATION WARNING: use filename as a command argument <filename>")
			filename = scriptFlags.Code
		} else {
			return nil, fmt.Errorf("provide a valide filename command argument")
		}

		if scriptFlags.Args != "" {
			fmt.Println("⚠️  DEPRECATION WARNING: use arg flag in Type:Value format or args-json for JSON format")

			if len(scriptFlags.Arg) == 0 && scriptFlags.ArgsJSON == "" {
				scriptFlags.ArgsJSON = scriptFlags.Args // backward compatible, args was in json format
			}
		}

		code, err := util.LoadFile(filename)
		if err != nil {
			return nil, fmt.Errorf("error loading script file: %w", err)
		}

		scriptArgs, err := flowcli.ParseArguments(scriptFlags.Arg, scriptFlags.ArgsJSON)
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
	},
}
