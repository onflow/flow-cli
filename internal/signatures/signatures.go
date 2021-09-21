package signatures

import (
	"github.com/spf13/cobra"
)

var Cmd = &cobra.Command{
	Use:              "signatures",
	Short:            "Signature verification and creation",
	TraverseChildren: true,
}

func init() {
	SignCommand.AddToParent(Cmd)
	VerifyCommand.AddToParent(Cmd)
}
