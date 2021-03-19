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
	"github.com/onflow/flow-cli/internal/command"
	"github.com/onflow/flow-cli/pkg/flow"
	"github.com/onflow/flow-cli/pkg/flow/services"
	"github.com/psiemens/sconfig"
	"github.com/spf13/cobra"
)

type flagsScripts struct {
	ArgsJSON string   `default:"" flag:"argsJSON" info:"arguments in JSON-Cadence format"`
	Args     []string `default:"" flag:"arg" info:"argument in Type:Value format"`
}

type cmdExecuteScript struct {
	cmd   *cobra.Command
	flags flagsScripts
}

// NewExecuteScriptCmd creates new script command
func NewExecuteScriptCmd() command.Command {
	return &cmdExecuteScript{
		cmd: &cobra.Command{
			Use:     "execute <filename>",
			Short:   "Execute a script",
			Example: `flow scripts execute script.cdc --arg String:"Hello" --arg String:"World"`,
			Args:    cobra.ExactArgs(1),
		},
	}
}

// Run script command
func (s *cmdExecuteScript) Run(
	cmd *cobra.Command,
	args []string,
	project *flow.Project,
	services *services.Services,
) (command.Result, error) {
	value, err := services.Scripts.Execute(args[0], s.flags.Args, s.flags.ArgsJSON)
	if err != nil {
		return nil, err
	}

	return &ScriptResult{value}, nil
}

// GetFlags for script
func (s *cmdExecuteScript) GetFlags() *sconfig.Config {
	return sconfig.New(&s.flags)
}

// GetCmd get command
func (s *cmdExecuteScript) GetCmd() *cobra.Command {
	return s.cmd
}
