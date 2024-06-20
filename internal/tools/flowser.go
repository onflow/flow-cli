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
	"os"
	"runtime"

	"github.com/onflow/flow-cli/internal/prompt"

	"github.com/onflowser/flowser/v3/pkg/flowser"
	"github.com/spf13/cobra"

	"github.com/onflow/flowkit/v2"
	"github.com/onflow/flowkit/v2/config"
	"github.com/onflow/flowkit/v2/output"

	"github.com/onflow/flow-cli/internal/command"
	"github.com/onflow/flow-cli/internal/settings"
)

type flagsFlowser struct{}

var flowserFlags = flagsWallet{}

var Flowser = &command.Command{
	Cmd: &cobra.Command{
		Use:     "flowser",
		Short:   "Run Flowser project explorer",
		Example: "flow flowser",
		Args:    cobra.ExactArgs(0),
		GroupID: "tools",
	},
	Flags: &flowserFlags,
	Run:   runFlowser,
}

func runFlowser(
	_ []string,
	_ command.GlobalFlags,
	_ output.Logger,
	reader flowkit.ReaderWriter,
	_ flowkit.Services,
) (command.Result, error) {
	flowser := flowser.New()

	installPath, err := settings.GetFlowserPath()
	if err != nil {
		return nil, fmt.Errorf("failure reading setting: %w", err)
	}

	if !flowser.Installed(installPath) {
		installPath, err = installFlowser(flowser, installPath)
		if err != nil {
			return nil, err
		}
	}

	projectPath, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	// check if current directory is existing flow project if not then don't pass project path to Flowser, so user can choose a project
	_, err = reader.ReadFile(config.DefaultPath)
	if os.IsNotExist(err) {
		projectPath = ""
	}

	fmt.Printf("%s Starting up Flowser, please wait...\n", output.SuccessEmoji())
	err = flowser.Run(installPath, projectPath)
	if err != nil {
		return nil, err
	}

	return nil, nil
}

func installFlowser(flowser *flowser.App, installPath string) (string, error) {
	fmt.Println("It looks like Flowser is not yet installed on your system.")
	installChoice := prompt.InstallPrompt()
	if installChoice == prompt.CancelInstall {
		return "", fmt.Errorf("user denied install")
	}

	// if user says it already installed it we only ask for path and return it
	if installChoice == prompt.AlreadyInstalled {
		installPath = prompt.InstallPathPrompt(installPath)
		_ = settings.SetFlowserPath(installPath)
		return installPath, nil
	}

	// MacOS apps must always be installed inside Application folder
	if runtime.GOOS != settings.Darwin {
		installPath = prompt.InstallPathPrompt(installPath)
		_ = settings.SetFlowserPath(installPath)
	}

	logger := output.NewStdoutLogger(output.InfoLog)
	logger.StartProgress(fmt.Sprintf("%s Installing Flowser, this may take few minutes, please wait ", output.TryEmoji()))
	defer logger.StopProgress()

	// create all folders if they don't exist, does nothing if they exist
	err := os.MkdirAll(installPath, os.ModePerm)
	if err != nil {
		return "", err
	}

	err = flowser.Install(installPath)
	if err != nil {
		return "", fmt.Errorf("could not install Flowser: %w", err)
	}

	return installPath, nil
}
