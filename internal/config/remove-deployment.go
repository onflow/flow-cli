package config

import (
	"fmt"

	"github.com/onflow/flow-cli/internal/command"
	"github.com/onflow/flow-cli/pkg/flowcli/output"
	"github.com/onflow/flow-cli/pkg/flowcli/project"
	"github.com/onflow/flow-cli/pkg/flowcli/services"
	"github.com/spf13/cobra"
)

type flagsRemoveDeployment struct{}

var removeDeploymentFlags = flagsRemoveDeployment{}

var RemoveDeploymentCommand = &command.Command{
	Cmd: &cobra.Command{
		Use:     "deployment <account> <network>",
		Short:   "Remove deployment from configuration",
		Example: "flow config remove deployment Foo testnet",
		Args:    cobra.MaximumNArgs(2),
	},
	Flags: &removeDeploymentFlags,
	Run: func(
		cmd *cobra.Command,
		args []string,
		globalFlags command.GlobalFlags,
		services *services.Services,
	) (command.Result, error) {
		p, err := project.Load(globalFlags.ConfigPath)
		if err != nil {
			return nil, fmt.Errorf("configuration does not exists")
		}

		account := ""
		network := ""
		if len(args) == 2 {
			account = args[0]
			network = args[1]
		} else {
			account, network = output.RemoveDeploymentPrompt(p.Config().Deployments)
		}

		err = p.Config().Deployments.Remove(account, network)
		if err != nil {
			return nil, err
		}

		err = p.SaveDefault()
		if err != nil {
			return nil, err
		}

		return &ConfigResult{
			result: "deployment removed",
		}, nil
	},
}

func init() {
	RemoveDeploymentCommand.AddToParent(RemoveCmd)
}
