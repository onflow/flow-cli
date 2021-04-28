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

package config

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/onflow/flow-cli/internal/command"
	"github.com/onflow/flow-cli/pkg/flowcli/output"
	"github.com/onflow/flow-cli/pkg/flowcli/project"
	"github.com/onflow/flow-cli/pkg/flowcli/services"
)

type flagsRemoveAccount struct{}

var removeAccountFlags = flagsRemoveAccount{}

var RemoveAccountCommand = &command.Command{
	Cmd: &cobra.Command{
		Use:     "account <name>",
		Short:   "Remove account from configuration",
		Example: "flow config remove account Foo",
		Args:    cobra.MaximumNArgs(1),
	},
	Flags: &removeAccountFlags,
	Run: func(
		cmd *cobra.Command,
		args []string,
		globalFlags command.GlobalFlags,
		services *services.Services,
	) (command.Result, error) {
		p, err := project.Load(globalFlags.ConfigPaths)
		if err != nil {
			return nil, fmt.Errorf("configuration does not exists")
		}
		conf := p.Config()

		name := ""
		if len(args) == 1 {
			name = args[0]
		} else {
			name = output.RemoveAccountPrompt(conf.Accounts)
		}

		err = p.RemoveAccount(name)
		if err != nil {
			return nil, err
		}

		err = p.SaveDefault()
		if err != nil {
			return nil, err
		}

		return &ConfigResult{
			result: "account removed",
		}, nil
	},
}

func init() {
	RemoveAccountCommand.AddToParent(RemoveCmd)
}
