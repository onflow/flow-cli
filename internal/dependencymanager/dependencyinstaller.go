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

package dependencymanager

import (
	"fmt"
	"github.com/onflow/flow-cli/internal/prompt"
	"github.com/onflow/flow-cli/internal/util"
	"github.com/onflow/flowkit/v2"
	"github.com/onflow/flowkit/v2/accounts"
	"github.com/onflow/flowkit/v2/deps"
	"github.com/psiemens/sconfig"
	"github.com/spf13/cobra"
)

func NewCliDependencyInstaller(state *flowkit.State, options ...deps.Option) (*deps.DependencyInstaller, error) {
	return deps.NewDependencyInstaller(state, cliPrompter{}, options...)
}

type Flags struct {
	skipDeployments bool `default:"false" flag:"skip-deployments" info:"Skip adding the dependency to deployments"`
	skipAlias       bool `default:"false" flag:"skip-alias" info:"Skip prompting for an alias"`
}

func (f *Flags) AddToCommand(cmd *cobra.Command) {
	err := sconfig.New(f).
		FromEnvironment(util.EnvPrefix).
		BindFlags(cmd.Flags()).
		Parse()

	if err != nil {
		panic(err)
	}
}

type cliPrompter struct{}

func (c cliPrompter) ShouldUpdateDependency(contractName string) bool {
	msg := fmt.Sprintf("The latest version of %s is different from the one you have locally. Do you want to update it?", contractName)
	return prompt.GenericBoolPrompt(msg)
}

func (c cliPrompter) AddContractToDeployment(networkName string, accounts accounts.Accounts, contractName string) *deps.DeploymentData {
	return prompt.AddContractToDeploymentPrompt(networkName, accounts, contractName)
}

func (c cliPrompter) AddressPromptOrEmpty(label, errorMessage string) string {
	return prompt.AddressPromptOrEmpty(label, errorMessage)
}

var _ deps.Prompter = cliPrompter{}
