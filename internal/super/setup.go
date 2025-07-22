/*
 * Flow CLI
 *
 * Copyright Flow Foundation
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

	"github.com/onflow/flowkit/v2"

	"github.com/spf13/afero"

	"github.com/onflow/flow-cli/internal/prompt"

	"github.com/spf13/cobra"

	"github.com/onflow/flowkit/v2/output"

	"github.com/onflow/flow-cli/internal/command"
	"github.com/onflow/flow-cli/internal/config"
	"github.com/onflow/flow-cli/internal/dependencymanager"
	"github.com/onflow/flow-cli/internal/super/generator"
	"github.com/onflow/flow-cli/internal/util"
)

type flagsSetup struct {
	ConfigOnly bool `default:"false" flag:"config-only" info:"Only create a flow.json default config"`
	Scaffold   bool `default:"" flag:"scaffold" info:"Interactively select a provided scaffold for project creation"`
	ScaffoldID int  `default:"" flag:"scaffold-id" info:"Use provided scaffold ID for project creation"`
}

var setupFlags = flagsSetup{}

// TODO: Add --config-only flag
var SetupCommand = &command.Command{
	Cmd: &cobra.Command{
		Use:     "init <project name>",
		Short:   "Start a new Flow project",
		Example: "flow init my-project",
		Args:    cobra.MaximumNArgs(1),
		GroupID: "super",
	},
	Flags: &setupFlags,
	Run:   create,
}

func create(
	args []string,
	_ command.GlobalFlags,
	logger output.Logger,
	readerWriter flowkit.ReaderWriter,
	_ flowkit.Services,
) (command.Result, error) {
	var targetDir string
	var err error

	if setupFlags.Scaffold || setupFlags.ScaffoldID != 0 {
		fmt.Println("`scaffold` and `scaffold-id` are deprecated, and will be removed in a future release.")
		// Error if no project name is given
		if len(args) < 1 || args[0] == "" {
			return nil, fmt.Errorf("no project name provided")
		}

		targetDir, err = handleScaffold(args[0], logger)
		if err != nil {
			return nil, err
		}
	} else if setupFlags.ConfigOnly {
		if len(args) > 0 {
			return nil, fmt.Errorf("project name not required when using --config-only flag")
		}

		err = createConfigOnly("", readerWriter)
		if err != nil {
			return nil, err
		}

		logger.Info(util.MessageWithEmojiPrefix("🎉", "Configuration created successfully!"))

		return nil, nil
	} else {
		targetDir, err = startInteractiveSetup(args, logger)
		if err != nil {
			return nil, err
		}
	}

	return &setupResult{targetDir: targetDir}, nil
}

func updateGitignore(targetDir string) error {
	rw := afero.Afero{
		Fs: afero.NewOsFs(),
	}
	return util.AddFlowEntriesToGitIgnore(targetDir, rw)
}

func updateCursorIgnore(targetDir string) error {
	rw := afero.Afero{
		Fs: afero.NewOsFs(),
	}
	return util.AddFlowEntriesToCursorIgnore(targetDir, rw)
}

func createConfigOnly(targetDir string, readerWriter flowkit.ReaderWriter) error {
	params := config.InitConfigParameters{
		ServiceKeySigAlgo:  "ECDSA_P256",
		ServiceKeyHashAlgo: "SHA3_256",
		Reset:              false,
		Global:             false,
		TargetDirectory:    targetDir,
	}
	state, err := config.InitializeConfiguration(params, readerWriter)
	if err != nil {
		return err
	}

	err = state.SaveDefault()
	if err != nil {
		return err
	}

	err = updateGitignore(targetDir)
	if err != nil {
		return err
	}

	err = updateCursorIgnore(targetDir)
	if err != nil {
		return err
	}

	return nil
}

func startInteractiveSetup(
	args []string,
	logger output.Logger,
) (string, error) {
	var targetDir string
	var err error

	rw := afero.Afero{
		Fs: afero.NewOsFs(),
	}

	// Ask for project name if not given
	if len(args) < 1 {
		userInput, err := prompt.RunTextInput("Enter the name of your project", "Type your project name here...")
		if err != nil {
			return "", fmt.Errorf("error running project name: %v", err)
		}

		targetDir, err = getTargetDirectory(userInput)
		if err != nil {
			return "", err
		}
	} else {
		targetDir, err = getTargetDirectory(args[0])
		if err != nil {
			return "", err
		}
	}

	// Create a temp directory which will later be moved to the target directory if successful
	tempDir, err := os.MkdirTemp("", "flow-cli-*")
	if err != nil {
		return "", fmt.Errorf("failed to create temp directory: %w", err)
	}

	defer func() {
		if err := os.RemoveAll(tempDir); err != nil {
			logger.Error(fmt.Sprintf("Failed to remove %s: %v", tempDir, err))
		}
	}()

	params := config.InitConfigParameters{
		ServiceKeySigAlgo:  "ECDSA_P256",
		ServiceKeyHashAlgo: "SHA3_256",
		Reset:              false,
		Global:             false,
		TargetDirectory:    tempDir,
	}
	state, err := config.InitializeConfiguration(params, rw)
	if err != nil {
		return "", fmt.Errorf("failed to initialize configuration: %w", err)
	}

	msg := "Would you like to install any core contracts and their dependencies?"
	if prompt.GenericBoolPrompt(msg) {
		err := dependencymanager.PromptInstallCoreContracts(logger, state, tempDir, nil)
		if err != nil {
			return "", err
		}
	}

	// Generate standard cadence files & README.md
	// cadence/contracts/DefaultContract.cdc
	// cadence/scripts/DefaultScript.cdc
	// cadence/transactions/DefaultTransaction.cdc
	// cadence/tests/DefaultContract_test.cdc
	// README.md

	templates := []generator.TemplateItem{
		generator.ContractTemplate{
			Name:         "Counter",
			TemplatePath: "contract_counter.cdc.tmpl",
		},
		generator.ScriptTemplate{
			Name:         "GetCounter",
			TemplatePath: "script_counter.cdc.tmpl",
			Data:         map[string]interface{}{"ContractName": "Counter"},
		},
		generator.TransactionTemplate{
			Name:         "IncrementCounter",
			TemplatePath: "transaction_counter.cdc.tmpl",
			Data:         map[string]interface{}{"ContractName": "Counter"},
		},
		generator.FileTemplate{
			TemplatePath: "README.md.tmpl",
			TargetPath:   "README.md",
			Data: map[string]interface{}{
				"Dependencies": (func() []map[string]interface{} {
					contracts := []map[string]interface{}{}
					for _, dep := range *state.Dependencies() {
						contracts = append(contracts, map[string]interface{}{
							"Name": dep.Name,
						})
					}
					return contracts
				})(),
				"Contracts": []map[string]interface{}{
					{"Name": "Counter"},
				},
				"Scripts": []map[string]interface{}{
					{"Name": "GetCounter"},
				},
				"Transactions": []map[string]interface{}{
					{"Name": "IncrementCounter"},
				},
				"Tests": []map[string]interface{}{
					{"Name": "Counter_test"},
				},
			},
		},
	}

	g := generator.NewGenerator(tempDir, state, logger, true, false)
	err = g.Create(templates...)
	if err != nil {
		return "", err
	}

	err = state.Save(filepath.Join(tempDir, "flow.json"))
	if err != nil {
		return "", err
	}

	err = updateGitignore(tempDir)
	if err != nil {
		return "", err
	}

	err = updateCursorIgnore(tempDir)
	if err != nil {
		return "", err
	}

	// Move the temp directory to the target directory
	err = os.Rename(tempDir, targetDir)
	if err != nil {
		return "", fmt.Errorf("failed to move temp directory to target directory: %w", err)
	}

	return targetDir, nil
}

// getTargetDirectory checks if the specified directory path is suitable for use.
// It verifies that the path points to an existing, empty directory.
// If the directory does not exist, the function returns the path without error,
// indicating that the path is available for use (assuming creation is handled elsewhere).
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
	out.WriteString(fmt.Sprintf("3. '%s' to test your project.\n\n", output.Bold("flow test")))
	out.WriteString(fmt.Sprintf("You should also read README.md to learn more about the development process!\n"))

	return out.String()
}

func (s *setupResult) Oneliner() string {
	return fmt.Sprintf("Project created inside %s", s.targetDir)
}

func (s *setupResult) JSON() any {
	return nil
}
