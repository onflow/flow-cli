package keys

import (
	"github.com/onflow/flow-cli/internal/command"
	"github.com/onflow/flow-cli/pkg/flow/services"
	"github.com/spf13/cobra"
)

type flagsDecode struct{}

var decodeFlags = flagsDecode{}

var DecodeCommand = &command.Command{
	Cmd: &cobra.Command{
		Use:   "decode <public key>",
		Short: "Decode a public account key hex string",
		Args:  cobra.ExactArgs(1),
	},
	Flags: &decodeFlags,
	Run: func(
		cmd *cobra.Command,
		args []string,
		services *services.Services,
	) (command.Result, error) {
		accountKey, err := services.Keys.Decode(args[0])
		if err != nil {
			return nil, err
		}

		pubKey := accountKey.PublicKey
		return &KeyResult{publicKey: &pubKey, accountKey: accountKey}, err
	},
}
