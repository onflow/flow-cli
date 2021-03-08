package accounts

import (
	"github.com/onflow/flow-cli/cmd"
	"github.com/onflow/flow-cli/flow/cli"
	"github.com/onflow/flow-cli/sharedlib/services"
	"github.com/psiemens/sconfig"
	"github.com/spf13/cobra"
)

type flagsGet struct {
	Host string `flag:"host" info:"Flow Access API host address"`
	Code bool   `default:"false" flag:"code" info:"Display code deployed to the account"`
}

type cmdGet struct {
	cmd   *cobra.Command
	flags flagsGet
}

func NewGetCmd() cmd.Command {
	return &cmdGet{
		cmd: &cobra.Command{
			Use:     "get <address>",
			Short:   "Gets an account by address",
			Aliases: []string{"fetch", "g"},
			Long:    `Gets an account by address (address, balance, keys, code)`,
			Args:    cobra.ExactArgs(1),
		},
	}
}

func (a *cmdGet) Run(cmd *cobra.Command, args []string, project *cli.Project, services *services.Services) (cmd.Result, error) {
	account, err := services.Accounts.Get(args[0])
	return &AccountResult{account}, err
}

func (a *cmdGet) GetFlags() *sconfig.Config {
	return sconfig.New(&a.flags)
}

func (a *cmdGet) GetCmd() *cobra.Command {
	return a.cmd
}
