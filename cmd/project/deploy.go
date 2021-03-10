package project

import (
	"github.com/onflow/flow-cli/cmd"
	"github.com/onflow/flow-cli/flow/cli"
	"github.com/onflow/flow-cli/sharedlib/services"
	"github.com/psiemens/sconfig"
	"github.com/spf13/cobra"
)

type flagsDeploy struct {
}

type cmdDeploy struct {
	cmd   *cobra.Command
	flags flagsDeploy
}

// NewExecuteScriptCmd creates new script command
func NewGetCmd() cmd.Command {
	return &cmdDeploy{
		cmd: &cobra.Command{
			Use:   "get <collection_id>",
			Short: "Get collection info",
			Args:  cobra.ExactArgs(1),
		},
	}
}

// Run script command
func (s *cmdDeploy) Run(
	cmd *cobra.Command,
	args []string,
	project *cli.Project,
	services *services.Services,
) (cmd.Result, error) {
	return nil, nil
}

// GetFlags for script
func (s *cmdDeploy) GetFlags() *sconfig.Config {
	return sconfig.New(&s.flags)
}

// GetCmd get command
func (s *cmdDeploy) GetCmd() *cobra.Command {
	return s.cmd
}
