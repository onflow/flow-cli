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

	"github.com/spf13/cobra"

	"github.com/onflow/flowkit/v2"
	"github.com/onflow/flowkit/v2/output"

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
	return PromptInstallContracts(logger, state, targetDir, excludeContracts)
}

func PromptInstallContracts(logger output.Logger, state *flowkit.State, targetDir string, excludeContracts []string) error {
	sections := GetAllContractSections()
	sectionMap := make(map[string][]string)
	allDependenciesByName := make(map[string]flowkitConfig.Dependency)

	var totalAvailable, totalInstalled int

	for _, section := range sections {
		var availableContracts []string
		sectionInstalled := 0

		for _, dep := range section.Dependencies {
			if dep.Source.NetworkName != flowkitConfig.MainnetNetwork.Name {
				continue
			}

			if slices.Contains(excludeContracts, dep.Name) {
				sectionInstalled++
				continue
			}

			availableContracts = append(availableContracts, dep.Name)
			allDependenciesByName[dep.Name] = dep
		}

		totalAvailable += len(availableContracts)
		totalInstalled += sectionInstalled

		if len(availableContracts) > 0 {
			sectionMap[section.Name] = availableContracts
		}
	}

	var footer string
	if totalInstalled > 0 {
		footer = fmt.Sprintf("ℹ️  Note: %d contracts already installed. Use 'flow dependencies list' to view them.", totalInstalled)
	}

	promptMessage := "Select any contracts you would like to install"
	selectedContractNames, err := prompt.RunContractList(sectionMap, promptMessage, footer)
	if err != nil {
		return fmt.Errorf("error running dependency selection: %v\n", err)
	}

	var dependencies []flowkitConfig.Dependency
	for _, contractName := range selectedContractNames {
		if dep, exists := allDependenciesByName[contractName]; exists {
			dependencies = append(dependencies, dep)
		}
	}

	if len(dependencies) == 0 {
		return nil
	}

	logger.Info("")
	logger.Info(util.MessageWithEmojiPrefix("🔄", "Installing selected contracts and dependencies..."))

	// Add the selected contracts as dependencies
	installer, err := NewDependencyInstaller(logger, state, false, targetDir, Flags{})
	if err != nil {
		return err
	}

	if err := installer.AddMany(dependencies); err != nil {
		return err
	}

	return nil
}
