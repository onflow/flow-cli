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

	"github.com/onflow/flow-cli/internal/command"
	"github.com/onflow/flow-cli/pkg/flowkit"
	"github.com/onflow/flow-cli/pkg/flowkit/output"
	"github.com/onflow/flow-cli/pkg/flowkit/services"
)

type flagsRemoveContract struct{}

var removeContractFlags = flagsRemoveContract{}

var RemoveContractCommand = &command.Command{
	Cmd: &cobra.Command{
		Use:     "contract <name>",
		Short:   "Remove contract from configuration",
		Example: "flow config remove contract Foo",
		Args:    cobra.MaximumNArgs(1),
	},
	Flags: &removeContractFlags,
	RunS:  removeContract,
}

func removeContract(
	args []string,
	_ flowkit.ReaderWriter,
	globalFlags command.GlobalFlags,
	_ *services.Services,
	state *flowkit.State,
) (command.Result, error) {
	name := ""
	if len(args) == 1 {
		name = args[0]
	} else {
		name = output.RemoveContractPrompt(*state.Contracts())
	}

	err := state.Contracts().Remove(name)
	if err != nil {
		return nil, err
	}

	err = state.SaveEdited(globalFlags.ConfigPaths)
	if err != nil {
		return nil, err
	}

	return &Result{
		result: "contract removed",
	}, nil
}
