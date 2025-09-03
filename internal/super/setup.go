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

type flagsSetup struct {
	ConfigOnly bool `default:"false" flag:"config-only" info:"Only create a flow.json default config"`
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

	if setupFlags.ConfigOnly {
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

func validateCurrentDirectoryForInit() error {
	pwd, err := os.Getwd()
	if err != nil {
		return err
	}

	// Only check for core Flow project files that would cause real conflicts
	coreFlowPaths := []string{
		"flow.json",
		"cadence",
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

// copyDirContents copies all files and directories from src to dst
func copyDirContents(src, dst string) error {
	srcInfo, err := os.Stat(src)
	if err != nil {
		return err
	}

	if !srcInfo.IsDir() {
		return fmt.Errorf("source is not a directory")
	}

	// Read all entries in the source directory
	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}

	// Copy each entry
	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())

		if entry.IsDir() {
			// Create directory and recursively copy its contents
			err := os.MkdirAll(dstPath, 0755)
			if err != nil {
				return err
			}
			err = copyDirContents(srcPath, dstPath)
			if err != nil {
				return err
			}
		} else {
			// Copy file
			err := copyFile(srcPath, dstPath)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// copyFile copies a single file from src to dst
func copyFile(src, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	dstFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	_, err = io.Copy(dstFile, srcFile)
	if err != nil {
		return err
	}

	// Copy file permissions
	srcInfo, err := srcFile.Stat()
	if err != nil {
		return err
	}
	return os.Chmod(dst, srcInfo.Mode())
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

func startInteractiveSetup(
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

	// Ask for project name if not given
	if len(args) < 1 {
		userInput, err := prompt.RunTextInput("Enter the name of your project (leave blank to use current directory)", "Type your project name here or press Enter for current directory...")
		if err != nil {
			return "", fmt.Errorf("error running project name: %v", err)
		}

		if strings.TrimSpace(userInput) == "" {
			// Validate current directory for Flow project conflicts
			if err := validateCurrentDirectoryForInit(); err != nil {
				return "", err
			}

			// Use current directory
			pwd, err := os.Getwd()
			if err != nil {
				return "", err
			}
			targetDir = pwd
		} else {
			targetDir, err = getTargetDirectory(userInput)
			if err != nil {
				return "", err
			}
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

	// Determine README filename - use README_flow.md if README.md already exists
	readmeFileName := "README.md"
	if _, err := os.Stat(filepath.Join(targetDir, "README.md")); err == nil {
		readmeFileName = "README_flow.md"
	}

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

	// Check if we created README_flow.md instead of README.md
	readmeFile := "README.md"
	if _, err := os.Stat(filepath.Join(s.targetDir, "README_flow.md")); err == nil {
		readmeFile = "README_flow.md"
		out.WriteString("📝 Note: Created README_flow.md since README.md already exists.\n\n")
	}

	out.WriteString("Start development by following these steps:\n")

	// Only show cd command if not current directory
	if s.targetDir != wd {
		out.WriteString(fmt.Sprintf("1. '%s' to change to your new project,\n", output.Bold(fmt.Sprintf("cd %s", relDir))))
		out.WriteString(fmt.Sprintf("2. '%s' or run Flowser to start the emulator,\n", output.Bold("flow emulator")))
		out.WriteString(fmt.Sprintf("3. '%s' to test your project.\n\n", output.Bold("flow test")))
	} else {
		out.WriteString(fmt.Sprintf("1. '%s' or run Flowser to start the emulator,\n", output.Bold("flow emulator")))
		out.WriteString(fmt.Sprintf("2. '%s' to test your project.\n\n", output.Bold("flow test")))
	}

	out.WriteString(fmt.Sprintf("You should also read %s to learn more about the development process!\n", readmeFile))

	return out.String()
}

func (s *setupResult) Oneliner() string {
	return fmt.Sprintf("Project created inside %s", s.targetDir)
}

func (s *setupResult) JSON() any {
	return nil
}
