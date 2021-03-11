package transactions

import (
	"github.com/onflow/flow-cli/cmd"
	"github.com/onflow/flow-cli/flow/cli"
	"github.com/onflow/flow-cli/sharedlib/services"
	"github.com/psiemens/sconfig"
	"github.com/spf13/cobra"
)

type flagsSend struct {
	ArgsJSON string   `default:"" flag:"argsJSON" info:"arguments in JSON-Cadence format"`
	Args     []string `default:"" flag:"arg" info:"argument in Type:Value format"`
	Signer   string   `default:"emulator-account" flag:"signer,s"`
}

type cmdSend struct {
	cmd   *cobra.Command
	flags flagsSend
}

func NewSendCmd() cmd.Command {
	return &cmdSend{
		cmd: &cobra.Command{
			Use:     "send <filename>",
			Short:   "Send a transaction",
			Example: `flow transactions send tx.cdc --args String:"Hello world"`,
		},
	}
}

func (a *cmdSend) Run(
	cmd *cobra.Command,
	args []string,
	project *cli.Project,
	services *services.Services,
) (cmd.Result, error) {
	tx, result, err := services.Transactions.Send(args[0], a.flags.Signer, a.flags.Args, a.flags.ArgsJSON)
	return &TransactionResult{
		result: result,
		tx:     tx,
	}, err
}

func (a *cmdSend) GetFlags() *sconfig.Config {
	return sconfig.New(&a.flags)
}

func (a *cmdSend) GetCmd() *cobra.Command {
	return a.cmd
}
