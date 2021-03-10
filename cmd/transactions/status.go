package transactions

import (
	"github.com/onflow/flow-cli/cmd"
	"github.com/onflow/flow-cli/flow/cli"
	"github.com/onflow/flow-cli/sharedlib/services"
	"github.com/psiemens/sconfig"
	"github.com/spf13/cobra"
)

type flagsStatus struct {
	Host   string `flag:"host" info:"Flow Access API host address"`
	Sealed bool   `default:"true" flag:"sealed" info:"Wait for a sealed result"`
	Code   bool   `default:"false" flag:"code" info:"Display transaction code"`
}

type cmdStatus struct {
	cmd   *cobra.Command
	flags flagsStatus
}

func NewStatusCmd() cmd.Command {
	return &cmdStatus{
		cmd: &cobra.Command{
			Use:   "status <tx_id>",
			Short: "Get the transaction status",
			Args:  cobra.ExactArgs(1),
		},
	}
}

// Run command
func (s *cmdStatus) Run(
	cmd *cobra.Command,
	args []string,
	project *cli.Project,
	services *services.Services,
) (cmd.Result, error) {
	tx, result, err := services.Transactions.GetStatus(args[0], s.flags.Sealed)
	return &TransactionResult{
		result: result,
		tx:     tx,
		code:   s.flags.Code,
	}, err
}

// GetFlags for command
func (s *cmdStatus) GetFlags() *sconfig.Config {
	return sconfig.New(&s.flags)
}

// GetCmd gets command
func (s *cmdStatus) GetCmd() *cobra.Command {
	return s.cmd
}
