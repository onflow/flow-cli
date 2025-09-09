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
	"strings"

	"github.com/spf13/cobra"

	"github.com/onflow/flowkit/v2"
	"github.com/onflow/flowkit/v2/output"

	"github.com/onflow/flow-cli/internal/command"
)

type ListResult struct {
	Dependencies []DependencyInfo `json:"dependencies"`
}

type DependencyInfo struct {
	Name        string `json:"name"`
	NetworkName string `json:"network"`
	Address     string `json:"address"`
	Contract    string `json:"contract"`
}

var listCommand = &command.Command{
	Cmd: &cobra.Command{
		Use:     "list",
		Short:   "List installed dependencies",
		Example: "flow dependencies list",
		Args:    cobra.NoArgs,
	},
	RunS:  list,
	Flags: &struct{}{},
}

func list(
	_ []string,
	globalFlags command.GlobalFlags,
	logger output.Logger,
	flow flowkit.Services,
	state *flowkit.State,
) (command.Result, error) {
	installedDeps := state.Dependencies()
	if installedDeps == nil || len(*installedDeps) == 0 {
		return &ListResult{Dependencies: []DependencyInfo{}}, nil
	}

	var dependencies []DependencyInfo
	for _, dep := range *installedDeps {
		dependencies = append(dependencies, DependencyInfo{
			Name:        dep.Name,
			NetworkName: dep.Source.NetworkName,
			Address:     dep.Source.Address.String(),
			Contract:    dep.Source.ContractName,
		})
	}

	return &ListResult{Dependencies: dependencies}, nil
}

func (r *ListResult) String() string {
	if len(r.Dependencies) == 0 {
		return "No dependencies installed"
	}

	var result strings.Builder
	result.WriteString(fmt.Sprintf("Installed dependencies (%d):\n\n", len(r.Dependencies)))

	// Find max widths for alignment
	maxNameWidth := 4    // "NAME"
	maxNetworkWidth := 7 // "NETWORK"
	maxAddressWidth := 7 // "ADDRESS"

	for _, dep := range r.Dependencies {
		if len(dep.Name) > maxNameWidth {
			maxNameWidth = len(dep.Name)
		}
		if len(dep.NetworkName) > maxNetworkWidth {
			maxNetworkWidth = len(dep.NetworkName)
		}
		if len(dep.Address) > maxAddressWidth {
			maxAddressWidth = len(dep.Address)
		}
	}

	result.WriteString(fmt.Sprintf("%-*s  %-*s  %-*s  %s\n",
		maxNameWidth, "NAME",
		maxNetworkWidth, "NETWORK",
		maxAddressWidth, "ADDRESS",
		"CONTRACT"))
	result.WriteString(strings.Repeat("-", maxNameWidth+maxNetworkWidth+maxAddressWidth+20) + "\n")

	for _, dep := range r.Dependencies {
		result.WriteString(fmt.Sprintf("%-*s  %-*s  %-*s  %s\n",
			maxNameWidth, dep.Name,
			maxNetworkWidth, dep.NetworkName,
			maxAddressWidth, dep.Address,
			dep.Contract))
	}

	return result.String()
}

func (r *ListResult) Oneliner() string {
	return fmt.Sprintf("Found %d installed dependencies", len(r.Dependencies))
}

func (r *ListResult) JSON() any {
	return r
}
