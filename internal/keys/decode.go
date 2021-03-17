package keys

import (
	"github.com/onflow/flow-cli/internal/command"
	"github.com/onflow/flow-cli/pkg/flow"
	"github.com/onflow/flow-cli/pkg/flow/services"
	"github.com/psiemens/sconfig"
	"github.com/spf13/cobra"
)

type flagsDecode struct{}

type cmdDecode struct {
	cmd   *cobra.Command
	flags flagsGenerate
}

// NewGenerateCmd return new command
func NewCmdDecode() command.Command {
	return &cmdDecode{
		cmd: &cobra.Command{
			Use:   "decode <public key>",
			Short: "Decode a public account key hex string",
			Args:  cobra.ExactArgs(1),
		},
	}
}

// Run command
func (a *cmdDecode) Run(
	cmd *cobra.Command,
	args []string,
	project *flow.Project,
	services *services.Services,
) (command.Result, error) {
	accountKey, err := services.Keys.Decode(args[0])
	if err != nil {
		return nil, err
	}

	pubKey := accountKey.PublicKey
	return &KeyResult{publicKey: &pubKey, accountKey: accountKey}, err
}

// GetFlags get command flags
func (a *cmdDecode) GetFlags() *sconfig.Config {
	return sconfig.New(&a.flags)
}

// GetCmd gets command
func (a *cmdDecode) GetCmd() *cobra.Command {
	return a.cmd
}
