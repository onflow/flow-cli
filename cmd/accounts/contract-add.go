package accounts

import (
	"github.com/onflow/flow-cli/cmd"
	"github.com/onflow/flow-cli/flow/cli"
	"github.com/onflow/flow-cli/sharedlib/services"
	"github.com/psiemens/sconfig"
	"github.com/spf13/cobra"
)

type flagsAddContract struct {
	Account string `default:"emulator-account" flag:"account,a"`
	Host    string `flag:"host" info:"Flow Access API host address"`
}

type cmdAddContract struct {
	cmd   *cobra.Command
	flags flagsAddContract
}

func NewAddContractCmd() cmd.Command {
	return &cmdAddContract{
		cmd: &cobra.Command{
			Use:   "add-contract <name> <filename>",
			Short: "Deploy a new contract to an account",
			Args:  cobra.ExactArgs(2),
		},
	}
}

func (a *cmdAddContract) Run(
	cmd *cobra.Command,
	args []string,
	project *cli.Project,
	services *services.Services,
) (cmd.Result, error) {

	account, err := services.Accounts.AddContract(a.flags.Account, args[0], args[1])
	return &AccountResult{
		Account:  account,
		showCode: true,
	}, err
}

func (a *cmdAddContract) GetFlags() *sconfig.Config {
	return sconfig.New(&a.flags)
}

func (a *cmdAddContract) GetCmd() *cobra.Command {
	return a.cmd
}
