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
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/onflow/flow-cli/internal/config"
	"github.com/onflow/flow-cli/internal/util"

	"github.com/onflow/flowkit"
	"github.com/onflow/flowkit/output"

	"github.com/onflow/flow-cli/internal/command"
)

type flagsSetup struct {
	ConfigOnly bool `default:"false" flag:"config-only" info:"Only create a flow.json default config"`
	Scaffold   bool `default:"" flag:"scaffold" info:"Interactively select a provided scaffold for project creation"`
	ScaffoldID int  `default:"" flag:"scaffold-id" info:"Use provided scaffold ID for project creation"`
}

var setupFlags = flagsSetup{}

var SetupCommand = &command.Command{
	Cmd: &cobra.Command{
		Use:     "setup <project name>",
		Short:   "Start a new Flow project",
		Example: "flow setup my-project",
		Args:    cobra.MaximumNArgs(1),
		GroupID: "super",
	},
	Flags: &setupFlags,
	RunS:  create,
}

func create(
	args []string,
	_ command.GlobalFlags,
	logger output.Logger,
	_ flowkit.Services,
	state *flowkit.State,
) (command.Result, error) {
	var targetDir string
	var err error

	if setupFlags.Scaffold || setupFlags.ScaffoldID != 0 {
		targetDir, err = getTargetDirectory(args[0])
		if err != nil {
			return nil, err
		}

		selectedScaffold, err := selectScaffold(logger)
		if err != nil {
			return nil, fmt.Errorf("error selecting scaffold %w", err)
		}

		logger.StartProgress(fmt.Sprintf("Creating your project %s", targetDir))
		defer logger.StopProgress()

		if selectedScaffold != nil {
			err = cloneScaffold(targetDir, *selectedScaffold)
			if err != nil {
				return nil, fmt.Errorf("failed creating scaffold %w", err)
			}
		}
	} else {
		// Ask for project name if not given
		if len(args) < 1 {
			name := util.NamePrompt()
			targetDir, err = getTargetDirectory(name)
			if err != nil {
				return nil, err
			}
		} else {
			targetDir, err = getTargetDirectory(args[0])
			if err != nil {
				return nil, err
			}
		}

		params := config.InitConfigParameters{
			ServiceKeySigAlgo:  "ECDSA_P256",
			ServiceKeyHashAlgo: "SHA3_256",
			Reset:              false,
			Global:             false,
			TargetDirectory:    targetDir,
		}
		_, err := config.InitializeConfiguration(params, logger, state.ReaderWriter())
		if err != nil {
			return nil, fmt.Errorf("failed to initialize configuration: %w", err)
		}

		// Generate standard cadence files
		// cadence/contracts/DefaultContract.cdc
		// cadence/scripts/DefaultScript.cdc
		// cadence/transactions/DefaultTransaction.cdc
		// cadence/tests/DefaultContract_test.cdc

		directoryPath := filepath.Join(targetDir, "cadence")

		_, err = generateNew([]string{"DefaultContract"}, "contract", directoryPath, logger, state)
		if err != nil {
			return nil, err
		}

		_, err = generateNew([]string{"DefaultScript"}, "script", directoryPath, logger, state)
		if err != nil {
			return nil, err
		}

		_, err = generateNew([]string{"DefaultTransaction"}, "transaction", directoryPath, logger, state)
		if err != nil {
			return nil, err
		}

	}

	return &setupResult{targetDir: targetDir}, nil
}

func getTargetDirectory(directory string) (string, error) {
	pwd, err := os.Getwd()
	if err != nil {
		return "", err
	}

	target := filepath.Join(pwd, directory)
	info, err := os.Stat(target)
	if !os.IsNotExist(err) {
		if !info.IsDir() {
			return "", fmt.Errorf("%s is a file", target)
		}

		file, err := os.Open(target)
		if err != nil {
			return "", err
		}
		defer file.Close()

		_, err = file.Readdirnames(1)
		if err != io.EOF {
			return "", fmt.Errorf("directory is not empty: %s", target)
		}
	}
	return target, nil
}

type setupResult struct {
	targetDir string
}

func (s *setupResult) String() string {
	wd, _ := os.Getwd()
	relDir, _ := filepath.Rel(wd, s.targetDir)
	out := bytes.Buffer{}

	out.WriteString(fmt.Sprintf("%s Congrats! your project was created.\n\n", output.SuccessEmoji()))
	out.WriteString("Start development by following these steps:\n")
	out.WriteString(fmt.Sprintf("1. '%s' to change to your new project,\n", output.Bold(fmt.Sprintf("cd %s", relDir))))
	out.WriteString(fmt.Sprintf("2. '%s' or run Flowser to start the emulator,\n", output.Bold("flow emulator")))
	out.WriteString(fmt.Sprintf("3. '%s' to start developing.\n\n", output.Bold("flow dev")))
	out.WriteString(fmt.Sprintf("You should also read README.md to learn more about the development process!\n"))

	return out.String()
}

func (s *setupResult) Oneliner() string {
	return fmt.Sprintf("Project created inside %s", s.targetDir)
}

func (s *setupResult) JSON() any {
	return nil
}
