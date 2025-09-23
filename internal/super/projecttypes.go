package super

import (
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
func installProjectDependencies(logger output.Logger, state *flowkit.State, targetDir string, coreContracts []string, customDependencies []flowkitConfig.Dependency) error {
	flags := dependencymanager.DependencyFlags{}
	installer, err := dependencymanager.NewDependencyInstaller(logger, state, false, targetDir, flags)
	if err != nil {
		return err
	}

	// Install core contracts
	for _, coreContract := range coreContracts {
		err = installer.AddByCoreContractName(coreContract)
		if err != nil {
			return err
		}
	}

	// Install custom dependencies
	if len(customDependencies) > 0 {
		err = installer.AddMany(customDependencies)
		if err != nil {
			return err
		}
	}

	return nil
}
