package accounts

import (
	"github.com/onflow/flow-cli/cmd"
	"github.com/onflow/flow-cli/flow/lib"
	"github.com/onflow/flow-cli/flow/services"
	"github.com/psiemens/sconfig"
	"github.com/spf13/cobra"
)

type flagsGet struct {
	Code bool `default:"false" flag:"code" info:"Display code deployed to the account"`
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

func (a *cmdGet) Run(
	cmd *cobra.Command,
	args []string,
	project *lib.Project,
	services *services.Services,
) (cmd.Result, error) {

	account, err := services.Accounts.Get(args[0])
	return &AccountResult{
		Account:  account,
		showCode: a.flags.Code,
	}, err
}

func (a *cmdGet) GetFlags() *sconfig.Config {
	return sconfig.New(&a.flags)
}

func (a *cmdGet) GetCmd() *cobra.Command {
	return a.cmd
}
