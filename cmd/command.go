package cmd

import (
	"github.com/onflow/flow-cli/flow/lib"
	"github.com/onflow/flow-cli/flow/services"
	"github.com/psiemens/sconfig"
	"github.com/spf13/cobra"
)

type Command interface {
	GetCmd() *cobra.Command
	GetFlags() *sconfig.Config
	Run(*cobra.Command, []string, *lib.Project, *services.Services) (Result, error)
}
