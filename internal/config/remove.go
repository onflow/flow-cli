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

	"github.com/onflow/flow-cli/internal/command"
	"github.com/onflow/flow-cli/pkg/flowcli/output"
	"github.com/onflow/flow-cli/pkg/flowcli/project"
	"github.com/onflow/flow-cli/pkg/flowcli/services"
	"github.com/spf13/cobra"
)

type flagsRemove struct {
	Name string `flag:"name" info:"Resource name"`
}

var removeFlags = flagsRemove{}

var RemoveCommand = &command.Command{
	Cmd: &cobra.Command{
		Use:       "remove <account|contract|deployment|network>",
		Short:     "Remove resource from configuration",
		Example:   "flow config remove account",
		ValidArgs: []string{"account", "contract", "deployment", "network"},
		Args:      cobra.ExactArgs(1),
	},
	Flags: &removeFlags,
	Run: func(
		cmd *cobra.Command,
		args []string,
		globalFlags command.GlobalFlags,
		services *services.Services,
	) (command.Result, error) {
		resource := args[0]

		p, err := project.Load(globalFlags.ConfigPath)
		if err != nil {
			return nil, fmt.Errorf("configuration does not exists")
		}
		conf := p.Config()

		switch resource {
		case "account":
			name := removeFlags.Name
			if name == "" {
				name = output.RemoveAccountPrompt(conf.Accounts)
			}

			err := services.Config.RemoveAccount(name)
			if err != nil {
				return nil, err
			}

			return &ConfigResult{
				result: "account removed",
			}, nil

		case "deployment":
			accountName, networkName := output.RemoveDeploymentPrompt(conf.Deployments)
			err := services.Config.RemoveDeployment(accountName, networkName)
			if err != nil {
				return nil, err
			}

			return &ConfigResult{
				result: "deployment removed",
			}, nil

		case "contract":
			name := removeFlags.Name
			if name == "" {
				name = output.RemoveContractPrompt(conf.Contracts)
			}

			err := services.Config.RemoveContract(name)
			if err != nil {
				return nil, err
			}

			return &ConfigResult{
				result: "contract removed",
			}, nil

		case "network":
			name := removeFlags.Name
			if name == "" {
				name = output.RemoveNetworkPrompt(conf.Networks)
			}

			err := services.Config.RemoveNetwork(name)
			if err != nil {
				return nil, err
			}

			return &ConfigResult{
				result: "network removed",
			}, nil
		}

		return nil, nil
	},
}
