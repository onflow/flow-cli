package accounts

import (
	"fmt"
	"time"

	"github.com/onflow/flow-cli/flowkit"
	"github.com/onflow/flow-cli/flowkit/output"
	"github.com/onflow/flow-cli/internal/command"
	flowsdk "github.com/onflow/flow-go-sdk"
	"github.com/pkg/browser"
	"github.com/spf13/cobra"
)

type flagsFund struct {
	Include []string `default:"" flag:"include" info:"Fields to include in the output. Valid values: contracts."`
}

var fundFlags = flagsFund{}

var fundCommand = &command.Command{
	Cmd: &cobra.Command{
		Use:     "fund <address>",
		Short:   "Funds an account by address through the Testnet Faucet",
		Example: "flow accounts fund 8e94eaa81771313a",
		Args:    cobra.ExactArgs(1),
	},
	Flags: &fundFlags,
	Run:   fund,
}

func fund(
	args []string,
	_ command.GlobalFlags,
	logger output.Logger,
	_ flowkit.ReaderWriter,
	flow flowkit.Services,
) (command.Result, error) {
	address := flowsdk.HexToAddress(args[0])
	if !address.IsValid(flowsdk.Testnet) {
		return nil, fmt.Errorf("unsupported address %s, faucet can only work for valid Testnet addresses", address.String())
	}

	link := fmt.Sprintf("https://testnet-faucet.onflow.org/fund-account?address=%s", address.String())

	logger.Info(
		fmt.Sprintf(
			"Opening the Testnet faucet to fund 0x%s on your native browser."+
				"\n\nIf there is an issue, please use this link instead: %s",
			address.String(),
			link,
		))
	// wait for the user to read the message
	time.Sleep(5 * time.Second)

	if err := browser.OpenURL(link); err != nil {
		return nil, err
	}

	return nil, nil
}
