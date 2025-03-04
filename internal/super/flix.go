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
	"context"
	"fmt"
	"os"

	"github.com/onflow/cadence/parser"
	"github.com/onflow/flixkit-go/v2/flixkit"

	"github.com/spf13/cobra"

	"github.com/onflow/flowkit/v2"
	"github.com/onflow/flowkit/v2/config"
	"github.com/onflow/flowkit/v2/output"

	"golang.org/x/exp/slices"

	"github.com/onflow/flow-cli/internal/command"
	"github.com/onflow/flow-cli/internal/scripts"
	"github.com/onflow/flow-cli/internal/transactions"
	"github.com/onflow/flow-cli/internal/util"
)

type flixFlags struct {
	ArgsJSON        string   `default:"" flag:"args-json" info:"arguments in JSON-Cadence format"`
	BlockID         string   `default:"" flag:"block-id" info:"block ID to execute the script at"`
	BlockHeight     uint64   `default:"" flag:"block-height" info:"block height to execute the script at"`
	Signer          string   `default:"" flag:"signer" info:"Account name from configuration used to sign the transaction as proposer, payer and suthorizer"`
	Proposer        string   `default:"" flag:"proposer" info:"Account name from configuration used as proposer"`
	Payer           string   `default:"" flag:"payer" info:"Account name from configuration used as payer"`
	Authorizers     []string `default:"" flag:"authorizer" info:"Name of a single or multiple comma-separated accounts used as authorizers from configuration"`
	Include         []string `default:"" flag:"include" info:"Fields to include in the output"`
	Exclude         []string `default:"" flag:"exclude" info:"Fields to exclude from the output (events)"`
	GasLimit        uint64   `default:"1000" flag:"gas-limit" info:"transaction gas limit"`
	PreFill         string   `default:"" flag:"pre-fill" info:"template path to pre fill the FLIX"`
	Lang            string   `default:"js" flag:"lang" info:"language to generate the template for"`
	ExcludeNetworks []string `default:"" flag:"exclude-networks" info:"Specify which networks to exclude when generating a FLIX template"`
}

type flixResult struct {
	flixQuery string
	result    string
}

var (
	flags   = flixFlags{}
	FlixCmd = &cobra.Command{
		Use:              "flix",
		Short:            "Commands to execute, generate, package FLIX templates",
		TraverseChildren: true,
		GroupID:          "tools",
		Example:          "flow flix execute transfer-flow 1 0x123456789",
	}
)

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
	flags command.GlobalFlags,
	logger output.Logger,
	flow flowkit.Services,
	state *flowkit.State,
) (result command.Result, err error) {
	flixService := flixkit.NewFlixService(&flixkit.FlixServiceConfig{
		FileReader: state,
	})
	return executeFlixCmd(args, flags, logger, flow, state, flixService)
}

func executeFlixCmd(
	args []string,
	_ command.GlobalFlags,
	logger output.Logger,
	flow flowkit.Services,
	state *flowkit.State,
	flixService flixkit.FlixService,
) (result command.Result, err error) {
	flixQuery := args[0]
	ctx := context.Background()
	cadenceWithImportsReplaced, err := flixService.GetTemplateAndReplaceImports(ctx, flixQuery, flow.Network().Name)
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
	// some reason sendTransaction clips the first argument
	return transactions.SendTransaction([]byte(cadenceWithImportsReplaced.Cadence), args, "", flow, state, transactionFlags)
}

func packageCmd(
	args []string,
	gFlags command.GlobalFlags,
	logger output.Logger,
	flow flowkit.Services,
	state *flowkit.State,
) (result command.Result, err error) {
	flixService := flixkit.NewFlixService(&flixkit.FlixServiceConfig{
		FileReader: state,
	})

	return packageFlixCmd(args, gFlags, logger, flow, state, flixService, flags)
}

func packageFlixCmd(
	args []string,
	gFlags command.GlobalFlags,
	_ output.Logger,
	_ flowkit.Services,
	_ *flowkit.State,
	flixService flixkit.FlixService,
	flags flixFlags,
) (result command.Result, err error) {
	flixQuery := args[0]
	ctx := context.Background()
	out, err := flixService.GetTemplateAndCreateBinding(ctx, flixQuery, flags.Lang, gFlags.Save)
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
	gFlags command.GlobalFlags,
	logger output.Logger,
	flow flowkit.Services,
	state *flowkit.State,
) (result command.Result, err error) {
	flixService := flixkit.NewFlixService(&flixkit.FlixServiceConfig{
		FileReader: state,
		Logger:     logger,
	})

	return generateFlixCmd(args, gFlags, logger, flow, state, flixService, flags)
}

func generateFlixCmd(
	args []string,
	_ command.GlobalFlags,
	_ output.Logger,
	_ flowkit.Services,
	state *flowkit.State,
	flixService flixkit.FlixService,
	flags flixFlags,
) (result command.Result, err error) {
	cadenceFile := args[0]
	depContracts := getContractsFromState(state, flags.ExcludeNetworks)
	if err != nil {
		return nil, fmt.Errorf("could not get core contracts %w", err)
	}

	if cadenceFile == "" {
		return nil, fmt.Errorf("no cadence code found")
	}

	code, err := state.ReadFile(cadenceFile)
	if err != nil {
		return nil, fmt.Errorf("could not read cadence file %s: %w", cadenceFile, err)
	}

	// get user's configured networks
	depNetworks := getNetworks(state)

	if len(flags.ExcludeNetworks) > 0 {
		excludeMap := make(map[string]bool)
		for _, net := range flags.ExcludeNetworks {
			excludeMap[net] = true
		}

		var filteredNetworks []flixkit.NetworkConfig
		for _, network := range depNetworks {
			if !excludeMap[network.Name] {
				filteredNetworks = append(filteredNetworks, network)
			}
		}

		depNetworks = filteredNetworks
		if len(depNetworks) == 0 {
			return nil, fmt.Errorf("all networks have been excluded")
		}
	}

	err = validateImports(string(code), depContracts, depNetworks)
	if err != nil {
		return nil, fmt.Errorf("could not validate imported contracts: %w", err)
	}

	ctx := context.Background()
	prettyJSON, err := flixService.CreateTemplate(ctx, depContracts, string(code), flags.PreFill, depNetworks)
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

func validateImports(code string, depContracts flixkit.ContractInfos, depNetworks []flixkit.NetworkConfig) error {
	// Check all imported contracts in the cadence code
	astProgram, err := parser.ParseProgram(nil, []byte(code), parser.Config{})
	if err != nil {
		return fmt.Errorf("could not parse Cadence code %w", err)
	}

	// Check for any missing string imports
	for _, imp := range astProgram.ImportDeclarations() {
		if len(imp.Identifiers) > 0 || imp.Location == nil {
			return fmt.Errorf("only string imports of the form `import \"ContractName\"` are supported")
		}

		contractName := imp.Location.String()

		if depContracts[contractName] == nil {
			if util.IsCoreContract(contractName) {
				return fmt.Errorf("contract %[1]s is not found in the flow.json configuration, if this refers to the %[1]s core contract, please add it using `flow deps install %[1]s`", contractName)
			}

			return fmt.Errorf("contract %[1]s is not found in the flow.json configuration, if it refers to an external contract, please add it using `flow deps install <network>://<address>.%[1]s`", contractName)
		}

		for _, network := range depNetworks {
			if depContracts[contractName][network.Name] == "" {
				return fmt.Errorf("contract %s was found in the flow.json configuration, but is missing an alias for network %s", contractName, network.Name)
			}
		}
	}

	return nil
}

func getContractsFromState(state *flowkit.State, excludeNetworks []string) flixkit.ContractInfos {
	allContracts := make(flixkit.ContractInfos)
	depNetworks := make([]string, 0)
	accountAddresses := make(map[string]string)
	// get all configured networks in flow.json
	for _, n := range *state.Networks() {
		depNetworks = append(depNetworks, n.Name)
	}

	// get account addresses
	for _, a := range *state.Accounts() {
		accountAddresses[a.Name] = a.Address.HexWithPrefix()
	}

	for _, d := range *state.Deployments() {
		addr := accountAddresses[d.Account]
		for _, c := range d.Contracts {
			if _, ok := allContracts[c.Name]; !ok {
				allContracts[c.Name] = make(flixkit.NetworkAddressMap)
			}
			if slices.Contains(excludeNetworks, d.Network) {
				continue
			}
			allContracts[c.Name][d.Network] = addr
		}
	}

	// get all deployed and alias contracts for configured networks
	for _, network := range depNetworks {
		cfg := config.Network{Name: network}
		contracts, err := state.DeploymentContractsByNetwork(cfg)
		if err != nil {
			continue
		}
		for _, c := range contracts {
			if _, ok := allContracts[c.Name]; !ok {
				allContracts[c.Name] = make(flixkit.NetworkAddressMap)
			}
			if slices.Contains(excludeNetworks, network) {
				continue
			}
			allContracts[c.Name][network] = c.AccountAddress.HexWithPrefix()
		}
		locAliases := state.AliasesForNetwork(cfg)
		for name, addr := range locAliases {
			if isPath(name) {
				continue
			}
			if _, ok := allContracts[name]; !ok {
				allContracts[name] = make(flixkit.NetworkAddressMap)
			}
			if slices.Contains(excludeNetworks, network) {
				continue
			}
			allContracts[name][network] = addr
		}
	}

	// add contracts that have not been deployed
	for _, c := range *state.Contracts() {
		if _, ok := allContracts[c.Name]; !ok {
			allContracts[c.Name] = make(flixkit.NetworkAddressMap)
		}
		for _, alias := range c.Aliases {
			if slices.Contains(excludeNetworks, alias.Network) {
				continue
			}
			allContracts[c.Name][alias.Network] = alias.Address.HexWithPrefix()
		}
	}

	return allContracts
}

func getNetworks(state *flowkit.State) []flixkit.NetworkConfig {
	networks := make([]flixkit.NetworkConfig, 0)
	for _, n := range *state.Networks() {
		networks = append(networks, flixkit.NetworkConfig{
			Name: n.Name,
			Host: n.Host,
			Key:  n.Key,
		})
	}
	return networks
}

func isPath(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
