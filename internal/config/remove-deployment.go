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
	"github.com/spf13/cobra"

	"github.com/onflow/flow-cli/internal/command"
	"github.com/onflow/flow-cli/pkg/flowkit"
	"github.com/onflow/flow-cli/pkg/flowkit/config"
	"github.com/onflow/flow-cli/pkg/flowkit/output"
	"github.com/onflow/flow-cli/pkg/flowkit/services"
)

type flagsRemoveDeployment struct{}

var removeDeploymentFlags = flagsRemoveDeployment{}

var RemoveDeploymentCommand = &command.Command{
	Cmd: &cobra.Command{
		Use:     "deployment <account> <network>",
		Short:   "Remove deployment from configuration",
		Example: "flow config remove deployment Foo testnet",
		Args:    cobra.MaximumNArgs(2),
	},
	Flags: &removeDeploymentFlags,
	RunS:  removeDeployment,
}

func removeDeployment(
	args []string,
	_ flowkit.ReaderWriter,
	_ command.GlobalFlags,
	_ *services.Services,
	state *flowkit.State,
) (command.Result, error) {
	if state == nil {
		return nil, config.ErrDoesNotExist
	}

	account := ""
	network := ""
	if len(args) == 2 {
		account = args[0]
		network = args[1]
	} else {
		account, network = output.RemoveDeploymentPrompt(*state.Deployments())
	}

	err := state.Deployments().Remove(account, network)
	if err != nil {
		return nil, err
	}

	err = state.SaveDefault()
	if err != nil {
		return nil, err
	}

	return &Result{
		result: "deployment removed",
	}, nil
}
