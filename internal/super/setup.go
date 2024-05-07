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

	"github.com/spf13/afero"

	"github.com/onflow/flow-cli/internal/util"

	flowsdk "github.com/onflow/flow-go-sdk"
	"golang.org/x/exp/slices"

	"github.com/onflow/flow-cli/internal/prompt"

	"github.com/onflow/flow-go/fvm/systemcontracts"
	flowGo "github.com/onflow/flow-go/model/flow"
	flowkitConfig "github.com/onflow/flowkit/config"

	"github.com/spf13/cobra"

	"github.com/onflow/flow-cli/internal/dependencymanager"

	"github.com/onflow/flowkit"
	"github.com/onflow/flowkit/output"

	"github.com/onflow/flow-cli/internal/config"

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
	Run:   create,
}

func create(
	args []string,
	_ command.GlobalFlags,
	logger output.Logger,
	_ flowkit.ReaderWriter,
	_ flowkit.Services,
) (command.Result, error) {
	var targetDir string
	var err error

	if setupFlags.Scaffold || setupFlags.ScaffoldID != 0 {
		// Error if no project name is given
		if len(args) < 1 || args[0] == "" {
			return nil, fmt.Errorf("no project name provided")
		}

		targetDir, err = handleScaffold(args[0], logger)
		if err != nil {
			return nil, err
		}
	} else {
		targetDir, err = startInteractiveSetup(args, logger)
		if err != nil {
			return nil, err
		}
	}

	return &setupResult{targetDir: targetDir}, nil
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
	//state, err := flowkit.Init(rw, crypto.ECDSA_P256, crypto.SHA3_256)
	//if err != nil {
	//	return "", err
	//}

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

	// Generate standard cadence files
	// cadence/contracts/DefaultContract.cdc
	// cadence/scripts/DefaultScript.cdc
	// cadence/transactions/DefaultTransaction.cdc
	// cadence/tests/DefaultContract_test.cdc

	directoryPath := filepath.Join(tempDir, "cadence")

	templates := TemplateMap{
		"contract": []TemplateItem{
			Contract{Name: "DefaultContract", Account: "example-account"},
		},
		"script": []TemplateItem{
			OtherTemplate{Name: "DefaultScripts"},
		},
		"transaction": []TemplateItem{
			OtherTemplate{Name: "DefaultTransaction"},
		},
	}

	generator := NewGenerator(directoryPath, state, logger, true, false)
	err = generator.Create(templates)
	if err != nil {
		return "", err
	}

	// Prompt to ask which core contracts should be installed
	sc := systemcontracts.SystemContractsForChain(flowGo.Mainnet)
	promptMessage := "Select the core contracts you'd like to install:"

	contractNames := make([]string, 0)

	for _, contract := range sc.All() {
		contractNames = append(contractNames, contract.Name)
	}

	selectedContractNames, err := prompt.RunSelectOptions(contractNames, promptMessage)
	if err != nil {
		return "", fmt.Errorf("error running dependency selection: %v\n", err)
	}

	var dependencies []flowkitConfig.Dependency

	// Loop standard contracts and add them to the dependencies if selected
	for _, contract := range sc.All() {
		if slices.Contains(selectedContractNames, contract.Name) {
			dependencies = append(dependencies, flowkitConfig.Dependency{
				Name: contract.Name,
				Source: flowkitConfig.Source{
					NetworkName:  flowkitConfig.MainnetNetwork.Name,
					Address:      flowsdk.HexToAddress(contract.Address.String()),
					ContractName: contract.Name,
				},
			})
		}
	}

	logger.Info(util.MessageWithEmojiPrefix("🔄", "Installing selected core contracts and dependencies..."))

	// Add the selected core contracts as dependencies
	installer, err := dependencymanager.NewDependencyInstaller(logger, state, false, tempDir, dependencymanager.Flags{})
	if err != nil {
		return "", err
	}

	if err := installer.AddMany(dependencies); err != nil {
		return "", err
	}

	err = state.Save(filepath.Join(tempDir, "flow.json"))
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
