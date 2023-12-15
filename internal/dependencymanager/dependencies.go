package dependencymanager

import (
	"github.com/spf13/cobra"
)

var Cmd = &cobra.Command{
	Use:              "dependencies",
	Short:            "Manage contracts and dependencies",
	TraverseChildren: true,
	GroupID:          "manager",
	Aliases:          []string{"deps"},
}

func init() {
	addCommand.AddToParent(Cmd)
	installCommand.AddToParent(Cmd)
}
