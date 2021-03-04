package cmd

import (
	"github.com/onflow/flow-cli/flow/cli"
	"github.com/psiemens/sconfig"
	"github.com/spf13/cobra"
)

type Command interface {
	GetCmd() *cobra.Command
	SetFlags() *sconfig.Config
	ValidateFlags() error
	Run(*cobra.Command, []string, *cli.Project) (Result, error)
}
