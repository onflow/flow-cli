package contracts

import (
	"github.com/spf13/cobra"
)

var Cmd = &cobra.Command{
	Use:              "contracts",
	Short:            "Manage contracts and dependencies",
	TraverseChildren: true,
	GroupID:          "manager",
}

func init() {
	installCommand.AddToParent(Cmd)
}
