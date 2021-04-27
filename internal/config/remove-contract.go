package config

import (
	"fmt"

	"github.com/onflow/flow-cli/internal/command"
	"github.com/onflow/flow-cli/pkg/flowcli/output"
	"github.com/onflow/flow-cli/pkg/flowcli/project"
	"github.com/onflow/flow-cli/pkg/flowcli/services"
	"github.com/spf13/cobra"
)

type flagsRemoveContract struct{}

var removeContractFlags = flagsRemoveContract{}

var RemoveContractCommand = &command.Command{
	Cmd: &cobra.Command{
		Use:     "contract <name>",
		Short:   "Remove contract from configuration",
		Example: "flow config remove contract Foo",
		Args:    cobra.MaximumNArgs(1),
	},
	Flags: &removeContractFlags,
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
			name = output.RemoveContractPrompt(p.Config().Contracts)
		}

		err = p.Config().Contracts.Remove(name)
		if err != nil {
			return nil, err
		}

		err = p.SaveDefault()
		if err != nil {
			return nil, err
		}

		return &ConfigResult{
			result: "contract removed",
		}, nil
	},
}

func init() {
	RemoveContractCommand.AddToParent(RemoveCmd)
}
