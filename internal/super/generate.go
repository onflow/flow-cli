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
	"embed"
	"fmt"
	"path/filepath"
	"strings"
	"text/template"

	flowsdk "github.com/onflow/flow-go-sdk"

	"github.com/onflow/flowkit/v2/config"

	"github.com/onflow/flowkit/v2"

	"github.com/onflow/flowkit/v2/output"

	"github.com/onflow/flow-cli/internal/command"

	"github.com/spf13/cobra"
)

//go:embed templates/*.tmpl
var templatesFS embed.FS

type generateFlagsDef struct {
	Directory string `default:"" flag:"dir" info:"Directory to generate files in"`
	SkipTests bool   `default:"false" flag:"skip-tests" info:"Skip generating test files"`
}

var generateFlags = generateFlagsDef{}

var GenerateCommand = &cobra.Command{
	Use:     "generate",
	Short:   "Generate template files for common Cadence code",
	GroupID: "super",
	Aliases: []string{"g"},
}

var GenerateContractCommand = &command.Command{
	Cmd: &cobra.Command{
		Use:     "contract <name>",
		Short:   "Generate Cadence smart contract template",
		Example: "flow generate contract HelloWorld",
		Args:    cobra.ExactArgs(1),
	},
	Flags: &generateFlags,
	RunS:  generateContract,
}

var GenerateTransactionCommand = &command.Command{
	Cmd: &cobra.Command{
		Use:     "transaction <name>",
		Short:   "Generate a Cadence transaction template",
		Example: "flow generate transaction SomeTransaction",
		Args:    cobra.ExactArgs(1),
	},
	Flags: &generateFlags,
	RunS:  generateTransaction,
}

var GenerateScriptCommand = &command.Command{
	Cmd: &cobra.Command{
		Use:     "script <name>",
		Short:   "Generate a Cadence script template",
		Example: "flow generate script SomeScript",
		Args:    cobra.ExactArgs(1),
	},
	Flags: &generateFlags,
	RunS:  generateScript,
}

func init() {
	GenerateContractCommand.AddToParent(GenerateCommand)
	GenerateTransactionCommand.AddToParent(GenerateCommand)
	GenerateScriptCommand.AddToParent(GenerateCommand)
}

func generateContract(
	args []string,
	_ command.GlobalFlags,
	logger output.Logger,
	flow flowkit.Services,
	state *flowkit.State,
) (result command.Result, err error) {
	return generateNew(args, "contract", logger, state)
}

func generateTransaction(
	args []string,
	_ command.GlobalFlags,
	logger output.Logger,
	flow flowkit.Services,
	state *flowkit.State,
) (result command.Result, err error) {
	return generateNew(args, "transaction", logger, state)
}

func generateScript(
	args []string,
	_ command.GlobalFlags,
	logger output.Logger,
	flow flowkit.Services,
	state *flowkit.State,
) (result command.Result, err error) {
	return generateNew(args, "script", logger, state)
}

func addCDCExtension(name string) string {
	if strings.HasSuffix(name, ".cdc") {
		return name
	}
	return fmt.Sprintf("%s.cdc", name)
}

func stripCDCExtension(name string) string {
	return strings.TrimSuffix(name, filepath.Ext(name))
}

// processTemplate reads a template file from the embedded filesystem and processes it with the provided data
// If you don't need to provide data, pass nil
func processTemplate(templatePath string, data map[string]interface{}) (string, error) {
	templateData, err := templatesFS.ReadFile(templatePath)
	if err != nil {
		return "", fmt.Errorf("failed to read template file: %w", err)
	}

	tmpl, err := template.New("template").Parse(string(templateData))
	if err != nil {
		return "", fmt.Errorf("failed to parse template: %w", err)
	}

	var executedTemplate bytes.Buffer
	// Execute the template with the provided data or nil if no data is needed
	if err = tmpl.Execute(&executedTemplate, data); err != nil {
		return "", fmt.Errorf("failed to execute template: %w", err)
	}

	return executedTemplate.String(), nil
}

func generateNew(
	args []string,
	templateType string,
	logger output.Logger,
	state *flowkit.State,
) (result command.Result, err error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("invalid number of arguments")
	}

	name := stripCDCExtension(args[0])
	filename := addCDCExtension(name)

	var fileToWrite string
	var testFileToWrite string
	var basePath string

	if generateFlags.Directory != "" {
		basePath = generateFlags.Directory
	} else {
		switch templateType {
		case "contract":
			basePath = "cadence/contracts"
		case "script":
			basePath = "cadence/scripts"
		case "transaction":
			basePath = "cadence/transactions"
		default:
			return nil, fmt.Errorf("invalid template type: %s", templateType)
		}
	}

	switch templateType {
	case "contract":
		nameData := map[string]interface{}{"Name": name}
		fileToWrite, err = processTemplate("templates/contract_init.cdc.tmpl", nameData)
		if err != nil {
			return nil, fmt.Errorf("error generating contract template: %w", err)
		}

		testFileToWrite, err = processTemplate("templates/contract_init_test.cdc.tmpl", nameData)
		if err != nil {
			return nil, fmt.Errorf("error generating contract test template: %w", err)
		}
	case "script":
		fileToWrite, err = processTemplate("templates/script_init.cdc.tmpl", nil)
		if err != nil {
			return nil, fmt.Errorf("error generating script template: %w", err)
		}
	case "transaction":
		fileToWrite, err = processTemplate("templates/transaction_init.cdc.tmpl", nil)
		if err != nil {
			return nil, fmt.Errorf("error generating transaction template: %w", err)
		}
	default:
		return nil, fmt.Errorf("invalid template type: %s", templateType)
	}

	filenameWithBasePath := filepath.Join(basePath, filename)

	// Check file existence
	if _, err := state.ReaderWriter().ReadFile(filenameWithBasePath); err == nil {
		return nil, fmt.Errorf("file already exists: %s", filenameWithBasePath)
	}

	// Ensure the directory exists
	if err := state.ReaderWriter().MkdirAll(basePath, 0755); err != nil {
		return nil, fmt.Errorf("error creating directories: %w", err)
	}

	// Write files
	err = state.ReaderWriter().WriteFile(filenameWithBasePath, []byte(fileToWrite), 0644)
	if err != nil {
		return nil, fmt.Errorf("error writing file: %w", err)
	}

	logger.Info(fmt.Sprintf("Generated new %s: %s at %s", templateType, name, filenameWithBasePath))

	if generateFlags.SkipTests != true && templateType == "contract" {
		testsBasePath := "cadence/tests"
		testFilenameWithBasePath := filepath.Join(testsBasePath, addCDCExtension(fmt.Sprintf("%s_test", name)))

		if _, err := state.ReaderWriter().ReadFile(testFilenameWithBasePath); err == nil {
			return nil, fmt.Errorf("file already exists: %s", testFilenameWithBasePath)
		}

		if err := state.ReaderWriter().MkdirAll(testsBasePath, 0755); err != nil {
			return nil, fmt.Errorf("error creating test directory: %w", err)
		}

		err = state.ReaderWriter().WriteFile(testFilenameWithBasePath, []byte(testFileToWrite), 0644)
		if err != nil {
			return nil, fmt.Errorf("error writing test file: %w", err)
		}

		logger.Info(fmt.Sprintf("Generated new test file: %s at %s", name, testFilenameWithBasePath))
	}

	if templateType == "contract" {
		var aliases config.Aliases

		if generateFlags.SkipTests != true {
			aliases = config.Aliases{{
				Network: config.TestingNetwork.Name,
				Address: flowsdk.HexToAddress("0x0000000000000007"),
			}}
		}

		contract := config.Contract{
			Name:     name,
			Location: filenameWithBasePath,
			Aliases:  aliases,
		}
		state.Contracts().AddOrUpdate(contract)
		err = state.SaveDefault()
		if err != nil {
			return nil, fmt.Errorf("error saving to flow.json: %w", err)
		}
	}

	return nil, err
}
