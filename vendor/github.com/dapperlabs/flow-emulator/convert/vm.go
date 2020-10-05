package convert

import (
	"github.com/onflow/flow-go/fvm"

	sdkConvert "github.com/dapperlabs/flow-emulator/convert/sdk"
	"github.com/dapperlabs/flow-emulator/types"
)

func VMTransactionResultToEmulator(tp *fvm.TransactionProcedure, txIndex int) types.TransactionResult {
	txID := sdkConvert.FlowIdentifierToSDK(tp.ID)

	sdkEvents := sdkConvert.RuntimeEventsToSDK(tp.Events, txID, txIndex)

	return types.TransactionResult{
		TransactionID: txID,
		Error:         VMErrorToEmulator(tp.Err),
		Logs:          tp.Logs,
		Events:        sdkEvents,
	}
}

func VMErrorToEmulator(vmError fvm.Error) error {
	if vmError == nil {
		return nil
	}

	return &types.FlowError{FlowError: vmError}
}
