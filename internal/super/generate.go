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
	"embed"

	"github.com/onflow/flow-cli/internal/super/generator"
	"github.com/onflow/flow-cli/internal/util"

	"github.com/onflow/flowkit/v2"

	"github.com/onflow/flowkit/v2/output"

	"github.com/onflow/flow-cli/internal/command"

	"github.com/spf13/afero"
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
	g := generator.NewGenerator(getTemplateFs(), "", state, logger, false, true)
	name := util.StripCDCExtension(args[0])
	err = g.Create(generator.ContractTemplate{Name: name})
	return nil, err
}

func generateTransaction(
	args []string,
	_ command.GlobalFlags,
	logger output.Logger,
	_ flowkit.Services,
	state *flowkit.State,
) (result command.Result, err error) {
	g := generator.NewGenerator(getTemplateFs(), "", state, logger, false, true)
	name := util.StripCDCExtension(args[0])
	err = g.Create(generator.TransactionTemplate{Name: name})
	return nil, err
}

func generateScript(
	args []string,
	_ command.GlobalFlags,
	logger output.Logger,
	_ flowkit.Services,
	state *flowkit.State,
) (result command.Result, err error) {
	g := generator.NewGenerator(getTemplateFs(), "", state, logger, false, true)
	name := util.StripCDCExtension(args[0])
	err = g.Create(generator.ScriptTemplate{Name: name})
	return nil, err
}

func getTemplateFs() *afero.Afero {
	fs := afero.Afero{Fs: afero.FromIOFS{FS: templatesFS}}
	return &afero.Afero{Fs: afero.NewBasePathFs(fs, "templates")}
}
