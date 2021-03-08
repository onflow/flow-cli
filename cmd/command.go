package cmd

import (
	"github.com/onflow/flow-cli/flow/cli"
	"github.com/onflow/flow-cli/sharedlib/services"
	"github.com/psiemens/sconfig"
	"github.com/spf13/cobra"
)

type Command interface {
	GetCmd() *cobra.Command
	GetFlags() *sconfig.Config
	Run(*cobra.Command, []string, *cli.Project, *services.Services) (Result, error)
}
