package config

import (
	"fmt"

	"github.com/onflow/flow-cli/internal/command"
	"github.com/onflow/flow-cli/pkg/flowcli/output"
	"github.com/onflow/flow-cli/pkg/flowcli/project"
	"github.com/onflow/flow-cli/pkg/flowcli/services"
	"github.com/spf13/cobra"
)

type flagsRemove struct{}

var removeFlags = flagsRemove{}

var RemoveCommand = &command.Command{
	Cmd: &cobra.Command{
		Use:       "remove <account|contract|deployment|network>",
		Short:     "Remove resource from configuration",
		Example:   "flow config remove account",
		ValidArgs: []string{"account", "contract", "deployment", "network"},
		Args:      cobra.ExactArgs(1),
	},
	Flags: &removeFlags,
	Run: func(
		cmd *cobra.Command,
		args []string,
		globalFlags command.GlobalFlags,
		services *services.Services,
	) (command.Result, error) {
		resource := args[0]

		p, err := project.Load(globalFlags.ConfigPath)
		if err != nil {
			return nil, fmt.Errorf("configuration does not exists")
		}
		conf := p.Config()

		switch resource {
		case "account":
			name := output.RemoveAccountPrompt(conf.Accounts)
			err := services.Config.RemoveAccount(name)
			if err != nil {
				return nil, err
			}

			return &ConfigResult{
				result: "account removed",
			}, nil

		case "deployment":
			accountName, networkName := output.RemoveDeploymentPrompt(conf.Deployments)
			err := services.Config.RemoveDeployment(accountName, networkName)
			if err != nil {
				return nil, err
			}

			return &ConfigResult{
				result: "deployment removed",
			}, nil

		case "contract":
			name := output.RemoveContractPrompt(conf.Contracts)
			err := services.Config.RemoveContract(name)
			if err != nil {
				return nil, err
			}

			return &ConfigResult{
				result: "contract removed",
			}, nil

		case "network":
			name := output.RemoveNetworkPrompt(conf.Networks)
			err := services.Config.RemoveNetwork(name)
			if err != nil {
				return nil, err
			}

			return &ConfigResult{
				result: "network removed",
			}, nil
		}

		return nil, nil
	},
}
