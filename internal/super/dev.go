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
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/onflow/cadence/runtime/parser"
	"github.com/onflow/flowkit/v2"
	"github.com/onflow/flowkit/v2/output"
	"github.com/spf13/cobra"

	"github.com/onflow/flow-cli/internal/command"
)

type flagsDev struct{}

var devFlags = flagsDev{}

var DevCommand = &command.Command{
	Cmd: &cobra.Command{
		Use:     "dev",
		Short:   "Build your Flow project",
		Args:    cobra.ExactArgs(0),
		Example: "flow dev",
		GroupID: "super",
	},
	Flags: &devFlags,
	RunS:  dev,
}

func dev(
	_ []string,
	_ command.GlobalFlags,
	logger output.Logger,
	flow flowkit.Services,
	state *flowkit.State,
) (command.Result, error) {
	dir, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	err = flow.Ping()
	if err != nil {
		logger.Error("Error connecting to emulator. Make sure you started an emulator using 'flow emulator' command.")
		logger.Info(fmt.Sprintf("%s This tool requires emulator to function. Emulator needs to be run inside the project root folder where the configuration file ('flow.json') exists.\n\n", output.TryEmoji()))
		return nil, nil
	}

	serviceAccount, err := state.EmulatorServiceAccount()
	if err != nil {
		return nil, err
	}

	flow.SetLogger(output.NewStdoutLogger(output.NoneLog))

	project, err := newProject(
		*serviceAccount,
		flow,
		state,
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

		var parseErr parser.Error
		if errors.As(err, &parseErr) {
			fmt.Println(err) // we just print the error but keep watching files for changes, since they might fix the issue
		} else {
			return nil, err
		}
	}

	err = project.watch()
	if err != nil {
		return nil, err
	}

	return nil, nil
}
