package super

import (
	"fmt"

	"github.com/onflow/flow-cli/common/branding"
	"github.com/onflow/flow-cli/internal/dependencymanager"
	"github.com/onflow/flow-cli/internal/super/generator"
	"github.com/onflow/flowkit/v2"
	flowkitConfig "github.com/onflow/flowkit/v2/config"
	"github.com/onflow/flowkit/v2/output"
)

// ProjectType represents the type of Flow project to create
type ProjectType string

const (
	ProjectTypeDefault               ProjectType = "default"
	ProjectTypeScheduledTransactions ProjectType = "scheduledtransactions"
	ProjectTypeCustom                ProjectType = "custom"
)

// getProjectTypeDescription returns a user-friendly description for the project type
func getProjectTypeDescription(projectType ProjectType) string {
	switch projectType {
	case ProjectTypeDefault:
		return "Basic Cadence project (no dependencies)"
	case ProjectTypeScheduledTransactions:
		return "Scheduled Transactions project"
	case ProjectTypeCustom:
		return "Custom project (select standard Flow contract dependencies)"
	default:
		return string(projectType)
	}
}

// getProjectTemplates returns a slice of templates based on the project type.
// Supported types: ProjectTypeDefault, ProjectTypeScheduledTransactions
func getProjectTemplates(projectType ProjectType, targetDir string, state *flowkit.State) []generator.TemplateItem {
	switch projectType {
	case ProjectTypeScheduledTransactions:
		// Same as default for now - will be customized later
		return []generator.TemplateItem{
			generator.ContractTemplate{
				Name:         "Counter",
				TemplatePath: "contract_counter.cdc.tmpl",
			},
			generator.ContractTemplate{
				Name:         "CounterTransactionHandler",
				TemplatePath: "contract_counter_transaction_handler.cdc.tmpl",
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
						{"Name": "CounterTransactionHandler"},
					},
					"Scripts": []map[string]interface{}{
						{"Name": "GetCounter"},
					},
					"Transactions": []map[string]interface{}{
						{"Name": "IncrementCounter"},
						{"Name": "ScheduleIncrementCounter"},
					},
				},
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
func installProjectDependencies(logger output.Logger, state *flowkit.State, targetDir string, coreContracts []string, customDependencies []flowkitConfig.Dependency, deploymentAccount string) error {
	logger.Info("Installing project dependencies...")

	flags := dependencymanager.DependencyFlags{}
	installer, err := dependencymanager.NewDependencyInstaller(logger, state, false, targetDir, flags)
	if err != nil {
		return err
	}

	if deploymentAccount != "" {
		installer.DeploymentAccount = deploymentAccount
	}

	installer.SkipAlias = true

	// Install core contracts
	for _, coreContract := range coreContracts {
		contractName := branding.PurpleStyle.Render(coreContract)
		logger.Info(fmt.Sprintf("Installing core contract: %s", contractName))
		err = installer.AddByCoreContractName(coreContract)
		if err != nil {
			return err
		}
	}

	// Install custom dependencies
	if len(customDependencies) > 0 {
		for _, dep := range customDependencies {
			contractName := branding.PurpleStyle.Render(dep.Name)
			logger.Info(fmt.Sprintf("Installing dependency: %s", contractName))
		}
		err = installer.AddMany(customDependencies)
		if err != nil {
			return err
		}
	}

	logger.Info("Dependencies installed successfully!")
	return nil
}

// addContractDeployments adds specific contracts to the deployment configuration
func addContractDeployments(state *flowkit.State, contractNames []string, deploymentAccount string) error {
	// Find existing deployment for emulator network and account, or create new one
	deployment := state.Deployments().ByAccountAndNetwork(deploymentAccount, "emulator")
	if deployment == nil {
		// Create new deployment
		deployment = &flowkitConfig.Deployment{
			Network: "emulator",
			Account: deploymentAccount,
		}
		state.Deployments().AddOrUpdate(*deployment)
		deployment = state.Deployments().ByAccountAndNetwork(deploymentAccount, "emulator")
	}

	// Add contracts to deployment if not already present
	for _, contractName := range contractNames {
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
