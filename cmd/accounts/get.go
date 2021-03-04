package accounts

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/onflow/flow-cli/cmd"
	"github.com/onflow/flow-cli/flow/cli"
	"github.com/onflow/flow-cli/services"
	"github.com/onflow/flow-go-sdk"
	"github.com/psiemens/sconfig"
	"github.com/spf13/cobra"
)

type Flags struct {
	Host string `flag:"host" info:"Flow Access API host address"`
	Code bool   `default:"false" flag:"code" info:"Display code deployed to the account"`
}

type AccountCmd struct {
	cmd   *cobra.Command
	flags Flags
}

func Init() cmd.Command {
	return &AccountCmd{
		cmd: &cobra.Command{
			Use:     "get <address>",
			Short:   "Gets an account by address",
			Aliases: []string{"fetch", "g"},
			Long:    `Gets an account by address (address, balance, keys, code)`,
			Args:    cobra.ExactArgs(1),
		},
	}
}

func (a *AccountCmd) SetFlags() *sconfig.Config {
	return sconfig.New(&a.flags)
}

func (a *AccountCmd) ValidateFlags() error {
	return nil
}

func (a *AccountCmd) Run(cmd *cobra.Command, args []string, project *cli.Project) (cmd.Result, error) {
	host := a.flags.Host

	if host == "" && project != nil {
		host = project.Host("emulator")
	} else if host == "" {
		return nil, errors.New("Host must be provided using --host flag or by initializing project: flow project init")
	}

	account, err := services.GetAccount(host, args[0])

	return &Result{account}, err
}

func (a *AccountCmd) GetCmd() *cobra.Command {
	return a.cmd
}

type Result struct {
	*flow.Account
}

func (r *Result) JSON() string {
	val, _ := json.Marshal(r.Account)
	return string(val)
}

func (r *Result) String() string {
	return fmt.Sprintf("address: %s balance: %v", r.Account.Address.String(), r.Account.Balance)
}
