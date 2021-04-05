package transactions

import (
	"fmt"

	"github.com/onflow/flow-cli/internal/command"
	"github.com/onflow/flow-cli/pkg/flowcli/services"
	"github.com/spf13/cobra"
)

type flagsStatus struct{}

var statusFlags = flagsStatus{}

var StatusCommand = &command.Command{
	Cmd: &cobra.Command{
		Use:   "status",
		Short: "⚠️  No longer supported: use 'transactions get' command",
		Args:  cobra.ExactArgs(1),
	},
	Flags: &statusFlags,
	Run: func(
		cmd *cobra.Command,
		args []string,
		globalFlags command.GlobalFlags,
		services *services.Services,
	) (command.Result, error) {
		return nil, fmt.Errorf("⚠️  No longer supported: use 'transactions get' command.")
	},
}
