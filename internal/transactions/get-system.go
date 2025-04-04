package transactions

import (
	"context"
	"strings"

	flowsdk "github.com/onflow/flow-go-sdk"
	"github.com/spf13/cobra"

	"github.com/onflow/flowkit/v2"
	"github.com/onflow/flowkit/v2/output"

	"github.com/onflow/flow-cli/internal/command"
)

type flagsGetSystem struct {
	Include []string `default:"" flag:"include" info:"Fields to include in the output. Valid values: signatures, code, payload, fee-events."`
	Exclude []string `default:"" flag:"exclude" info:"Fields to exclude from the output. Valid values: events."`
}

var getSystemFlags = flagsGetSystem{}

var getSystemCommand = &command.Command{
	Cmd: &cobra.Command{
		Use:     "get-system <block_id>",
		Short:   "Get the system transaction by block ID",
		Example: "flow transactions get-system a1b2c3...",
		Args:    cobra.ExactArgs(1),
	},
	Flags: &getSystemFlags,
	Run:   getSystemTransaction,
}

func getSystemTransaction(
	args []string,
	_ command.GlobalFlags,
	_ output.Logger,
	_ flowkit.ReaderWriter,
	flow flowkit.Services,
) (command.Result, error) {
	blockID := flowsdk.HexToID(strings.TrimPrefix(args[0], "0x"))

	tx, result, err := flow.GetSystemTransaction(context.Background(), blockID)
	if err != nil {
		return nil, err
	}

	return &transactionResult{
		result:  result,
		tx:      tx,
		include: getSystemFlags.Include,
		exclude: getSystemFlags.Exclude,
	}, nil
}
