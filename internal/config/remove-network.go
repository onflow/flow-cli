package config

import (
	"fmt"

	"github.com/onflow/flow-cli/internal/command"
	"github.com/onflow/flow-cli/pkg/flowcli/output"
	"github.com/onflow/flow-cli/pkg/flowcli/project"
	"github.com/onflow/flow-cli/pkg/flowcli/services"
	"github.com/spf13/cobra"
)

type flagsRemoveNetwork struct{}

var removeNetworkFlags = flagsRemoveNetwork{}

var RemoveNetworkCommand = &command.Command{
	Cmd: &cobra.Command{
		Use:     "network <name>",
		Short:   "Remove network from configuration",
		Example: "flow config remove network Foo",
		Args:    cobra.MaximumNArgs(1),
	},
	Flags: &removeNetworkFlags,
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

		name := ""
		if len(args) == 1 {
			name = args[0]
		} else {
			name = output.RemoveNetworkPrompt(p.Config().Networks)
		}

		err = p.Config().Networks.Remove(name)
		if err != nil {
			return nil, err
		}

		err = p.SaveDefault()
		if err != nil {
			return nil, err
		}

		return &ConfigResult{
			result: "network removed",
		}, nil
	},
}

func init() {
	RemoveNetworkCommand.AddToParent(RemoveCmd)
}
