package types

import (
	flowgo "github.com/onflow/flow-go/model/flow"
	"github.com/onflow/cadence"
	"github.com/onflow/flow-go-sdk"
)

type StorableTransactionResult struct {
	ErrorCode    int
	ErrorMessage string
	Logs         []string
	Events       []flowgo.Event
}

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

// TODO - this class should be part of SDK for consistency

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
