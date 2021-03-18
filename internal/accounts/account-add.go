package accounts

import (
	"bytes"
	"fmt"
	"text/tabwriter"

	"github.com/onflow/flow-cli/internal/command"
	"github.com/onflow/flow-cli/pkg/flow"
	"github.com/onflow/flow-cli/pkg/flow/services"
	"github.com/psiemens/sconfig"
	"github.com/spf13/cobra"
)

type flagsAdd struct {
	SigAlgo    string `default:"ECDSA_P256" flag:"algo,a" info:"Signature algorithm"`
	HashAlgo   string `default:"SHA3_256" flag:"service-hash-algo" info:"Hashing algorithm"`
	KeyIndex   int    `flag:"index" info:"Account key index"`
	KeyHex     string `flag:"privateKey" info:"Private key in hex format"`
	KeyContext string `flag:"context" info:"Projects/<PROJECTID>/locations/<LOCATION>/keyRings/<KEYRINGID>/cryptoKeys/<KEYID>/cryptoKeyVersions/<KEYVERSION>"`
	Overwrite  bool   `flag:"overwrite" info:"Overwrite an existing account"`
}

type cmdAdd struct {
	cmd   *cobra.Command
	flags flagsAdd
}

func NewAddCmd() command.Command {
	return &cmdAdd{
		cmd: &cobra.Command{
			Use:     "add <name> <address>",
			Short:   "Add account by name to config",
			Example: `flow accounts add Alice 18d6e0586b0a20c5 --privateKey=11c5dfdeb0ff03a7a73ef39788563b62c89adea67bbb21ab95e5f710bd1d40b7`,
			Args:    cobra.ExactArgs(2),
		},
	}
}

func (a *cmdAdd) Run(
	cmd *cobra.Command,
	args []string,
	project *flow.Project,
	services *services.Services,
) (command.Result, error) {
	account, err := services.Accounts.Add(
		args[0],
		args[1],
		a.flags.SigAlgo,
		a.flags.HashAlgo,
		a.flags.KeyIndex,
		a.flags.KeyHex,
		a.flags.KeyContext,
		a.flags.Overwrite,
		flow.ConfigPath,
	)

	return &AccountAddResult{account}, err
}

func (a *cmdAdd) GetFlags() *sconfig.Config {
	return sconfig.New(&a.flags)
}

func (a *cmdAdd) GetCmd() *cobra.Command {
	return a.cmd
}

// AccountResult represent result from all account commands
type AccountAddResult struct {
	*flow.Account
}

// JSON convert result to JSON
func (r *AccountAddResult) JSON() interface{} {
	return r
}

// String convert result to string
func (r *AccountAddResult) String() string {
	var b bytes.Buffer
	writer := tabwriter.NewWriter(&b, 0, 8, 1, '\t', tabwriter.AlignRight)

	fmt.Fprintf(writer, "Address\t 0x%s\n", r.Account.Address())
	fmt.Fprintf(writer, "Hash Algo\t %s\n", r.Account.DefaultKey().HashAlgo())
	fmt.Fprintf(writer, "Signature Algo\t %s\n", r.Account.DefaultKey().SigAlgo())
	fmt.Fprintf(writer, "Key Index\t %d\n", r.Account.DefaultKey().Index())

	writer.Flush()

	return b.String()
}

// Oneliner show result as one liner grep friendly
func (r *AccountAddResult) Oneliner() string {
	return fmt.Sprintf(
		"Address: 0x%s, Hash Algo: %s, Sig Algo: %s",
		r.Address(),
		r.Account.DefaultKey().HashAlgo(),
		r.Account.DefaultKey().SigAlgo(),
	)
}
