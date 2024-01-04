package accounts

import (
	"context"
	"fmt"

	"github.com/onflow/flow-cli/flowkit"
	"github.com/onflow/flow-cli/flowkit/output"
	"github.com/onflow/flow-cli/internal/command"
	flowsdk "github.com/onflow/flow-go-sdk"
	"github.com/spf13/cobra"
)

var fundCommand = &command.Command{
	Cmd: &cobra.Command{
		Use:     "fund <address>",
		Short:   "Funds an account by address through the Testnet Faucet",
		Example: "flow accounts fund 8e94eaa81771313a",
		Args:    cobra.ExactArgs(1),
	},
	Run: fund,
}

func fund(
	args []string,
	_ command.GlobalFlags,
	logger output.Logger,
	_ flowkit.ReaderWriter,
	flow flowkit.Services,
) (command.Result, error) {
	ctx := context.Background()
	address := flowsdk.HexToAddress(args[0])

	if !address.IsValid(flowsdk.Testnet) {
		return nil, fmt.Errorf("unsupported address %s, faucet can only work for Testnet", address.String())
	}

	logger.StartProgress(fmt.Sprintf("Funding account %s...", address.String()))
	fmt.Println("hello")
	//if err := browser.OpenURL(fmt.Sprintf("https://testnet-faucet.onflow.org/fund-account?address=%s", address.String())); err != nil {
	//	return nil, err
	//}
	//err := chromedp.Run(context.Background(),
	//	chromedp.Navigate(fmt.Sprintf("https://testnet-faucet.onflow.org/fund-account?address=%s", address.String())),
	//	// Add more tasks here if needed
	//)
	//
	//if err != nil {
	//	return nil, err
	//}

	account, err := flow.GetAccount(ctx, address)
	if err != nil {
		return nil, err
	}

	return &accountResult{
		Account: account,
	}, nil
}
