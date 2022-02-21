package tools

import (
	devWallet "github.com/onflow/fcl-dev-wallet"
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
		Short: "Starts a dev wallet",
	},
	Flags: &walletFlags,
	RunS:  wallet,
}

func wallet(
	_ []string,
	_ flowkit.ReaderWriter,
	_ command.GlobalFlags,
	_ *services.Services,
	state *flowkit.State,
) (command.Result, error) {
	service, err := state.EmulatorServiceAccount()
	if err != nil {
		return nil, err
	}
	// todo check if this is ok and make sure emulator is running
	emulator, err := state.Networks().ByName("emulator")
	if err != nil {
		return nil, err
	}

	key := service.Key().ToConfig()
	conf := devWallet.Config{
		Address:    service.Address().String(),
		PrivateKey: key.PrivateKey.String(),
		PublicKey:  key.PrivateKey.PublicKey().String(),
		AccessNode: emulator.Host,
	}

	srv, err := devWallet.NewHTTPServer(1234, &conf)
	if err != nil {
		return nil, err
	}

	err = srv.Start()
	if err != nil {
		return nil, err
	}

	return nil, nil
}
