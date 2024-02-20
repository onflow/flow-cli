package evm_gateway

import (
	"github.com/onflow/flow-evm-gateway/bootstrap"
	"github.com/onflow/flow-evm-gateway/config"
	"github.com/onflow/flowkit/v2"
	"github.com/onflow/flowkit/v2/output"
	"github.com/spf13/cobra"

	"github.com/onflow/flow-cli/internal/command"
)

type GatewayFlag struct{}

var gatewyFlag = GatewayFlag{}

var GatewayCmd = &command.Command{
	Cmd: &cobra.Command{
		Use:     "snapshot <create|load|list> [snapshotName]",
		Short:   "Create/Load/List emulator snapshots",
		Example: "flow emulator snapshot create testSnapshot",
		Args:    cobra.RangeArgs(1, 2),
	},
	Flags: &gatewyFlag,
	Run: func(
		args []string,
		globalFlags command.GlobalFlags,
		logger output.Logger,
		readerWriter flowkit.ReaderWriter,
		flow flowkit.Services,
	) (command.Result, error) {
		cfg, err := config.FromFlags()
		if err != nil {
			return nil, err
		}

		return nil, bootstrap.Start(cfg)
	},
}
