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

	"github.com/spf13/cobra"

	"github.com/onflow/flow-cli/internal/command"
	"github.com/onflow/flow-cli/pkg/flowcli/services"
)

type flagsScripts struct {
	ArgsJSON string   `default:"" flag:"args-json" info:"arguments in JSON-Cadence format"`
	Arg      []string `default:"" flag:"arg" info:"argument in Type:Value format"`
	Code     bool     `default:"false" flag:"code" info:"⚠️  No longer supported: use filename argument"`
	Args     string   `default:"false" flag:"args" info:"⚠️  No longer supported: use arg or args-json flag"`
}

var scriptFlags = flagsScripts{}

var ExecuteCommand = &command.Command{
	Cmd: &cobra.Command{
		Use:     "execute <filename>",
		Short:   "Execute a script",
		Example: `flow scripts execute script.cdc --arg String:"Meow" --arg String:"Woof"`,
		Args:    cobra.ExactArgs(1),
	},
	Flags: &scriptFlags,
	Run: func(
		cmd *cobra.Command,
		args []string,
		globalFlags command.GlobalFlags,
		services *services.Services,
	) (command.Result, error) {
		if scriptFlags.Code {
			return nil, fmt.Errorf("⚠️  No longer supported: use filename argument")
		}
		if scriptFlags.Args != "" {
			return nil, fmt.Errorf("⚠️  No longer supported: use arg flag in Type:Value format or arg-json for JSON format")
		}

		value, err := services.Scripts.Execute(
			args[0], // filename
			scriptFlags.Arg,
			scriptFlags.ArgsJSON,
		)
		if err != nil {
			return nil, err
		}

		return &ScriptResult{value}, nil
	},
}
