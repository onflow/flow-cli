package convert

import (
	"github.com/dapperlabs/flow-go/engine/execution/computation/virtualmachine"

	"github.com/dapperlabs/flow-emulator/types"
)

func ToStorableResult(tr *virtualmachine.TransactionResult, txIndex uint32) (types.StorableTransactionResult, error) {
	var errorCode int
	var errorMessage string

	if tr.Error != nil {
		errorCode = int(tr.Error.StatusCode())
		errorMessage = tr.Error.ErrorMessage()
	}

	events, err := virtualmachine.ConvertEvents(txIndex, tr)
	if err != nil {
		return types.StorableTransactionResult{}, err
	}
	return types.StorableTransactionResult{
		ErrorCode:    errorCode,
		ErrorMessage: errorMessage,
		Logs:         tr.Logs,
		Events:       events,
	}, nil
}
