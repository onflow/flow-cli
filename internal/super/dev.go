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
	"github.com/onflow/flow-cli/internal/command"
	"github.com/onflow/flow-cli/pkg/flowkit"
	"github.com/onflow/flow-cli/pkg/flowkit/output"
	"github.com/onflow/flow-cli/pkg/flowkit/services"
	"os"

	"github.com/spf13/cobra"
)

type flagsDev struct{}

var devFlags = flagsDev{}

var DevCommand = &command.Command{
	Cmd: &cobra.Command{
		Use:     "dev",
		Short:   "Monitor your project during development", // todo improve
		Args:    cobra.ExactArgs(0),
		Example: "flow dev",
	},
	Flags: &devFlags,
	RunS:  dev,
}

func dev(
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

	// todo dev work if not run on top root directory - at least have a warning
	// todo handle emulator running as part of this service or part of existing running emulator
	// todo possible bug, investigate when new account is created whether we deploy too soon before it was even founded. This is error: ommand Error: failure to startup: execution error code 1103: [Error Code: 1103] The account with address (120e725050340cab) uses 783 bytes of storage which is over its capacity (0 bytes). Capacity can be increased by adding FLOW tokens to the account.

	service, err := state.EmulatorServiceAccount()
	if err != nil {
		return nil, err
	}

	services.SetLogger(output.NewStdoutLogger(output.NoneLog))

	project, err := newProject(
		service,
		services,
		state,
		readerWriter,
		newProjectFiles(dir),
	)
	if err != nil {
		return nil, err
	}

	err = project.watch()
	if err != nil {
		return nil, err
	}

	return nil, nil
}
