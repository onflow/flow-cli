package languageserver

import (
	"github.com/onflow/cadence/languageserver"
	"github.com/spf13/cobra"
)

var Cmd = &cobra.Command{
	Use:   "language-server",
	Short: "Start the Cadence language server",
	Run: func(cmd *cobra.Command, args []string) {
		languageserver.RunWithStdio()
	},
}
