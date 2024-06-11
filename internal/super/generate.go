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

	flowsdk "github.com/onflow/flow-go-sdk"
	"github.com/onflow/flowkit/v2/config"

	"github.com/onflow/flow-cli/internal/util"

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
	generator := NewGenerator("", state, logger, false, true)
	contract := Contract{Name: args[0], Account: ""}
	err = generator.Create(TemplateMap{ContractType: []TemplateItem{contract}})
	return nil, err
}

func generateTransaction(
	args []string,
	_ command.GlobalFlags,
	logger output.Logger,
	_ flowkit.Services,
	state *flowkit.State,
) (result command.Result, err error) {
	generator := NewGenerator("", state, logger, false, true)
	transaction := ScriptTemplate{Name: args[0]}
	err = generator.Create(TemplateMap{TransactionType: []TemplateItem{transaction}})
	return nil, err
}

func generateScript(
	args []string,
	_ command.GlobalFlags,
	logger output.Logger,
	_ flowkit.Services,
	state *flowkit.State,
) (result command.Result, err error) {
	generator := NewGenerator("", state, logger, false, true)
	script := ScriptTemplate{Name: args[0]}
	err = generator.Create(TemplateMap{ScriptType: []TemplateItem{script}})
	return nil, err
}

// TemplateItem is an interface for different template types
type TemplateItem interface {
	GetName() string
	GetTemplate() string
	GetAccount() string
	GetData() map[string]interface{}
}

// Contract contains properties for contracts
type Contract struct {
	Name     string
	Template string
	Account  string
	Data     map[string]interface{}
}

// GetName returns the name of the contract
func (c Contract) GetName() string {
	return c.Name
}

// GetTemplate returns the template of the contract
func (c Contract) GetTemplate() string {
	if c.Template == "" {
		return "contract_init"
	}

	return c.Template
}

// GetAccount returns the account of the contract
func (c Contract) GetAccount() string {
	return c.Account
}

// GetData returns the data of the contract
func (c Contract) GetData() map[string]interface{} {
	return c.Data
}

// ScriptTemplate contains only a name property for scripts and transactions
type ScriptTemplate struct {
	Name     string
	Template string
	Data     map[string]interface{}
}

// GetName returns the name of the script or transaction
func (o ScriptTemplate) GetName() string {
	return o.Name
}

// GetTemplate returns an empty string for scripts and transactions
func (o ScriptTemplate) GetTemplate() string {
	if o.Template == "" {
		return "script_init"
	}

	return o.Template
}

// GetAccount returns an empty string for scripts and transactions
func (o ScriptTemplate) GetAccount() string {
	return ""
}

// GetData returns the data of the script or transaction
func (o ScriptTemplate) GetData() map[string]interface{} {
	return o.Data
}

// TransactionTemplate contains only a name property for scripts and transactions
type TransactionTemplate struct {
	Name     string
	Template string
	Data     map[string]interface{}
}

// GetName returns the name of the script or transaction
func (o TransactionTemplate) GetName() string {
	return o.Name
}

// GetTemplate returns an empty string for scripts and transactions
func (o TransactionTemplate) GetTemplate() string {
	if o.Template == "" {
		return "transaction_init"
	}

	return o.Template
}

// GetAccount returns an empty string for scripts and transactions
func (o TransactionTemplate) GetAccount() string {
	return ""
}

// GetData returns the data of the script or transaction
func (o TransactionTemplate) GetData() map[string]interface{} {
	return o.Data
}

// TemplateMap holds all templates with flexibility
type TemplateMap map[string][]TemplateItem

type Generator struct {
	directory   string
	state       *flowkit.State
	logger      output.Logger
	disableLogs bool
	saveState   bool
}

func NewGenerator(directory string, state *flowkit.State, logger output.Logger, disableLogs, saveState bool) *Generator {
	return &Generator{
		directory:   directory,
		state:       state,
		logger:      logger,
		disableLogs: disableLogs,
		saveState:   saveState,
	}
}

func (g *Generator) Create(typeNames TemplateMap) error {
	for templateType, items := range typeNames {
		for _, item := range items {
			err := g.generate(templateType, item.GetTemplate(), item.GetName(), item.GetAccount(), item.GetData())
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (g *Generator) generate(templateType, templateName, name, account string, data map[string]interface{}) error {

	name = util.StripCDCExtension(name)
	filename := util.AddCDCExtension(name)

	var fileToWrite string
	var testFileToWrite string
	var rootDir string
	var basePath string
	var testsBasePath = filepath.Join(DefaultCadenceDirectory, "tests")
	var err error

	if g.directory != "" {
		rootDir = g.directory
	}

	templatePath := fmt.Sprintf("templates/%s.cdc.tmpl", templateName)

	switch templateType {
	case ContractType:
		basePath = filepath.Join(DefaultCadenceDirectory, "contracts")
		fileData := map[string]interface{}{"Name": name}
		for k, v := range data {
			fileData[k] = v
		}
		fileToWrite, err = processTemplate(templatePath, fileData)
		if err != nil {
			return fmt.Errorf("error generating contract template: %w", err)
		}

		testFileToWrite, err = processTemplate("templates/contract_init_test.cdc.tmpl", fileData)
		if err != nil {
			return fmt.Errorf("error generating contract test template: %w", err)
		}
	case ScriptType:
		basePath = filepath.Join(DefaultCadenceDirectory, "scripts")
		fileData := map[string]interface{}{}
		for k, v := range data {
			fileData[k] = v
		}
		fileToWrite, err = processTemplate(templatePath, fileData)
		if err != nil {
			return fmt.Errorf("error generating script template: %w", err)
		}
	case TransactionType:
		basePath = filepath.Join(DefaultCadenceDirectory, "transactions")
		fileData := map[string]interface{}{}
		for k, v := range data {
			fileData[k] = v
		}
		fileToWrite, err = processTemplate(templatePath, fileData)
		if err != nil {
			return fmt.Errorf("error generating transaction template: %w", err)
		}
	default:
		return fmt.Errorf("invalid template type: %s", templateType)
	}

	fmt.Println("account: ", account)

	directoryWithBasePath := filepath.Join(rootDir, basePath, account)
	filenameWithBasePath := filepath.Join(rootDir, basePath, account, filename)
	relativeFilenameWithBasePath := filepath.Join(basePath, account, filename)

	// Check file existence
	if _, err := g.state.ReaderWriter().ReadFile(filenameWithBasePath); err == nil {
		return fmt.Errorf("file already exists: %s", filenameWithBasePath)
	}

	// Ensure the directory exists
	if err := g.state.ReaderWriter().MkdirAll(directoryWithBasePath, 0755); err != nil {
		return fmt.Errorf("error creating directories: %w", err)
	}

	// Write files
	err = g.state.ReaderWriter().WriteFile(filenameWithBasePath, []byte(fileToWrite), 0644)
	if err != nil {
		return fmt.Errorf("error writing file: %w", err)
	}

	if !g.disableLogs {
		g.logger.Info(fmt.Sprintf("Generated new %s: %s at %s", templateType, name, filenameWithBasePath))
	}

	if generateFlags.SkipTests != true && templateType == ContractType {
		testDirectoryWithBasePath := filepath.Join(rootDir, testsBasePath)
		testFilenameWithBasePath := filepath.Join(rootDir, testsBasePath, util.AddCDCExtension(fmt.Sprintf("%s_test", name)))

		if _, err := g.state.ReaderWriter().ReadFile(testFilenameWithBasePath); err == nil {
			return fmt.Errorf("file already exists: %s", testFilenameWithBasePath)
		}

		if err := g.state.ReaderWriter().MkdirAll(testDirectoryWithBasePath, 0755); err != nil {
			return fmt.Errorf("error creating test directory: %w", err)
		}

		err := g.state.ReaderWriter().WriteFile(testFilenameWithBasePath, []byte(testFileToWrite), 0644)
		if err != nil {
			return fmt.Errorf("error writing test file: %w", err)
		}

		if !g.disableLogs {
			g.logger.Info(fmt.Sprintf("Generated new test file: %s at %s", name, testFilenameWithBasePath))
		}
	}

	if templateType == ContractType {
		fmt.Println("directoryWithBasePath: ", directoryWithBasePath)
		fmt.Println("filenameWithBasePath: ", filenameWithBasePath)
		fmt.Println("relativeFilenameWithBasePath: ", relativeFilenameWithBasePath)
		err := g.updateContractsState(name, relativeFilenameWithBasePath)
		if err != nil {
			return err
		}
	}

	return nil
}

func (g *Generator) updateContractsState(name, location string) error {
	var aliases config.Aliases

	if generateFlags.SkipTests != true {
		aliases = config.Aliases{{
			Network: config.TestingNetwork.Name,
			Address: flowsdk.HexToAddress("0x0000000000000007"),
		}}
	}

	contract := config.Contract{
		Name:     name,
		Location: location,
		Aliases:  aliases,
	}

	g.state.Contracts().AddOrUpdate(contract)

	if g.saveState {
		err := g.state.SaveDefault() // TODO: Support adding a target project directory
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
