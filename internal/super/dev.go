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
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/onflow/flow-cli/internal/command"
	"github.com/onflow/flow-cli/pkg/flowkit"
	"github.com/onflow/flow-cli/pkg/flowkit/output"
	"github.com/onflow/flow-cli/pkg/flowkit/services"
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

	_, err = services.Status.Ping(network)
	if err != nil {
		fmt.Printf("%s Error connecting to emulator. Make sure you started an emulator using 'flow emulator' command.\n", output.ErrorEmoji())
		fmt.Printf("%s This tool requires emulator to function. Emulator needs to be run inside the project root folder where the configuration file ('flow.json') exists.\n\n", output.TryEmoji())
		return nil, nil
	}

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
		fmt.Printf("%s Failed to run the command, please make sure you ran 'flow setup' command first and that you are running this command inside the project ROOT folder.\n\n", output.TryEmoji())
		return nil, err
	}

	err = project.startup()
	if err != nil {
		if strings.Contains(err.Error(), "does not have a valid signature") {
			fmt.Printf("%s Failed to run the command, please make sure you started the emulator inside the project ROOT folder by running 'flow emulator'.\n\n", output.TryEmoji())
			return nil, nil
		}

		fmt.Println(err) // we just print the error but keep watching files for changes, since they might fix the error
	}

	err = project.watch()
	if err != nil {
		return nil, err
	}

	return nil, nil
}
