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

package super

import (
	"github.com/onflow/flow-cli/internal/scripts"
	"github.com/onflow/flow-cli/internal/transactions"
	"os"

	"github.com/onflow/flow-cli/internal/command"
	"github.com/onflow/flow-cli/pkg/flowkit"
	"github.com/onflow/flow-cli/pkg/flowkit/output"
	"github.com/onflow/flow-cli/pkg/flowkit/services"

	"github.com/spf13/cobra"
)

type flagsExec struct{}

var execFlags = flagsExec{}

var ExecCommand = &command.Command{
	Cmd: &cobra.Command{
		Use:     "exec <name>",
		Short:   "Send a transaction or execute a script",
		Args:    cobra.ExactArgs(1),
		Example: "flow exec getResult",
	},
	Flags: &execFlags,
	RunS:  exec,
}

func exec(
	args []string,
	readerWriter flowkit.ReaderWriter,
	globalFlags command.GlobalFlags,
	services *services.Services,
	state *flowkit.State,
) (command.Result, error) {
	dir, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	name := args[0]

	service, err := state.EmulatorServiceAccount()
	if err != nil {
		return nil, err
	}

	services.SetLogger(output.NewStdoutLogger(output.NoneLog))

	project, err := newProject(
		*service,
		services,
		state,
		readerWriter,
		newProjectFiles(dir),
	)
	if err != nil {
		return nil, err
	}

	scriptRes, tx, txRes, err := project.exec(name)
	if err != nil {
		return nil, err
	}

	if scriptRes != nil {
		return &scripts.ScriptResult{Value: scriptRes}, nil
	}

	return &transactions.TransactionResult{
		Result: txRes,
		Tx:     tx,
	}, nil
}
