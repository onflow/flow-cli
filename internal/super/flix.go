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
	"context"
	"fmt"
	"net/url"
	"os"
	"path/filepath"

	"github.com/onflow/flixkit-go/flixkit"

	"github.com/onflow/flow-cli/flowkit"
	"github.com/onflow/flow-cli/flowkit/config"
	"github.com/onflow/flow-cli/flowkit/output"
	"github.com/onflow/flow-cli/internal/command"
	"github.com/onflow/flow-cli/internal/scripts"
	"github.com/onflow/flow-cli/internal/transactions"

	"github.com/spf13/cobra"
)

type flixFlags struct {
	ArgsJSON    string   `default:"" flag:"args-json" info:"arguments in JSON-Cadence format"`
	BlockID     string   `default:"" flag:"block-id" info:"block ID to execute the script at"`
	BlockHeight uint64   `default:"" flag:"block-height" info:"block height to execute the script at"`
	Signer      string   `default:"" flag:"signer" info:"Account name from configuration used to sign the transaction as proposer, payer and suthorizer"`
	Proposer    string   `default:"" flag:"proposer" info:"Account name from configuration used as proposer"`
	Payer       string   `default:"" flag:"payer" info:"Account name from configuration used as payer"`
	Authorizers []string `default:"" flag:"authorizer" info:"Name of a single or multiple comma-separated accounts used as authorizers from configuration"`
	Include     []string `default:"" flag:"include" info:"Fields to include in the output"`
	Exclude     []string `default:"" flag:"exclude" info:"Fields to exclude from the output (events)"`
	GasLimit    uint64   `default:"1000" flag:"gas-limit" info:"transaction gas limit"`
	PreFill     string   `default:"" flag:"pre-fill" info:"template path to pre fill the FLIX"`
	Lang        string   `default:"js" flag:"lang" info:"language to generate the template for"`
}

type flixResult struct {
	flixQuery string
	result    string
}

var flags = flixFlags{}
var FlixCmd = &cobra.Command{
	Use:              "flix",
	Short:            "execute, generate, package",
	TraverseChildren: true,
	GroupID:          "tools",
}

var executeCommand = &command.Command{
	Cmd: &cobra.Command{
		Use:     "execute <id | name | path | url>",
		Short:   "execute FLIX template with a given id, name, local filename, or url",
		Example: "flow flix execute transfer-flow 1 0x123456789",
		Args:    cobra.MinimumNArgs(1),
	},
	Flags: &flags,
	RunS:  executeCmd,
}

var packageCommand = &command.Command{
	Cmd: &cobra.Command{
		Use:     "package <id | name | path | url> --lang <lang>",
		Short:   "package file for FLIX template fcl-js is default",
		Example: "flow flix package multiply.template.json --lang js",
		Args:    cobra.MinimumNArgs(1),
	},
	Flags: &flags,
	RunS:  packageCmd,
}

var generateCommand = &command.Command{
	Cmd: &cobra.Command{
		Use:     "generate <cadence.cdc>",
		Short:   "generate FLIX json template given local Cadence filename",
		Example: "flow flix generate multiply.cdc",
		Args:    cobra.MinimumNArgs(1),
	},
	Flags: &flags,
	RunS:  generateCmd,
}

func init() {
	executeCommand.AddToParent(FlixCmd)
	packageCommand.AddToParent(FlixCmd)
	generateCommand.AddToParent(FlixCmd)
}

func executeCmd(
	args []string,
	_ command.GlobalFlags,
	logger output.Logger,
	flow flowkit.Services,
	state *flowkit.State,
) (result command.Result, err error) {
	flixService := flixkit.NewFlixService(&flixkit.Config{
		FileReader: state,
	})
	flixQuery := args[0]
	ctx := context.Background()
	cadenceWithImportsReplaced, err := flixService.GetAndReplaceCadenceImports(ctx, flixQuery, flow.Network().Name)
	if err != nil {
		logger.Error("could not replace imports")
		return nil, err
	}

	if cadenceWithImportsReplaced.IsScript {
		scriptsFlags := scripts.Flags{
			ArgsJSON:    flags.ArgsJSON,
			BlockID:     flags.BlockID,
			BlockHeight: flags.BlockHeight,
		}
		return scripts.SendScript([]byte(cadenceWithImportsReplaced.Cadence), args[1:], "", flow, scriptsFlags)
	}

	transactionFlags := transactions.Flags{
		ArgsJSON:    flags.ArgsJSON,
		Signer:      flags.Signer,
		Proposer:    flags.Proposer,
		Payer:       flags.Payer,
		Authorizers: flags.Authorizers,
		Include:     flags.Include,
		Exclude:     flags.Exclude,
		GasLimit:    flags.GasLimit,
	}
	return transactions.SendTransaction([]byte(cadenceWithImportsReplaced.Cadence), args[1:], "", flow, state, transactionFlags)
}

func packageCmd(
	args []string,
	gFlags command.GlobalFlags,
	logger output.Logger,
	flow flowkit.Services,
	state *flowkit.State,
) (result command.Result, err error) {
	flixService := flixkit.NewFlixService(&flixkit.Config{
		FileReader: state,
	})
	flixQuery := args[0]
	ctx := context.Background()
	template, err := flixService.GetTemplate(ctx, flixQuery)
	if err != nil {
		return nil, err
	}
	if !isUrl(flixQuery) {
		if gFlags.Save != "" {
			// resolve template file location to relative path to be used by binding file
			flixQuery, err = GetRelativePath(flixQuery, gFlags.Save)
			if err != nil {
				logger.Error("could not resolve relative path to template")
				return nil, err
			}
		}
	}

	var out string
	var gen flixkit.FclGenerator
	switch flags.Lang {
	case "js":
		gen = *flixkit.NewFclJSGenerator()
	case "ts":
		gen = *flixkit.NewFclTSGenerator()
	default:
		return nil, fmt.Errorf("language %s not supported", flags.Lang)
	}
	out, err = gen.Generate(template, flixQuery)
	if err != nil {
		return nil, err
	}

	return &flixResult{
		flixQuery: flixQuery,
		result:    out,
	}, err
}

func generateCmd(
	args []string,
	_ command.GlobalFlags,
	logger output.Logger,
	flow flowkit.Services,
	state *flowkit.State,
) (result command.Result, err error) {
	cadenceFile := args[0]

	if cadenceFile == "" {
		return nil, fmt.Errorf("no cadence code found")
	}

	code, err := state.ReadFile(cadenceFile)
	if err != nil {
		return nil, fmt.Errorf("could not read cadence file %s: %w", cadenceFile, err)
	}

	depContracts := GetDeployedContracts(state)
	generator, err := flixkit.NewGenerator(depContracts, logger)
	if err != nil {
		return nil, fmt.Errorf("could not create flix generator %w", err)
	}

	ctx := context.Background()
	var template string
	if flags.PreFill != "" {
		flixService := flixkit.NewFlixService(&flixkit.Config{
			FileReader: state,
		})
		template, err = flixService.GetTemplate(ctx, flags.PreFill)
		if err != nil {
			return nil, fmt.Errorf("could not parse template from pre fill %w", err)
		}
	}

	prettyJSON, err := generator.Generate(ctx, string(code), template)
	if err != nil {
		return nil, fmt.Errorf("could not generate flix %w", err)
	}

	return &flixResult{
		flixQuery: cadenceFile,
		result:    prettyJSON,
	}, err

}

func (fr *flixResult) JSON() any {
	result := make(map[string]any)
	result["flixQuery"] = fr.flixQuery
	result["result"] = fr.result
	return result
}

func (fr *flixResult) String() string {
	return fr.result
}

func (fr *flixResult) Oneliner() string {
	return fr.result
}

// GetRelativePath computes the relative path from generated file to flix json file.
// This path is used in the binding file to reference the flix json file.
func GetRelativePath(configFile, bindingFile string) (string, error) {
	relPath, err := filepath.Rel(filepath.Dir(bindingFile), configFile)
	if err != nil {
		return "", err
	}

	// If the file is in the same directory and doesn't start with "./", prepend it.
	if !filepath.IsAbs(relPath) && relPath[0] != '.' {
		relPath = "./" + relPath
	}

	// Currently binding files are js, we need to convert the path to unix style
	return filepath.ToSlash(relPath), nil
}

func GetDeployedContracts(state *flowkit.State) flixkit.ContractInfos {
	allContracts := make(flixkit.ContractInfos)
	depNetworks := make([]string, 0)
	// get all configured networks in flow.json
	for _, n := range *state.Networks() {
		depNetworks = append(depNetworks, n.Name)
	}

	// get all deployed and alias contracts for configured networks
	for _, network := range depNetworks {
		contracts, err := state.DeploymentContractsByNetwork(config.Network{Name: network})
		if err != nil {
			continue
		}
		for _, c := range contracts {
			if _, ok := allContracts[c.Name]; !ok {
				allContracts[c.Name] = make(flixkit.NetworkAddressMap)
			}
			allContracts[c.Name][network] = c.AccountAddress.String()
		}
		locAliases := state.AliasesForNetwork(config.Network{Name: network})
		for name, addr := range locAliases {
			if isPath(name) {
				continue
			}
			if _, ok := allContracts[name]; !ok {
				allContracts[name] = make(flixkit.NetworkAddressMap)
			}
			allContracts[name][network] = addr
		}
	}
	return allContracts
}

func isUrl(str string) bool {
	u, err := url.Parse(str)
	return err == nil && u.Scheme != "" && u.Host != ""
}

func isPath(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
