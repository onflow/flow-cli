package emulator

import (
	"github.com/onflow/cadence"
	"github.com/onflow/flow-go-sdk"

	"github.com/dapperlabs/flow-emulator/types"
)

// A TransactionResult is the result of executing a transaction.
type TransactionResult struct {
	TransactionID flow.Identifier
	Error         error
	Logs          []string
	Events        []flow.Event
}

// Succeeded returns true if the transaction executed without errors.
func (r TransactionResult) Succeeded() bool {
	return r.Error == nil
}

// Reverted returns true if the transaction executed with errors.
func (r TransactionResult) Reverted() bool {
	return !r.Succeeded()
}

func (r TransactionResult) ToStorableResult() types.StorableTransactionResult {
	var errorCode int
	var errorMessage string

	if r.Error != nil {
		errorCode = 1
		errorMessage = r.Error.Error()
	}

	return types.StorableTransactionResult{
		ErrorCode:    errorCode,
		ErrorMessage: errorMessage,
		Logs:         r.Logs,
		Events:       r.Events,
	}
}

// A ScriptResult is the result of executing a script.
type ScriptResult struct {
	ScriptID flow.Identifier
	Value    cadence.Value
	Error    error
	Logs     []string
	Events   []flow.Event
}

// Succeeded returns true if the script executed without errors.
func (r ScriptResult) Succeeded() bool {
	return r.Error == nil
}

// Reverted returns true if the script executed with errors.
func (r ScriptResult) Reverted() bool {
	return !r.Succeeded()
}
