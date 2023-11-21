package contracts

import (
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
	_ flowkit.Services,
	_ *flowkit.State,
) (result command.Result, err error) {
	logger.Info("Fetching contract and dependencies...")
	return nil, err
}
