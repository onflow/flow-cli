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
	"os"
	"path/filepath"
	"strings"

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

type flagsInit struct {
	ConfigOnly bool `default:"false" flag:"config-only" info:"Only create a flow.json default config"`
}

var initFlags = flagsInit{}

const (
	// File permissions for created directories
	defaultDirPerm = 0755
	// Core Flow project files that indicate an existing Flow project
	flowConfigFile = "flow.json"
	// README files
	defaultReadmeFile = "README.md"
	flowReadmeFile    = "README_flow.md"
)

// TODO: Add --config-only flag
var InitCommand = &command.Command{
	Cmd: &cobra.Command{
		Use:     "init <project name>",
		Short:   "Start a new Flow project",
		Example: "flow init my-project",
		Args:    cobra.MaximumNArgs(1),
		GroupID: "super",
	},
	Flags: &initFlags,
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

	if initFlags.ConfigOnly {
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
		targetDir, err = startInteractiveInit(args, logger)
		if err != nil {
			return nil, err
		}
	}

	return &initResult{targetDir: targetDir}, nil
}

func validateCurrentDirectoryForInit() error {
	pwd, err := os.Getwd()
	if err != nil {
		return err
	}

	// Only check for core Flow project files that would cause real conflicts
	coreFlowPaths := []string{
		flowConfigFile,
		cadenceDir,
	}

	var conflicts []string
	for _, path := range coreFlowPaths {
		fullPath := filepath.Join(pwd, path)
		if _, err := os.Stat(fullPath); err == nil {
			conflicts = append(conflicts, path)
		}
	}

	if len(conflicts) > 0 {
		return fmt.Errorf("Flow project files already exist: %s. Cannot initialize Flow project in directory with existing Flow files", strings.Join(conflicts, ", "))
	}

	return nil
}

// resolveTargetDirectory determines the target directory for the Flow project
// based on user input. Empty input means current directory.
func resolveTargetDirectory(userInput string) (string, error) {
	if strings.TrimSpace(userInput) == "" {
		// Validate current directory for Flow project conflicts
		if err := validateCurrentDirectoryForInit(); err != nil {
			return "", err
		}

		// Use current directory
		pwd, err := os.Getwd()
		if err != nil {
			return "", fmt.Errorf("failed to get current working directory: %w", err)
		}
		return pwd, nil
	}

	// Use provided name to create new directory
	return getTargetDirectory(userInput)
}

func updateGitignore(targetDir string, readerWriter flowkit.ReaderWriter) error {
	return util.AddFlowEntriesToGitIgnore(targetDir, readerWriter)
}

func updateCursorIgnore(targetDir string, readerWriter flowkit.ReaderWriter) error {
	return util.AddFlowEntriesToCursorIgnore(targetDir, readerWriter)
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

	err = updateGitignore(targetDir, readerWriter)
	if err != nil {
		return err
	}

	err = updateCursorIgnore(targetDir, readerWriter)
	if err != nil {
		return err
	}

	return nil
}

func startInteractiveInit(
	args []string,
	logger output.Logger,
) (string, error) {
	var targetDir string
	var err error

	asciiArt := `   ___  ___
 /'___\/\_ \
/\ \__/\//\ \     ___   __  __  __
\ \ ,__\ \ \ \   / __` + "`" + `\/\ \/\ \/\ \
 \ \ \_/  \_\ \_/\ \L\ \ \ \_/ \_/ \
  \ \_\   /\____\ \____/\ \___x___/'
   \/_/   \/____/\/___/  \/__//__/

`
	logger.Info(asciiArt)

	rw := afero.Afero{
		Fs: afero.NewOsFs(),
	}

	// Resolve target directory from arguments or user input
	var userInput string
	if len(args) < 1 {
		userInput, err = prompt.RunTextInput("Enter the name of your project (leave blank to use current directory)", "Type your project name here or press Enter for current directory...")
		if err != nil {
			return "", fmt.Errorf("error running project name: %v", err)
		}
	} else {
		userInput = args[0]
	}

	targetDir, err = resolveTargetDirectory(userInput)
	if err != nil {
		return "", err
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

	// Determine README filename - avoid conflicts with existing README.md
	readmeFileName := getReadmeFileName(targetDir)

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
			TargetPath:   readmeFileName,
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

	err = updateGitignore(tempDir, state.ReaderWriter())
	if err != nil {
		return "", err
	}

	err = updateCursorIgnore(tempDir, state.ReaderWriter())
	if err != nil {
		return "", err
	}

	// Move or copy the temp directory contents to the target directory
	pwd, _ := os.Getwd()
	if targetDir == pwd {
		// For current directory, copy contents instead of moving the directory
		err = copyDirContents(tempDir, targetDir)
		if err != nil {
			return "", fmt.Errorf("failed to copy temp directory contents to current directory: %w", err)
		}
	} else {
		// For new directory, move the entire temp directory
		err = os.Rename(tempDir, targetDir)
		if err != nil {
			return "", fmt.Errorf("failed to move temp directory to target directory: %w", err)
		}
	}

	return targetDir, nil
}

type initResult struct {
	targetDir string
}

func (s *initResult) String() string {
	wd, _ := os.Getwd()
	relDir, _ := filepath.Rel(wd, s.targetDir)
	out := bytes.Buffer{}

	out.WriteString(fmt.Sprintf("%s Congrats! your project was created.\n\n", output.SuccessEmoji()))

	// Check if we created README_flow.md instead of README.md
	readmeFile := defaultReadmeFile
	if _, err := os.Stat(filepath.Join(s.targetDir, flowReadmeFile)); err == nil {
		readmeFile = flowReadmeFile
		out.WriteString("📝 Note: Created README_flow.md since README.md already exists.\n\n")
	}

	out.WriteString("Start development by following these steps:\n")

	// Only show cd command if not current directory
	if s.targetDir != wd {
		out.WriteString(fmt.Sprintf("1. '%s' to change to your new project,\n", output.Bold(fmt.Sprintf("cd %s", relDir))))
		out.WriteString(fmt.Sprintf("2. '%s' to start the emulator,\n", output.Bold("flow emulator")))
		out.WriteString(fmt.Sprintf("3. '%s' to test your project.\n\n", output.Bold("flow test")))
	} else {
		out.WriteString(fmt.Sprintf("1. '%s' to start the emulator,\n", output.Bold("flow emulator")))
		out.WriteString(fmt.Sprintf("2. '%s' to test your project.\n\n", output.Bold("flow test")))
	}

	out.WriteString(fmt.Sprintf("You should also read %s to learn more about the development process!\n", readmeFile))

	return out.String()
}

func (s *initResult) Oneliner() string {
	return fmt.Sprintf("Project created inside %s", s.targetDir)
}

func (s *initResult) JSON() any {
	return nil
}
