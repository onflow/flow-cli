package state

import (
	"github.com/onflow/flow-cli/internal/command"
	"github.com/onflow/flow-cli/pkg/flowcli/output"
	"github.com/onflow/flow-cli/pkg/flowcli/services"
	"github.com/spf13/cobra"
)

type flagsRemove struct {
	Name string `flag:"name" info:"Name of the snapshot"`
}

var removeFlags = flagsRemove{}

var RemoveCommand = &command.Command{
	Cmd: &cobra.Command{
		Use:     "remove",
		Short:   "remove snapshot",
		Example: "flow state remove",
		Args:    cobra.NoArgs,
	},
	Flags: &removeFlags,
	Run: func(
		cmd *cobra.Command,
		args []string,
		globalFlags command.GlobalFlags,
		services *services.Services,
	) (command.Result, error) {
		name := removeFlags.Name

		if name == "" {
			names, err := services.State.List()
			if err != nil {
				return nil, err
			}

			name = output.SnapshotPrompt(names)
		}

		err := services.State.Remove(name)
		if err != nil {
			return nil, err
		}

		return &StateResult{
			result: "snapshot removed",
		}, nil
	},
}

func init() {
	RemoveCommand.AddToParent(Cmd)
}
