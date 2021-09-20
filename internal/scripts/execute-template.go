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

	"github.com/onflow/flow-cli/pkg/flowkit/templates"

	"github.com/onflow/cadence"
	"github.com/spf13/cobra"

	"github.com/onflow/flow-cli/internal/command"
	"github.com/onflow/flow-cli/pkg/flowkit"
	"github.com/onflow/flow-cli/pkg/flowkit/services"
)

type flagsTemplate struct {
	ArgsJSON string `default:"" flag:"args-json" info:"arguments in JSON-Cadence format"`
}

var templateFlags = flagsTemplate{}

var ExecuteTemplateCommand = &command.Command{
	Cmd: &cobra.Command{
		Use:     "execute-template <template name> [<argument> <argument> ...]",
		Short:   "Execute a script template",
		Example: `flow scripts execute-template fusd-balance 0x01cf0e2f2f715450`,
		Args:    cobra.MinimumNArgs(1),
	},
	Flags: &templateFlags,
	Run:   executeTemplate,
}

func executeTemplate(
	args []string,
	_ flowkit.ReaderWriter,
	globalFlags command.GlobalFlags,
	services *services.Services,
) (command.Result, error) {
	templateName := args[0]
	template, err := templates.ScriptByName(templateName)
	if err != nil {
		return nil, err
	}

	source, err := template.Source(globalFlags.Network)
	if err != nil {
		return nil, err
	}

	var scriptArgs []cadence.Value
	if templateFlags.ArgsJSON != "" {
		scriptArgs, err = flowkit.ParseArgumentsJSON(templateFlags.ArgsJSON)
	} else {
		scriptArgs, err = flowkit.ParseArgumentsWithoutType("", source, args[1:])
	}

	if err != nil {
		return nil, fmt.Errorf("error parsing script arguments: %w", err)
	}

	value, err := services.Scripts.Execute(
		source,
		scriptArgs,
		"",
		globalFlags.Network,
	)
	if err != nil {
		return nil, err
	}

	return &ScriptResult{value}, nil
}
