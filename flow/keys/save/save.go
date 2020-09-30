package save

import (
	"github.com/spf13/cobra"

	"github.com/onflow/flow-cli/flow/keys/save/hex"
	"github.com/onflow/flow-cli/flow/keys/save/kms"
)

var Cmd = &cobra.Command{
	Use:              "save",
	Short:            "save a key to the config",
	TraverseChildren: true,
}

func init() {
	Cmd.AddCommand(hex.Cmd)
	Cmd.AddCommand(kms.Cmd)
}
