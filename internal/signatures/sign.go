package signatures

import (
	"context"

	"github.com/onflow/flow-cli/internal/command"
	"github.com/onflow/flow-cli/pkg/flowkit"
	"github.com/onflow/flow-cli/pkg/flowkit/services"
	"github.com/spf13/cobra"
)

type flagsSign struct {
	Signer string `default:"emulator-account" flag:"signer" info:"name of the account used to sign"`
}

var signFlags = flagsSign{}

var SignCommand = &command.Command{
	Cmd: &cobra.Command{
		Use:     "sign <payload>",
		Short:   "Sign the payload data",
		Example: "flow signatures sign 'The quick brown fox jumps over the lazy dog' --signer alice",
		Args:    cobra.ExactArgs(1),
	},
	Flags: &signFlags,
	RunS:  sign,
}

func sign(
	args []string,
	_ flowkit.ReaderWriter,
	_ command.GlobalFlags,
	_ *services.Services,
	state *flowkit.State,
) (command.Result, error) {
	payload := []byte(args[0])
	accountName := signFlags.Signer
	acc, err := state.Accounts().ByName(accountName)
	if err != nil {
		return nil, err
	}

	s, err := acc.Key().Signer(context.Background())
	if err != nil {
		return nil, err
	}

	signed, err := s.Sign(payload)
	if err != nil {
		return nil, err
	}

	return &SignatureResult{
		result: string(signed),
	}, nil
}
