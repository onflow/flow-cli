package events

import (
	"github.com/onflow/flow-cli/cmd"
	"github.com/onflow/flow-cli/flow/cli"
	"github.com/onflow/flow-cli/sharedlib/services"
	"github.com/psiemens/sconfig"
	"github.com/spf13/cobra"
)

type flagsGenerate struct {
}

type cmdGet struct {
	cmd   *cobra.Command
	flags flagsGenerate
}

// NewGetCmd return new command
func NewGetCmd() cmd.Command {
	return &cmdGet{
		cmd: &cobra.Command{
			Use:     "get <event_name> <block_height_range_start> <optional:block_height_range_end|latest>",
			Short:   "Get events in a block range",
			Args:    cobra.RangeArgs(2, 3),
			Example: "flow events get A.1654653399040a61.FlowToken.TokensDeposited 11559500 11559600",
		},
	}
}

func (a *cmdGet) Run(
	cmd *cobra.Command,
	args []string,
	project *cli.Project,
	services *services.Services,
) (cmd.Result, error) {
	end := ""
	if len(args) == 3 {
		end = args[2]
	}

	events, err := services.Events.Get(args[0], args[1], end)
	return &EventResult{BlockEvents: events}, err
}

func (a *cmdGet) GetFlags() *sconfig.Config {
	return sconfig.New(&a.flags)
}

func (a *cmdGet) GetCmd() *cobra.Command {
	return a.cmd
}
