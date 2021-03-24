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

	"github.com/onflow/flow-cli/internal/command"
	"github.com/onflow/flow-cli/pkg/flowcli/services"
	"github.com/spf13/cobra"
)

type flagsScripts struct {
	ArgsJSON string   `default:"" flag:"args-json" info:"arguments in JSON-Cadence format"`
	Args     []string `default:"" flag:"arg" info:"argument in Type:Value format"`
	Code     bool     `default:"false" flag:"code" info:"⚠️  DEPRECATED: use filename argument"`
}

var scriptFlags = flagsScripts{}

var ExecuteCommand = &command.Command{
	Cmd: &cobra.Command{
		Use:     "execute <filename>",
		Short:   "Execute a script",
		Example: `flow scripts execute script.cdc --arg String:"Hello" --arg String:"World"`,
		Args:    cobra.ExactArgs(1),
	},
	Flags: &scriptFlags,
	Run: func(
		cmd *cobra.Command,
		args []string,
		services *services.Services,
	) (command.Result, error) {
		if scriptFlags.Code {
			return nil, fmt.Errorf("⚠️  DEPRECATED: use filename argument")
		}

		value, err := services.Scripts.Execute(
			args[0], // filename
			scriptFlags.Args,
			scriptFlags.ArgsJSON,
		)
		if err != nil {
			return nil, err
		}

		return &ScriptResult{value}, nil
	},
}
