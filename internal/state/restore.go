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

package state

import (
	"github.com/onflow/flow-cli/internal/command"
	"github.com/onflow/flow-cli/pkg/flowcli/output"
	"github.com/onflow/flow-cli/pkg/flowcli/services"
	"github.com/spf13/cobra"
)

type flagsRestore struct {
	Name string `flag:"name" info:"Name of the snapshot"`
}

var restoreFlags = flagsRestore{}

var RestoreCommand = &command.Command{
	Cmd: &cobra.Command{
		Use:     "restore",
		Short:   "Restore to named snapshot",
		Example: "flow state restore",
		Args:    cobra.NoArgs,
	},
	Flags: &restoreFlags,
	Run: func(
		cmd *cobra.Command,
		args []string,
		globalFlags command.GlobalFlags,
		services *services.Services,
	) (command.Result, error) {
		name := restoreFlags.Name

		if name == "" {
			names, err := services.State.List()
			if err != nil {
				return nil, err
			}

			name = output.SnapshotPrompt(names)
		}

		err := services.State.Restore(name)
		if err != nil {
			return nil, err
		}

		return &StateResult{
			result: "state restored",
		}, nil
	},
}

func init() {
	RestoreCommand.AddToParent(Cmd)
}
