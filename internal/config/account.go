package config

import (
	"github.com/onflow/flow-cli/internal/command"
	"github.com/spf13/cobra"
)

type flagsAccount struct{}

var accountFlags = flagsAccount{}

var AccountCommand = &command.Command{
	Cmd: &cobra.Command{
		Use:       "add account",
		Short:     "Add resource to configuration",
		Example:   "flow config add account",
		ValidArgs: []string{"account", "contract", "deployment", "network"},
		Args:      cobra.ExactArgs(1),
	},
}
