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
	"github.com/spf13/cobra"
	"io"
	"os"
	"path/filepath"

	"github.com/onflow/flowkit"
	"github.com/onflow/flowkit/output"

	"github.com/onflow/flow-cli/internal/command"
)

type flagsSetup struct {
	Scaffold   bool `default:"" flag:"scaffold" info:"Interactively select a provided scaffold for project creation"`
	ScaffoldID int  `default:"" flag:"scaffold-id" info:"Use provided scaffold ID for project creation"`
}

var setupFlags = flagsSetup{}

var SetupCommand = &command.Command{
	Cmd: &cobra.Command{
		Use:     "setup <project name>",
		Short:   "Start a new Flow project",
		Example: "flow setup my-project",
		Args:    cobra.ExactArgs(1),
		GroupID: "super",
	},
	Flags: &setupFlags,
	Run:   create,
}

func create(
	args []string,
	_ command.GlobalFlags,
	logger output.Logger,
	_ flowkit.ReaderWriter,
	_ flowkit.Services,
) (command.Result, error) {
	targetDir, err := getTargetDirectory(args[0])
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
