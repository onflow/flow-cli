package contracts

import (
	"context"
	"fmt"

	"github.com/onflow/flow-cli/flowkit"
	"github.com/onflow/flow-cli/flowkit/output"
	"github.com/onflow/flow-cli/internal/command"
	flowsdk "github.com/onflow/flow-go-sdk"
	"github.com/spf13/cobra"
)

type flagsCollection struct{}

var installFlags = flagsCollection{}

var installCommand = &command.Command{
	Cmd: &cobra.Command{
		Use:   "install",
		Short: "Install contract and dependencies.",
	},
	Flags: &installFlags,
	RunS:  install,
}

func install(
	args []string,
	_ command.GlobalFlags,
	logger output.Logger,
	flow flowkit.Services,
	_ *flowkit.State,
) (result command.Result, err error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("please specify a contract address")
	}

	address := flowsdk.HexToAddress(args[0])
	logger.Info(fmt.Sprintf("Fetching contract and dependencies for %s", address))
	account, err := flow.GetAccount(context.Background(), address)
	if err != nil {
		return nil, err
	}

	fmt.Println("account: ", len(account.Contracts))

	for name, code := range account.Contracts {
		fmt.Println("name: ", name)
		fmt.Println("code: ", string(code))
	}

	return nil, err
}
