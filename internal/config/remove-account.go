package config

import (
	"fmt"

	"github.com/onflow/flow-cli/internal/command"
	"github.com/onflow/flow-cli/pkg/flowcli/output"
	"github.com/onflow/flow-cli/pkg/flowcli/project"
	"github.com/onflow/flow-cli/pkg/flowcli/services"
	"github.com/spf13/cobra"
)

type flagsRemoveAccount struct{}

var removeAccountFlags = flagsRemoveAccount{}

var RemoveAccountCommand = &command.Command{
	Cmd: &cobra.Command{
		Use:     "account <name>",
		Short:   "Remove account from configuration",
		Example: "flow config remove account Foo",
		Args:    cobra.MaximumNArgs(1),
	},
	Flags: &removeAccountFlags,
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
		conf := p.Config()

		name := ""
		if len(args) == 1 {
			name = args[0]
		} else {
			name = output.RemoveAccountPrompt(conf.Accounts)
		}

		err = p.RemoveAccount(name)
		if err != nil {
			return nil, err
		}

		err = p.SaveDefault()
		if err != nil {
			return nil, err
		}

		return &ConfigResult{
			result: "account removed",
		}, nil
	},
}

func init() {
	RemoveAccountCommand.AddToParent(RemoveCmd)
}
