package dependencymanager

import (
	"fmt"

	"github.com/onflow/flow-cli/flowkit"
	"github.com/onflow/flow-cli/flowkit/output"
	"github.com/onflow/flow-cli/internal/command"
	"github.com/spf13/cobra"
)

type addFlagsCollection struct {
	name string `default:"" flag:"name" info:"Name of the dependency"`
}

var addFlags = addFlagsCollection{}

var addCommand = &command.Command{
	Cmd: &cobra.Command{
		Use:   "add",
		Short: "Add a single contract and its dependencies.",
		Args:  cobra.ExactArgs(1),
	},
	Flags: &addFlags,
	RunS:  add,
}

func add(
	args []string,
	_ command.GlobalFlags,
	logger output.Logger,
	flow flowkit.Services,
	state *flowkit.State,
) (result command.Result, err error) {
	logger.StartProgress(fmt.Sprintf("Installing dependencies for %s...", args[0]))
	defer logger.StopProgress()

	dep := args[0]

	installer := NewDepdencyInstaller(logger, state)
	if err := installer.add(dep, addFlags.name); err != nil {
		logger.Error(fmt.Sprintf("Error: %v", err))
		return nil, err
	}

	logger.Info("✅  Dependencies installed. Check your flow.json")

	return nil, nil
}
