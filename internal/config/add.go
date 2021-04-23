package config

import (
	"fmt"

	"github.com/onflow/flow-cli/internal/command"
	"github.com/onflow/flow-cli/pkg/flowcli/config"
	"github.com/onflow/flow-cli/pkg/flowcli/output"
	"github.com/onflow/flow-cli/pkg/flowcli/project"
	"github.com/onflow/flow-cli/pkg/flowcli/services"
	"github.com/spf13/cobra"
)

type flagsAdd struct{}

var addFlags = flagsAdd{}

var AddCommand = &command.Command{
	Cmd: &cobra.Command{
		Use:       "add account",
		Short:     "Add resource to configuration",
		Example:   "flow config add account",
		ValidArgs: []string{"account", "contract", "deployment", "network"},
		Args:      cobra.ExactArgs(1),
	},
	Flags: &addFlags,
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
			accountData := output.NewAccountPrompt()
			account, err := config.StringToAccount(
				accountData["name"],
				accountData["address"],
				accountData["keyIndex"],
				accountData["sigAlgo"],
				accountData["hashAlgo"],
				accountData["key"],
			)
			if err != nil {
				return nil, err
			}

			err = services.Config.AddAccount(*account)
			if err != nil {
				return nil, err
			}

			return &ConfigResult{
				result: "account added",
			}, nil

		case "contract":
			contractData := output.NewContractPrompt()
			contracts := config.StringToContracts(
				contractData["name"],
				contractData["source"],
				contractData["emulator"],
				contractData["testnet"],
			)

			err := services.Config.AddContracts(contracts)
			if err != nil {
				return nil, err
			}

			return &ConfigResult{
				result: "contract added",
			}, nil

		case "deployment":
			deployData := output.NewDeploymentPrompt(conf.Networks, conf.Accounts, conf.Contracts)
			deployment := config.StringToDeployment(
				deployData["network"].(string),
				deployData["account"].(string),
				deployData["contracts"].([]string),
			)
			err := services.Config.AddDeployment(deployment)
			if err != nil {
				return nil, err
			}

			return &ConfigResult{
				result: "deploy added",
			}, nil

		case "network":
			networkData := output.NewNetworkPrompt()
			network := config.StringToNetwork(networkData["name"], networkData["host"])

			err := services.Config.AddNetwork(network)
			if err != nil {
				return nil, err
			}

			return &ConfigResult{
				result: "network added",
			}, nil
		}

		return nil, nil
	},
}
