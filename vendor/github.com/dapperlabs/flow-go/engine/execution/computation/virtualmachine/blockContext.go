package virtualmachine

import (
	"errors"
	"fmt"

	"github.com/dapperlabs/cadence/runtime"

	"github.com/dapperlabs/flow-go/model/flow"
	"github.com/dapperlabs/flow-go/model/hash"
)

// A BlockContext is used to execute transactions in the context of a block.
type BlockContext interface {
	// ExecuteTransaction computes the result of a transaction.
	ExecuteTransaction(
		ledger Ledger,
		tx *flow.TransactionBody,
		options ...TransactionContextOption,
	) (*TransactionResult, error)

	// ExecuteScript computes the result of a read-only script.
	ExecuteScript(ledger Ledger, script []byte) (*ScriptResult, error)
}

type blockContext struct {
	vm     *virtualMachine
	header *flow.Header
}

func (bc *blockContext) newTransactionContext(
	ledger Ledger,
	tx *flow.TransactionBody,
	options ...TransactionContextOption,
) *TransactionContext {

	signingAccounts := make([]runtime.Address, len(tx.ScriptAccounts))
	for i, addr := range tx.ScriptAccounts {
		signingAccounts[i] = runtime.Address(addr)
	}

	ctx := &TransactionContext{
		ledger:          ledger,
		signingAccounts: signingAccounts,
	}

	for _, option := range options {
		option(ctx)
	}

	return ctx
}

func (bc *blockContext) newScriptContext(ledger Ledger) *TransactionContext {
	return &TransactionContext{
		ledger: ledger,
	}
}

// ExecuteTransaction computes the result of a transaction.
//
// Register updates are recorded in the provided ledger view. An error is returned
// if an unexpected error occurs during execution. If the transaction reverts due to
// a normal runtime error, the error is recorded in the transaction result.
func (bc *blockContext) ExecuteTransaction(
	ledger Ledger,
	tx *flow.TransactionBody,
	options ...TransactionContextOption,
) (*TransactionResult, error) {

	txID := tx.ID()
	location := runtime.TransactionLocation(txID[:])

	ctx := bc.newTransactionContext(ledger, tx, options...)

	err := bc.vm.executeTransaction(tx.Script, ctx, location)
	if err != nil {
		if errors.As(err, &runtime.Error{}) {
			// runtime errors occur when the execution reverts
			return &TransactionResult{
				TransactionID: txID,
				Error:         err,
				Logs:          ctx.Logs(),
			}, nil
		}

		// other errors are unexpected and should be treated as fatal
		return nil, fmt.Errorf("failed to execute transaction: %w", err)
	}

	return &TransactionResult{
		TransactionID: txID,
		Error:         nil,
		Events:        ctx.Events(),
		Logs:          ctx.Logs(),
	}, nil
}

func (bc *blockContext) ExecuteScript(ledger Ledger, script []byte) (*ScriptResult, error) {
	scriptHash := hash.DefaultHasher.ComputeHash(script)

	location := runtime.ScriptLocation(scriptHash)

	ctx := bc.newScriptContext(ledger)

	value, err := bc.vm.executeScript(script, ctx, location)
	if err != nil {
		if errors.As(err, &runtime.Error{}) {
			// runtime errors occur when the execution reverts
			return &ScriptResult{
				ScriptID: flow.HashToID(scriptHash),
				Error:    err,
				Logs:     ctx.Logs(),
			}, nil
		}

		return nil, fmt.Errorf("failed to execute script: %w", err)
	}

	return &ScriptResult{
		ScriptID: flow.HashToID(scriptHash),
		Value:    value,
		Logs:     ctx.Logs(),
	}, nil
}
