package blocks

import (
	"github.com/onflow/flow-cli/cmd"
	"github.com/onflow/flow-cli/flow/lib"
	"github.com/onflow/flow-cli/flow/services"
	"github.com/psiemens/sconfig"
	"github.com/spf13/cobra"
)

type flagsBlocks struct {
	Events  string `default:"" flag:"events" info:"List events of this type for the block"`
	Verbose bool   `default:"false" flag:"verbose" info:"Display transactions in block"`
}

type cmdGet struct {
	cmd   *cobra.Command
	flags flagsBlocks
}

// NewGetCmd creates new get command
func NewGetCmd() cmd.Command {
	return &cmdGet{
		cmd: &cobra.Command{
			Use:   "get <block_id|latest|block_height>",
			Short: "Get block info",
		},
	}
}

// Run script command
func (s *cmdGet) Run(
	cmd *cobra.Command,
	args []string,
	project *lib.Project,
	services *services.Services,
) (cmd.Result, error) {
	block, events, err := services.Blocks.GetBlock(args[0], s.flags.Events)

	return &BlockResult{
		block:   block,
		events:  events,
		verbose: false,
	}, err
}

// GetFlags for script
func (s *cmdGet) GetFlags() *sconfig.Config {
	return sconfig.New(&s.flags)
}

// GetCmd get command
func (s *cmdGet) GetCmd() *cobra.Command {
	return s.cmd
}
