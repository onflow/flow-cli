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
	"text/template"

	"github.com/onflow/flow-cli/internal/util"

	flowsdk "github.com/onflow/flow-go-sdk"

	"github.com/onflow/flowkit/config"

	"github.com/onflow/flowkit"

	"github.com/onflow/flowkit/output"

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

const (
	DefaultCadenceDirectory = "cadence"
	ContractType            = "contract"
	TransactionType         = "transaction"
	ScriptType              = "script"
)

func generateContract(
	args []string,
	_ command.GlobalFlags,
	logger output.Logger,
	_ flowkit.Services,
	state *flowkit.State,
) (result command.Result, err error) {
	options := GeneratorOptions{
		Directory: DefaultCadenceDirectory,
		State:     state,
		Logger:    logger,
	}
	generator := NewGenerator(options)
	err = generator.Create(TemplateMap{ContractType: args[0]})
	return nil, err
}

func generateTransaction(
	args []string,
	_ command.GlobalFlags,
	logger output.Logger,
	_ flowkit.Services,
	state *flowkit.State,
) (result command.Result, err error) {
	options := GeneratorOptions{
		Directory: DefaultCadenceDirectory,
		State:     state,
		Logger:    logger,
	}
	generator := NewGenerator(options)
	err = generator.Create(TemplateMap{TransactionType: args[0]})
	return nil, err
}

func generateScript(
	args []string,
	_ command.GlobalFlags,
	logger output.Logger,
	_ flowkit.Services,
	state *flowkit.State,
) (result command.Result, err error) {
	options := GeneratorOptions{
		Directory: DefaultCadenceDirectory,
		State:     state,
		Logger:    logger,
	}
	generator := NewGenerator(options)
	err = generator.Create(TemplateMap{ScriptType: args[0]})
	return nil, err
}

// TemplateMap defines a map of template types to their specific names
type TemplateMap map[string]string

type GeneratorOptions struct {
	Directory   string
	State       *flowkit.State
	Logger      output.Logger
	DisableLogs bool
}

type Generator struct {
	Options GeneratorOptions
}

func NewGenerator(options GeneratorOptions) *Generator {
	return &Generator{
		Options: options,
	}
}

func (g *Generator) Create(typeNames TemplateMap) error {
	for templateType, name := range typeNames {
		err := g.generate(templateType, name)
		if err != nil {
			return err
		}
	}
	return nil
}

func (g *Generator) generate(templateType, name string) error {

	name = util.StripCDCExtension(name)
	filename := util.AddCDCExtension(name)

	var fileToWrite string
	var testFileToWrite string
	var rootDir = DefaultCadenceDirectory
	var basePath string
	var testsBasePath = "tests"
	var err error

	if g.Options.Directory != "" {
		rootDir = g.Options.Directory
	}

	switch templateType {
	case ContractType:
		basePath = "contracts"
		nameData := map[string]interface{}{"Name": name}
		fileToWrite, err = processTemplate("templates/contract_init.cdc.tmpl", nameData)
		if err != nil {
			return fmt.Errorf("error generating contract template: %w", err)
		}

		testFileToWrite, err = processTemplate("templates/contract_init_test.cdc.tmpl", nameData)
		if err != nil {
			return fmt.Errorf("error generating contract test template: %w", err)
		}
	case ScriptType:
		basePath = "scripts"
		fileToWrite, err = processTemplate("templates/script_init.cdc.tmpl", nil)
		if err != nil {
			return fmt.Errorf("error generating script template: %w", err)
		}
	case TransactionType:
		basePath = "transactions"
		fileToWrite, err = processTemplate("templates/transaction_init.cdc.tmpl", nil)
		if err != nil {
			return fmt.Errorf("error generating transaction template: %w", err)
		}
	default:
		return fmt.Errorf("invalid template type: %s", templateType)
	}

	directoryWithBasePath := filepath.Join(rootDir, basePath)
	filenameWithBasePath := filepath.Join(rootDir, basePath, filename)

	// Check file existence
	if _, err := g.Options.State.ReaderWriter().ReadFile(filenameWithBasePath); err == nil {
		return fmt.Errorf("file already exists: %s", filenameWithBasePath)
	}

	// Ensure the directory exists
	if err := g.Options.State.ReaderWriter().MkdirAll(directoryWithBasePath, 0755); err != nil {
		return fmt.Errorf("error creating directories: %w", err)
	}

	// Write files
	err = g.Options.State.ReaderWriter().WriteFile(filenameWithBasePath, []byte(fileToWrite), 0644)
	if err != nil {
		return fmt.Errorf("error writing file: %w", err)
	}

	if !g.Options.DisableLogs {
		g.Options.Logger.Info(fmt.Sprintf("Generated new %s: %s at %s", templateType, name, filenameWithBasePath))
	}

	if generateFlags.SkipTests != true && templateType == ContractType {
		testDirectoryWithBasePath := filepath.Join(rootDir, testsBasePath)
		testFilenameWithBasePath := filepath.Join(rootDir, testsBasePath, util.AddCDCExtension(fmt.Sprintf("%s_test", name)))

		if _, err := g.Options.State.ReaderWriter().ReadFile(testFilenameWithBasePath); err == nil {
			return fmt.Errorf("file already exists: %s", testFilenameWithBasePath)
		}

		if err := g.Options.State.ReaderWriter().MkdirAll(testDirectoryWithBasePath, 0755); err != nil {
			return fmt.Errorf("error creating test directory: %w", err)
		}

		err := g.Options.State.ReaderWriter().WriteFile(testFilenameWithBasePath, []byte(testFileToWrite), 0644)
		if err != nil {
			return fmt.Errorf("error writing test file: %w", err)
		}

		if !g.Options.DisableLogs {
			g.Options.Logger.Info(fmt.Sprintf("Generated new test file: %s at %s", name, testFilenameWithBasePath))
		}
	}

	if templateType == ContractType {
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
		g.Options.State.Contracts().AddOrUpdate(contract)
		err := g.Options.State.SaveDefault()
		if err != nil {
			return fmt.Errorf("error saving to flow.json: %w", err)
		}
	}

	return nil
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
