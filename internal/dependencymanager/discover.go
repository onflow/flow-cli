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

package dependencymanager

import (
	"fmt"
	"slices"

	flowsdk "github.com/onflow/flow-go-sdk"
	"github.com/onflow/flow-go/fvm/systemcontracts"

	"github.com/spf13/cobra"

	"github.com/onflow/flowkit/v2"
	"github.com/onflow/flowkit/v2/output"

	flowGo "github.com/onflow/flow-go/model/flow"
	flowkitConfig "github.com/onflow/flowkit/v2/config"

	"github.com/onflow/flow-cli/internal/command"
	"github.com/onflow/flow-cli/internal/prompt"
	"github.com/onflow/flow-cli/internal/util"
)

type DiscoverResult struct {
	Contracts []string `json:"contracts"`
}

var discoverCommand = &command.Command{
	Cmd: &cobra.Command{
		Use:     "discover",
		Short:   "Discover available contracts to add to your project.",
		Example: "flow dependencies discover",
		Args:    cobra.NoArgs,
	},
	RunS:  discover,
	Flags: &struct{}{},
}

func discover(
	_ []string,
	globalFlags command.GlobalFlags,
	logger output.Logger,
	flow flowkit.Services,
	state *flowkit.State,
) (command.Result, error) {
	installedDeps := state.Dependencies()
	if installedDeps == nil {
		installedDeps = new(flowkitConfig.Dependencies)
	}

	installedContracts := make([]string, 0)
	for _, dep := range *installedDeps {
		installedContracts = append(installedContracts, dep.Name)
	}

	err := PromptInstallCoreContracts(logger, state, "", installedContracts)
	if err != nil {
		return nil, err
	}

	err = state.SaveDefault()
	return nil, err
}

func PromptInstallCoreContracts(logger output.Logger, state *flowkit.State, targetDir string, excludeContracts []string) error {
	// Prompt to ask which core contracts should be installed
	sc := systemcontracts.SystemContractsForChain(flowGo.Mainnet)
	promptMessage := "Select any core contracts you would like to install or skip to continue."

	contractNames := make([]string, 0)

	for _, contract := range sc.All() {
		if slices.Contains(excludeContracts, contract.Name) {
			continue
		}
		contractNames = append(contractNames, contract.Name)
	}

	var footer string
	totalContracts := len(sc.All())
	availableContracts := len(contractNames)
	installedCount := totalContracts - availableContracts

	if installedCount > 0 {
		footer = fmt.Sprintf("ℹ️  Note: %d core contracts already installed. Use 'flow deps list' to view them.", installedCount)
	}

	selectedContractNames, err := prompt.RunSelectOptionsWithFooter(contractNames, promptMessage, footer)
	if err != nil {
		return fmt.Errorf("error running dependency selection: %v\n", err)
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

	logger.Info("")
	logger.Info(util.MessageWithEmojiPrefix("🔄", "Installing selected core contracts and dependencies..."))

	// Add the selected core contracts as dependencies
	installer, err := NewDependencyInstaller(logger, state, false, targetDir, Flags{})
	if err != nil {
		return err
	}

	if err := installer.AddMany(dependencies); err != nil {
		return err
	}

	return nil
}
