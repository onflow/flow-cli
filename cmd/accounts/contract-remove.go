package accounts

import (
	"github.com/onflow/flow-cli/cmd"
	"github.com/onflow/flow-cli/flow/cli"
	"github.com/onflow/flow-cli/sharedlib/services"
	"github.com/psiemens/sconfig"
	"github.com/spf13/cobra"
)

type flagsRemoveContract struct {
	Account string `default:"emulator-account" flag:"account,a"`
}

type cmdRemoveContract struct {
	cmd   *cobra.Command
	flags flagsRemoveContract
}

func NewRemoveContractCmd() cmd.Command {
	return &cmdRemoveContract{
		cmd: &cobra.Command{
			Use:   "remove-contract <name>",
			Short: "Remove a contract deployed to an account",
			Args:  cobra.ExactArgs(1),
		},
	}
}

func (c *cmdRemoveContract) Run(
	cmd *cobra.Command,
	args []string,
	project *cli.Project,
	services *services.Services,
) (cmd.Result, error) {

	account, err := services.Accounts.RemoveContract(args[0], c.flags.Account)
	return &AccountResult{
		Account:  account,
		showCode: false,
	}, err
}

func (c *cmdRemoveContract) GetFlags() *sconfig.Config {
	return sconfig.New(&c.flags)
}

func (c *cmdRemoveContract) GetCmd() *cobra.Command {
	return c.cmd
}
