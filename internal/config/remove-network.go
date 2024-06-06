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

package config

import (
	"github.com/spf13/cobra"

	"github.com/onflow/flow-cli/internal/prompt"

	"github.com/onflow/flowkit/v2"
	"github.com/onflow/flowkit/v2/output"

	"github.com/onflow/flow-cli/internal/command"
)

type flagsRemoveNetwork struct{}

var removeNetworkFlags = flagsRemoveNetwork{}

var removeNetworkCommand = &command.Command{
	Cmd: &cobra.Command{
		Use:     "network <name>",
		Short:   "Remove network from configuration",
		Example: "flow config remove network Foo",
		Args:    cobra.MaximumNArgs(1),
	},
	Flags: &removeNetworkFlags,
	RunS:  removeNetwork,
}

func removeNetwork(
	args []string,
	globalFlags command.GlobalFlags,
	_ output.Logger,
	_ flowkit.Services,
	state *flowkit.State,
) (command.Result, error) {
	name := ""
	if len(args) == 1 {
		name = args[0]
	} else {
		name = prompt.RemoveNetworkPrompt(*state.Networks())
	}

	err := state.Networks().Remove(name)
	if err != nil {
		return nil, err
	}

	err = state.SaveEdited(globalFlags.ConfigPaths)
	if err != nil {
		return nil, err
	}

	return &result{
		result: "network removed",
	}, nil
}
