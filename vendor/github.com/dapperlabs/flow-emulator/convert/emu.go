package convert

import (
	"github.com/dapperlabs/flow-go/fvm"

	"github.com/dapperlabs/flow-emulator/types"
)

func ToStorableResult(tp *fvm.TransactionProcedure, txIndex uint32) (types.StorableTransactionResult, error) {
	var errorCode int
	var errorMessage string

	if tp.Err != nil {
		errorCode = int(tp.Err.Code())
		errorMessage = tp.Err.Error()
	}

	events, err := tp.ConvertEvents(txIndex)
	if err != nil {
		return types.StorableTransactionResult{}, err
	}

	return types.StorableTransactionResult{
		ErrorCode:    errorCode,
		ErrorMessage: errorMessage,
		Logs:         tp.Logs,
		Events:       events,
	}, nil
}
