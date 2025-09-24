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

	"github.com/onflow/flow-cli/common/branding"
	"github.com/onflow/flow-cli/internal/prompt"

	"github.com/spf13/cobra"

	"github.com/onflow/flowkit/v2/output"

	"github.com/onflow/flow-cli/internal/command"
	"github.com/onflow/flow-cli/internal/config"
	"github.com/onflow/flow-cli/internal/dependencymanager"
	"github.com/onflow/flow-cli/internal/super/generator"
	"github.com/onflow/flow-cli/internal/util"
	flowsdk "github.com/onflow/flow-go-sdk"
	flowkitConfig "github.com/onflow/flowkit/v2/config"
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
// based on user input by resolving a candidate path and checking if it equals
// the current working directory.
func resolveTargetDirectory(userInput string) (string, error) {
	pwd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("failed to get current working directory: %w", err)
	}

	trimmed := strings.TrimSpace(userInput)

	// Build a candidate absolute path for comparison
	var candidate string
	if trimmed == "" {
		candidate = pwd
	} else if filepath.IsAbs(trimmed) {
		candidate = filepath.Clean(trimmed)
	} else {
		candidate = filepath.Clean(filepath.Join(pwd, trimmed))
	}

	// If candidate resolves to current directory, validate and use it
	if candidate == filepath.Clean(pwd) {
		if err := validateCurrentDirectoryForInit(); err != nil {
			return "", err
		}
		return pwd, nil
	}

	// Otherwise, use provided name/path to create or validate new directory
	return getTargetDirectory(trimmed)
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

	logger.Info(branding.GreenStyle.Render(branding.FlowASCII) + "\n")

	rw := afero.Afero{
		Fs: afero.NewOsFs(),
	}

	// Resolve target directory from arguments or user input
	var userInput string
	if len(args) < 1 {
		userInput, err = prompt.RunTextInput("Enter the name of your project (leave blank to use current directory)", "Project name or Enter for current directory")
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

	projectTypes := []ProjectType{
		ProjectTypeDefault,
		ProjectTypeScheduledTransactions,
		ProjectTypeCustom,
	}
	projectOptions := make([]string, len(projectTypes))
	descriptionToType := make(map[string]ProjectType)
	for i, pt := range projectTypes {
		description := getProjectTypeDescription(pt)
		projectOptions[i] = description
		descriptionToType[description] = pt
	}

	msg := "What type of Flow project would you like to create?"
	selectedProject, err := prompt.RunSingleSelect(projectOptions, msg)
	if err != nil {
		return "", err
	}

	projectType := descriptionToType[selectedProject]
	switch projectType {
	case ProjectTypeCustom:
		err := dependencymanager.PromptInstallCoreContracts(logger, state, tempDir, nil, dependencymanager.DependencyFlags{})
		if err != nil {
			return "", err
		}
		projectType = ProjectTypeDefault
	case ProjectTypeScheduledTransactions:
		// TODO: Add FlowTransactionScheduler as core contract once it's available
		// coreContracts := []string{"FlowTransactionScheduler"}
		coreContracts := []string{}
		customDeps := []flowkitConfig.Dependency{
			{
				Name: "FlowTransactionScheduler",
				Source: flowkitConfig.Source{
					NetworkName:  flowkitConfig.TestnetNetwork.Name,
					Address:      flowsdk.HexToAddress("8c5303eaa26202d6"),
					ContractName: "FlowTransactionScheduler",
				},
				Aliases: flowkitConfig.Aliases{
					{
						Network: "emulator",
						Address: flowsdk.HexToAddress("f8d6e0586b0a20c7"),
					},
				},
			},
			{
				Name: "FlowTransactionSchedulerUtils",
				Source: flowkitConfig.Source{
					NetworkName:  flowkitConfig.TestnetNetwork.Name,
					Address:      flowsdk.HexToAddress("8c5303eaa26202d6"),
					ContractName: "FlowTransactionSchedulerUtils",
				},
				Aliases: flowkitConfig.Aliases{
					{
						Network: "emulator",
						Address: flowsdk.HexToAddress("f8d6e0586b0a20c7"),
					},
				},
			},
		}
		err := installProjectDependencies(logger, state, tempDir, coreContracts, customDeps)
		if err != nil {
			return "", err
		}
	}

	templates := getProjectTemplates(projectType, targetDir, state)

	g := generator.NewGenerator(tempDir, state, logger, true, false)
	err = g.Create(templates...)
	if err != nil {
		return "", err
	}

	// Add project-specific contract deployments for scheduled transactions
	if projectType == ProjectTypeScheduledTransactions {
		err = addContractDeployments(state, []string{"Counter", "CounterTransactionHandler"}, "emulator-account")
		if err != nil {
			return "", err
		}
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

	// Colorized success message
	successMsg := branding.GreenStyle.Render(fmt.Sprintf("%s Congrats! your project was created.", output.SuccessEmoji()))
	out.WriteString(fmt.Sprintf("%s\n\n", successMsg))

	// Check if we created README_flow.md instead of README.md
	readmeFile := defaultReadmeFile
	if _, err := os.Stat(filepath.Join(s.targetDir, flowReadmeFile)); err == nil {
		readmeFile = flowReadmeFile
		noteMsg := branding.PurpleStyle.Render("📝 Note: Created README_flow.md since README.md already exists.")
		out.WriteString(fmt.Sprintf("%s\n\n", noteMsg))
	}

	// Colorized section header
	headerMsg := branding.PurpleStyle.Render("Start development by following these steps:")
	out.WriteString(fmt.Sprintf("%s\n", headerMsg))

	// Only show cd command if not current directory
	if s.targetDir != wd {
		cdCmd := branding.GreenStyle.Render(fmt.Sprintf("cd %s", relDir))
		emulatorCmd := branding.GreenStyle.Render("flow emulator")
		testCmd := branding.GreenStyle.Render("flow test")
		out.WriteString(fmt.Sprintf("1. '%s' to change to your new project,\n", cdCmd))
		out.WriteString(fmt.Sprintf("2. '%s' to start the emulator,\n", emulatorCmd))
		out.WriteString(fmt.Sprintf("3. '%s' to test your project.\n\n", testCmd))
	} else {
		emulatorCmd := branding.GreenStyle.Render("flow emulator")
		testCmd := branding.GreenStyle.Render("flow test")
		out.WriteString(fmt.Sprintf("1. '%s' to start the emulator,\n", emulatorCmd))
		out.WriteString(fmt.Sprintf("2. '%s' to test your project.\n\n", testCmd))
	}

	// Colorized footer message
	readmeMsg := branding.GrayStyle.Render(fmt.Sprintf("You should also read %s to learn more about the development process!", readmeFile))
	out.WriteString(fmt.Sprintf("%s\n", readmeMsg))

	return out.String()
}

func (s *initResult) Oneliner() string {
	return fmt.Sprintf("Project created inside %s", s.targetDir)
}

func (s *initResult) JSON() any {
	return nil
}
