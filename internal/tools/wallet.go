package tools

import (
	"github.com/onflow/flow-cli/internal/command"
	"github.com/onflow/flow-cli/pkg/flowkit"
	"github.com/onflow/flow-cli/pkg/flowkit/services"
	"github.com/spf13/cobra"
)

type FlagsWallet struct {
}

var walletFlags = FlagsWallet{}

var DevWallet = &command.Command{
	Cmd: &cobra.Command{
		Use:   "dev-wallet",
		Short: "Start a dev wallet",
	},
	Flags: &walletFlags,
	RunS:  wallet,
}

func wallet(
	_ []string,
	_ flowkit.ReaderWriter,
	globalFlags command.GlobalFlags,
	services *services.Services,
	_ *flowkit.State,
) (command.Result, error) {

	return nil, nil
}
