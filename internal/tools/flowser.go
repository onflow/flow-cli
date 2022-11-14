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

package tools

import (
	"fmt"
	"github.com/onflow/flow-cli/internal/command"
	"github.com/onflow/flow-cli/pkg/flowkit"
	"github.com/onflow/flow-cli/pkg/flowkit/output"
	"github.com/onflow/flow-cli/pkg/flowkit/services"
	"github.com/onflowser/flowser/pkg/flowser"
	"os"
	"runtime"

	"github.com/spf13/cobra"
)

type FlagsFlowser struct{}

var flowserFlags = FlagsWallet{}

var Flowser = &command.Command{
	Cmd: &cobra.Command{
		Use:     "flowser",
		Short:   "Starts Flowser explorer",
		Example: "flow flowser",
		Args:    cobra.ExactArgs(0),
	},
	Flags: &flowserFlags,
	RunS:  runFlowser,
}

func runFlowser(
	_ []string,
	_ flowkit.ReaderWriter,
	_ command.GlobalFlags,
	_ *services.Services,
	state *flowkit.State,
) (command.Result, error) {
	flowser := flowser.New()

	// todo here we actually have to also check if non-default path was written in configuration
	defaultPath, err := getDefaultInstallDir()
	if err != nil {
		return nil, err
	}

	if !flowser.Installed(defaultPath) {
		err := installFlowser(flowser, defaultPath)
		if err != nil {
			return nil, err
		}
	}

	projectPath, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	err = flowser.Run(defaultPath, projectPath)
	if err != nil {
		return nil, err
	}

	return nil, nil
}

func installFlowser(flowser *flowser.App, installPath string) error {
	fmt.Println("It looks like Flowser is not yet installed on your system.")
	if !output.InstallPrompt() {
		return fmt.Errorf("user denied install")
	}

	// we only allow custom paths on Windows since on MacOS apps needs to be installed inside Application folder
	if runtime.GOOS == windows {
		installPath = output.InstallPathPrompt(installPath)
	}

	logger := output.NewStdoutLogger(output.InfoLog)
	logger.StartProgress("Installing Flowser, please wait")
	defer logger.StopProgress()

	err := flowser.Install(installPath)
	if err != nil {
		return fmt.Errorf("could not install Flowser: %w", err)
	}

	return nil
}
