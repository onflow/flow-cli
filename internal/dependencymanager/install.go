package dependencymanager

import (
	"fmt"

	"github.com/onflow/flow-cli/flowkit"
	"github.com/onflow/flow-cli/flowkit/output"
	"github.com/onflow/flow-cli/internal/command"
	"github.com/spf13/cobra"
)

type installFlagsCollection struct{}

var installFlags = installFlagsCollection{}

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
	logger.StartProgress("Installing dependencies from flow.json...")
	defer logger.StopProgress()

	installer := NewContractInstaller(logger, state)
	if err := installer.install(); err != nil {
		logger.Error(fmt.Sprintf("Error: %v", err))
		return nil, err
	}

	logger.Info("✅  Dependencies installed. Check your flow.json")

	return nil, nil
}
