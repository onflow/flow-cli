package contracts

import (
	"fmt"

	"github.com/onflow/flow-cli/flowkit"
	"github.com/onflow/flow-cli/flowkit/output"
	"github.com/onflow/flow-cli/internal/command"
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
	_ []string,
	_ command.GlobalFlags,
	logger output.Logger,
	flow flowkit.Services,
	state *flowkit.State,
) (result command.Result, err error) {
	//address := flowsdk.HexToAddress(args[0])
	//logger.Info(fmt.Sprintf("Fetching contract and dependencies for %s", address))
	//account, err := flow.GetAccount(context.Background(), address)
	//if err != nil {
	//	return nil, err
	//}
	//
	//fmt.Println("account: ", len(account.Contracts))

	for _, dependency := range *state.Dependencies() {
		fmt.Println("dependency: ", dependency.Name)
	}

	return nil, err
}
