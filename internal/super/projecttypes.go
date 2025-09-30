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
	"fmt"

	flowsdk "github.com/onflow/flow-go-sdk"
	"github.com/onflow/flowkit/v2"
	flowkitConfig "github.com/onflow/flowkit/v2/config"
	"github.com/onflow/flowkit/v2/output"

	"github.com/onflow/flow-cli/internal/dependencymanager"
	"github.com/onflow/flow-cli/internal/super/generator"
)

// ProjectType represents the type of Flow project to create
type ProjectType string

const (
	ProjectTypeDefault               ProjectType = "default"
	ProjectTypeScheduledTransactions ProjectType = "scheduledtransactions"
	ProjectTypeCustom                ProjectType = "custom"
)

// ProjectTypeConfig holds configuration for a specific project type
type ProjectTypeConfig struct {
	Description        string
	CoreContracts      []string
	CustomDependencies []flowkitConfig.Dependency
	ContractNames      []string // For deployments
	DeploymentAccount  string   // Default deployment account
}

// getProjectTypeConfigs returns a map of all project type configurations
func getProjectTypeConfigs() map[ProjectType]*ProjectTypeConfig {
	return map[ProjectType]*ProjectTypeConfig{
		ProjectTypeDefault: {
			Description:        "Basic Cadence project (no dependencies)",
			CoreContracts:      []string{},
			CustomDependencies: []flowkitConfig.Dependency{},
			ContractNames:      []string{"Counter"},
			DeploymentAccount:  "emulator-account",
		},
		ProjectTypeScheduledTransactions: {
			Description:   "Scheduled Transactions project",
			CoreContracts: []string{}, // TODO: Add FlowTransactionScheduler as core contract once available
			CustomDependencies: []flowkitConfig.Dependency{
				{
					Name: "FlowTransactionScheduler",
					Source: flowkitConfig.Source{
						NetworkName:  flowkitConfig.TestnetNetwork.Name,
						Address:      flowsdk.HexToAddress("8c5303eaa26202d6"),
						ContractName: "FlowTransactionScheduler",
					},
					Aliases: flowkitConfig.Aliases{
						{
							Network: "emulator",
							Address: flowsdk.HexToAddress("f8d6e0586b0a20c7"),
						},
						{
							Network: "testnet",
							Address: flowsdk.HexToAddress("8c5303eaa26202d6"),
						},
						{
							Network: "mainnet",
							Address: flowsdk.HexToAddress("e467b9dd11fa00df"),
						},
					},
				},
				{
					Name: "FlowTransactionSchedulerUtils",
					Source: flowkitConfig.Source{
						NetworkName:  flowkitConfig.TestnetNetwork.Name,
						Address:      flowsdk.HexToAddress("8c5303eaa26202d6"),
						ContractName: "FlowTransactionSchedulerUtils",
					},
					Aliases: flowkitConfig.Aliases{
						{
							Network: "emulator",
							Address: flowsdk.HexToAddress("f8d6e0586b0a20c7"),
						},
						{
							Network: "testnet",
							Address: flowsdk.HexToAddress("8c5303eaa26202d6"),
						},
						{
							Network: "mainnet",
							Address: flowsdk.HexToAddress("e467b9dd11fa00df"),
						},
					},
				},
			},
			ContractNames:     []string{"Counter", "CounterTransactionHandler"},
			DeploymentAccount: "emulator-account",
		},
		ProjectTypeCustom: {
			Description:        "Custom project (select standard Flow contract dependencies)",
			CoreContracts:      []string{},
			CustomDependencies: []flowkitConfig.Dependency{},
			ContractNames:      []string{"Counter"},
			DeploymentAccount:  "emulator-account",
		},
	}
}

// getProjectTypeConfig returns the configuration for a given project type
func getProjectTypeConfig(projectType ProjectType) *ProjectTypeConfig {
	configs := getProjectTypeConfigs()
	if config, exists := configs[projectType]; exists {
		return config
	}
	// Return default configuration if not found
	return configs[ProjectTypeDefault]
}

// getProjectTemplates returns a slice of templates based on the project type.
// Supported types: ProjectTypeDefault, ProjectTypeScheduledTransactions
func getProjectTemplates(projectType ProjectType, targetDir string, state *flowkit.State) []generator.TemplateItem {
	switch projectType {
	case ProjectTypeScheduledTransactions:
		return []generator.TemplateItem{
			generator.ContractTemplate{
				Name:         "Counter",
				TemplatePath: "contract_counter.cdc.tmpl",
			},
			generator.ContractTemplate{
				Name:         "CounterTransactionHandler",
				TemplatePath: "contract_counter_transaction_handler.cdc.tmpl",
				SkipTests:    true,
			},
			generator.ScriptTemplate{
				Name:         "GetCounter",
				TemplatePath: "script_counter.cdc.tmpl",
				Data:         map[string]interface{}{"ContractName": "Counter"},
			},
			generator.TransactionTemplate{
				Name:         "IncrementCounter",
				TemplatePath: "transaction_counter.cdc.tmpl",
				Data:         map[string]interface{}{"ContractName": "Counter"},
			},
			generator.TransactionTemplate{
				Name:         "ScheduleIncrementCounter",
				TemplatePath: "transaction_schedule_increment_counter.cdc.tmpl",
			},
			generator.TransactionTemplate{
				Name:         "InitSchedulerManager",
				TemplatePath: "transaction_init_schedule_manager.cdc.tmpl",
			},
			generator.TransactionTemplate{
				Name:         "InitCounterTransactionHandler",
				TemplatePath: "transaction_init_counter_transaction_handler.cdc.tmpl",
			},
			generator.FileTemplate{
				TemplatePath: "README_scheduled_transactions.md.tmpl",
				TargetPath:   getReadmeFileName(targetDir),
				Data: map[string]interface{}{
					"Dependencies": (func() []map[string]interface{} {
						contracts := []map[string]interface{}{}
						for _, dep := range *state.Dependencies() {
							contracts = append(contracts, map[string]interface{}{
								"Name": dep.Name,
							})
						}
						return contracts
					})(),
					"Contracts": []map[string]interface{}{
						{"Name": "Counter"},
						{"Name": "CounterTransactionHandler"},
					},
					"Scripts": []map[string]interface{}{
						{"Name": "GetCounter"},
					},
					"Transactions": []map[string]interface{}{
						{"Name": "IncrementCounter"},
						{"Name": "ScheduleIncrementCounter"},
						{"Name": "InitSchedulerManager"},
						{"Name": "InitCounterTransactionHandler"},
					},
				},
			},
			generator.FileTemplate{
				TemplatePath: "cursor/agent_rules.mdc.tmpl",
				TargetPath:   ".cursor/rules/scheduledtransactions/agent-rules.mdc",
				Data:         map[string]interface{}{},
			},
			generator.FileTemplate{
				TemplatePath: "cursor/flip.md.tmpl",
				TargetPath:   ".cursor/rules/scheduledtransactions/flip.md",
				Data:         map[string]interface{}{},
			},
			generator.FileTemplate{
				TemplatePath: "cursor/index.md.tmpl",
				TargetPath:   ".cursor/rules/scheduledtransactions/index.md",
				Data:         map[string]interface{}{},
			},
			generator.FileTemplate{
				TemplatePath: "cursor/quick_checklist.md.tmpl",
				TargetPath:   ".cursor/rules/scheduledtransactions/quick-checklist.md",
				Data:         map[string]interface{}{},
			},
		}
	default:
		// Return default templates if unknown project type
		return []generator.TemplateItem{
			generator.ContractTemplate{
				Name:         "Counter",
				TemplatePath: "contract_counter.cdc.tmpl",
			},
			generator.ScriptTemplate{
				Name:         "GetCounter",
				TemplatePath: "script_counter.cdc.tmpl",
				Data:         map[string]interface{}{"ContractName": "Counter"},
			},
			generator.TransactionTemplate{
				Name:         "IncrementCounter",
				TemplatePath: "transaction_counter.cdc.tmpl",
				Data:         map[string]interface{}{"ContractName": "Counter"},
			},
			generator.FileTemplate{
				TemplatePath: "README.md.tmpl",
				TargetPath:   getReadmeFileName(targetDir),
				Data: map[string]interface{}{
					"Dependencies": (func() []map[string]interface{} {
						contracts := []map[string]interface{}{}
						for _, dep := range *state.Dependencies() {
							contracts = append(contracts, map[string]interface{}{
								"Name": dep.Name,
							})
						}
						return contracts
					})(),
					"Contracts": []map[string]interface{}{
						{"Name": "Counter"},
					},
					"Scripts": []map[string]interface{}{
						{"Name": "GetCounter"},
					},
					"Transactions": []map[string]interface{}{
						{"Name": "IncrementCounter"},
					},
				},
			},
		}
	}
}

// installProjectDependencies installs both core contracts and custom dependencies for a project type
func installProjectDependencies(logger output.Logger, state *flowkit.State, targetDir string, projectType ProjectType) error {
	config := getProjectTypeConfig(projectType)
	logger.Info("\nInstalling project dependencies...")

	flags := dependencymanager.DependencyFlags{}
	installer, err := dependencymanager.NewDependencyInstaller(logger, state, false, targetDir, flags)
	if err != nil {
		return err
	}

	installer.SkipAlias = true
	installer.SkipDeployments = true

	// Install core contracts
	for _, coreContract := range config.CoreContracts {
		err = installer.AddByCoreContractName(coreContract)
		if err != nil {
			return err
		}
	}

	// Install custom dependencies
	if len(config.CustomDependencies) > 0 {
		err = installer.AddMany(config.CustomDependencies)
		if err != nil {
			return err
		}
	}

	// Show installation summary
	count := installer.GetInstallCount()
	if count > 0 {
		logger.Info(fmt.Sprintf("\n✅ Successfully installed %d dependencies!\n", count))
	} else {
		logger.Info("No dependencies to install.\n")
	}
	return nil
}

// addContractDeployments adds specific contracts to the deployment configuration
func addContractDeployments(state *flowkit.State, projectType ProjectType) error {
	config := getProjectTypeConfig(projectType)
	// Find existing deployment for emulator network and account, or create new one
	deployment := state.Deployments().ByAccountAndNetwork(config.DeploymentAccount, "emulator")
	if deployment == nil {
		// Create new deployment
		deployment = &flowkitConfig.Deployment{
			Network: "emulator",
			Account: config.DeploymentAccount,
		}
		state.Deployments().AddOrUpdate(*deployment)
		deployment = state.Deployments().ByAccountAndNetwork(config.DeploymentAccount, "emulator")
	}

	// Add contracts to deployment if not already present
	for _, contractName := range config.ContractNames {
		found := false
		for _, existingContract := range deployment.Contracts {
			if existingContract.Name == contractName {
				found = true
				break
			}
		}
		if !found {
			deployment.Contracts = append(deployment.Contracts, flowkitConfig.ContractDeployment{
				Name: contractName,
			})
		}
	}

	state.Deployments().AddOrUpdate(*deployment)
	return nil
}
