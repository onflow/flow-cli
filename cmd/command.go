package cmd

import (
	"github.com/onflow/flow-cli/flow/cli"
	"github.com/onflow/flow-cli/sharedlib/services"
	"github.com/psiemens/sconfig"
	"github.com/spf13/cobra"
)

type Command interface {
	GetCmd() *cobra.Command
	SetFlags() *sconfig.Config
	Run(*cobra.Command, []string, *cli.Project, services.Service) (Result, error)
	// ValidateFlags
	// ValidateArgs
}
