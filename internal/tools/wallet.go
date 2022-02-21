package tools

import (
	"fmt"
	devWallet "github.com/onflow/fcl-dev-wallet"
	"github.com/onflow/flow-cli/internal/command"
	"github.com/onflow/flow-cli/pkg/flowkit"
	"github.com/onflow/flow-cli/pkg/flowkit/output"
	"github.com/onflow/flow-cli/pkg/flowkit/services"
	"github.com/spf13/cobra"
)

type FlagsWallet struct {
	Port         uint `default:"8081" flag:"port" info:"Dev wallet port to listen on"`
	EmulatorPort uint `default:"3569" flag:"emulator-port" info:"The port emulator is listening on"`
}

var walletFlags = FlagsWallet{}

var DevWallet = &command.Command{
	Cmd: &cobra.Command{
		Use:     "dev-wallet",
		Short:   "Starts a dev wallet",
		Example: "flow dev-wallet",
		Args:    cobra.ExactArgs(0),
	},
	Flags: &walletFlags,
	RunS:  wallet,
}

func wallet(
	_ []string,
	_ flowkit.ReaderWriter,
	global command.GlobalFlags,
	_ *services.Services,
	state *flowkit.State,
) (command.Result, error) {
	service, err := state.EmulatorServiceAccount()
	if err != nil {
		return nil, err
	}

	key := service.Key().ToConfig()
	conf := devWallet.Config{
		Address:    service.Address().String(),
		PrivateKey: key.PrivateKey.String(),
		PublicKey:  key.PrivateKey.PublicKey().String(),
		AccessNode: fmt.Sprintf("localhost:%d", walletFlags.EmulatorPort),
	}

	srv, err := devWallet.NewHTTPServer(walletFlags.Port, &conf)
	if err != nil {
		return nil, err
	}

	fmt.Printf("%s Starting dev wallet server on port %d\n", output.SuccessEmoji(), walletFlags.Port)
	fmt.Printf("%s  Make sure the emulator is running and listening on port %d\n", output.WarningEmoji(), walletFlags.EmulatorPort)

	err = srv.Start()
	if err != nil {
		return nil, err
	}

	return nil, nil
}
