package collections

import (
	"github.com/onflow/flow-cli/cmd"
	"github.com/onflow/flow-cli/flow/cli"
	"github.com/onflow/flow-cli/sharedlib/services"
	"github.com/psiemens/sconfig"
	"github.com/spf13/cobra"
)

type flagsCollections struct {
}

type cmdGet struct {
	cmd   *cobra.Command
	flags flagsCollections
}

// NewExecuteScriptCmd creates new script command
func NewGetCmd() cmd.Command {
	return &cmdGet{
		cmd: &cobra.Command{
			Use:   "get <collection_id>",
			Short: "Get collection info",
			Args:  cobra.ExactArgs(1),
		},
	}
}

// Run script command
func (s *cmdGet) Run(
	cmd *cobra.Command,
	args []string,
	project *cli.Project,
	services *services.Services,
) (cmd.Result, error) {
	collection, err := services.Collections.Get(args[0])
	return &CollectionResult{collection}, err
}

// GetFlags for script
func (s *cmdGet) GetFlags() *sconfig.Config {
	return sconfig.New(&s.flags)
}

// GetCmd get command
func (s *cmdGet) GetCmd() *cobra.Command {
	return s.cmd
}
