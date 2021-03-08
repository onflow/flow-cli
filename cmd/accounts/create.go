package accounts

import (
	"github.com/onflow/flow-cli/cmd"
	"github.com/onflow/flow-cli/flow/cli"
	"github.com/onflow/flow-cli/sharedlib/services"
	"github.com/psiemens/sconfig"
	"github.com/spf13/cobra"
)

type flagsCreate struct {
	Signer    string   `default:"emulator-account" flag:"signer,s"`
	Keys      []string `flag:"key,k" info:"Public keys to attach to account"`
	SigAlgo   string   `default:"ECDSA_P256" flag:"sig-algo" info:"Signature algorithm used to generate the keys"`
	HashAlgo  string   `default:"SHA3_256" flag:"hash-algo" info:"Hash used for the digest"`
	Host      string   `flag:"host" info:"Flow Access API host address"`
	Results   bool     `default:"false" flag:"results" info:"Display the results of the transaction"`
	Contracts []string `flag:"contract,c" info:"Contract to be deployed during account creation. <name:path>"`
}

type cmdCreate struct {
	cmd   *cobra.Command
	flags flagsCreate
}

func NewCreateCmd() cmd.Command {
	return &cmdCreate{
		cmd: &cobra.Command{
			Use:     "create",
			Short:   "Create a new account",
			Aliases: []string{"create"},
			Long:    `Create new account with keys`,
		},
	}
}

func (a *cmdCreate) Run(
	cmd *cobra.Command,
	args []string,
	project *cli.Project,
	services *services.Services,
) (cmd.Result, error) {

	account, err := services.Accounts.Create(
		a.flags.Signer,
		a.flags.Keys,
		a.flags.SigAlgo,
		a.flags.HashAlgo,
		a.flags.Contracts,
	)

	return &AccountResult{account}, err
}

func (a *cmdCreate) GetFlags() *sconfig.Config {
	return sconfig.New(&a.flags)
}

func (a *cmdCreate) GetCmd() *cobra.Command {
	return a.cmd
}
