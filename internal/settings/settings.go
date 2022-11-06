package settings

import (
	"github.com/spf13/cobra"
)

var Cmd = &cobra.Command{
	Use:              "settings",
	Short:            "Manage persisted global settings",
	TraverseChildren: true,
}

func init() {
	Cmd.AddCommand(MetricsSettings)
}
