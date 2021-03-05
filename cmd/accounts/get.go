package accounts

import (
	"bytes"
	"fmt"
	"github.com/onflow/flow-cli/cmd"
	"github.com/onflow/flow-cli/flow/cli"
	"github.com/onflow/flow-cli/sharedlib/services"
	"github.com/onflow/flow-go-sdk"
	"github.com/psiemens/sconfig"
	"github.com/spf13/cobra"
	"strings"
	"text/tabwriter"
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

func (a *AccountCmd) Run(cmd *cobra.Command, args []string, project *cli.Project, services services.Service) (cmd.Result, error) {
	account, err := services.GetAccount(args[0])
	return &Result{account}, err
}

func (a *AccountCmd) SetFlags() *sconfig.Config {
	return sconfig.New(&a.flags)
}

func (a *AccountCmd) GetCmd() *cobra.Command {
	return a.cmd
}

type Result struct {
	*flow.Account
}

func (r *Result) JSON() interface{} {
	return r
}

func (r *Result) String() string {
	var b bytes.Buffer
	writer := tabwriter.NewWriter(&b, 0, 8, 1, '\t', tabwriter.AlignRight)

	fmt.Fprintf(writer, "Address\t %s\n", r.Address)
	fmt.Fprintf(writer, "Balance\t %d\n", r.Balance)

	fmt.Fprintf(writer, "Keys\t %d\n", len(r.Keys))

	for i, key := range r.Keys {
		fmt.Fprintf(writer, "\nKey %d\tPublic Key\t %x\n", i, key.PublicKey.Encode())
		fmt.Fprintf(writer, "\tWeight\t %d\n", key.Weight)
		fmt.Fprintf(writer, "\tSignature Algorithm\t %s\n", key.SigAlgo)
		fmt.Fprintf(writer, "\tHash Algorithm\t %s\n", key.HashAlgo)
		fmt.Fprintf(writer, "\n")
	}

	fmt.Fprintf(writer, "\nCode\t\t %s", strings.ReplaceAll(string(r.Code), "\n", "\n\t "))
	fmt.Fprintf(writer, "\n")

	writer.Flush()

	return b.String()
}

func (r *Result) Oneliner() string {
	return fmt.Sprintf("Address: %s, Balance: %v, Keys: %s", r.Address, r.Balance, r.Keys[0].PublicKey)
}
