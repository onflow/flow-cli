package accounts

import (
	"github.com/onflow/flow-cli/cmd"
	"github.com/onflow/flow-cli/flow/cli"
	"github.com/onflow/flow-cli/sharedlib/services"
	"github.com/psiemens/sconfig"
	"github.com/spf13/cobra"
)

type flagsUpdateContract struct {
	Account string `default:"emulator-account" flag:"account,a"`
	Host    string `flag:"host" info:"Flow Access API host address"`
}

type cmdUpdateContract struct {
	cmd   *cobra.Command
	flags flagsUpdateContract
}

func NewUpdateContractCmd() cmd.Command {
	return &cmdUpdateContract{
		cmd: &cobra.Command{
			Use:   "update-contract <name> <filename>",
			Short: "Update a contract deployed to an account",
			Args:  cobra.ExactArgs(2),
		},
	}
}

func (c *cmdUpdateContract) Run(
	cmd *cobra.Command,
	args []string,
	project *cli.Project,
	services *services.Services,
) (cmd.Result, error) {

	account, err := services.Accounts.AddContract(c.flags.Account, args[0], args[1], true)
	return &AccountResult{
		Account:  account,
		showCode: true,
	}, err

}

func (c *cmdUpdateContract) GetFlags() *sconfig.Config {
	return sconfig.New(&c.flags)
}

func (c *cmdUpdateContract) GetCmd() *cobra.Command {
	return c.cmd
}
