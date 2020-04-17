package emulator

import (
	"errors"
	"fmt"

	"github.com/dapperlabs/cadence"
	"github.com/dapperlabs/cadence/runtime"
	"github.com/dapperlabs/flow-go-sdk"
	"github.com/dapperlabs/flow-go/crypto"

	"github.com/dapperlabs/flow-emulator/execution"
	"github.com/dapperlabs/flow-emulator/types"
)

// A computer uses a runtime instance to execute transactions and scripts.
type computer struct {
	runtime        runtime.Runtime
	onEventEmitted func(event flow.Event, blockHeight uint64, txID flow.Identifier)
}

// newComputer returns a new computer initialized with a runtime.
func newComputer(
	runtime runtime.Runtime,
) *computer {
	return &computer{
		runtime: runtime,
	}
}

// ExecuteTransaction executes the provided transaction in the runtime.
//
// This function initializes a new runtime context using the provided ledger view, as well as
// the accounts that authorized the transaction.
//
// An error is returned if the transaction script cannot be parsed or reverts during execution.
func (c *computer) ExecuteTransaction(ledger *types.LedgerView, tx flow.Transaction) (*TransactionResult, error) {
	runtimeContext := execution.NewRuntimeContext(ledger)

	if tx.ProposalKey == (flow.ProposalKey{}) {
		// TODO: add dedicated error type
		return nil, fmt.Errorf("missing sequence number")
	}

	valid, updatedSeqNum, err := runtimeContext.CheckAndIncrementSequenceNumber(
		tx.ProposalKey.Address,
		tx.ProposalKey.KeyID,
		tx.ProposalKey.SequenceNumber,
	)
	if err != nil {
		return nil, err
	}

	if !valid {
		return &TransactionResult{
			TransactionID: tx.ID(),
			// TODO: add dedicated error type
			Error: fmt.Errorf(
				"invalid sequence number: expected %d, got %d",
				updatedSeqNum,
				tx.ProposalKey.SequenceNumber,
			),
			Logs:   nil,
			Events: nil,
		}, nil
	}

	runtimeContext.SetChecker(func(code []byte, location runtime.Location) error {
		return c.runtime.ParseAndCheckProgram(code, runtimeContext, location)
	})

	signers := make([]flow.Address, len(tx.Authorizers))
	for i, addr := range tx.Authorizers {
		signers[i] = addr
	}

	runtimeContext.SetSigningAccounts(signers)

	location := runtime.TransactionLocation(tx.ID().Bytes())

	executionErr := c.runtime.ExecuteTransaction(tx.Script, runtimeContext, location)

	convertedEvents, err := convertEvents(runtimeContext.Events(), tx.ID())
	if err != nil {
		return nil, err
	}

	if executionErr != nil {
		if errors.As(executionErr, &runtime.Error{}) {
			// runtime errors occur when the execution reverts
			return &TransactionResult{
				TransactionID: tx.ID(),
				Error:         executionErr,
				Logs:          runtimeContext.Logs(),
				Events:        convertedEvents,
			}, nil
		}

		// other errors are unexpected and should be treated as fatal
		return nil, executionErr
	}

	return &TransactionResult{
		TransactionID: tx.ID(),
		Error:         nil,
		Logs:          runtimeContext.Logs(),
		Events:        convertedEvents,
	}, nil
}

// ExecuteScript executes a plain script in the runtime.
//
// This function initializes a new runtime context using the provided registers view.
func (c *computer) ExecuteScript(view *types.LedgerView, script []byte) (*ScriptResult, error) {
	runtimeContext := execution.NewRuntimeContext(view)

	hasher := crypto.NewSHA3_256()
	scriptID := flow.HashToID(hasher.ComputeHash(script))

	location := runtime.ScriptLocation(scriptID.Bytes())

	value, executionErr := c.runtime.ExecuteScript(script, runtimeContext, location)

	convertedEvents, err := convertEvents(runtimeContext.Events(), flow.ZeroID)
	if err != nil {
		return nil, err
	}

	if executionErr != nil {
		if errors.As(executionErr, &runtime.Error{}) {
			// runtime errors occur when the execution reverts
			return &ScriptResult{
				ScriptID: scriptID,
				Value:    nil,
				Error:    executionErr,
				Logs:     runtimeContext.Logs(),
				Events:   convertedEvents,
			}, nil
		}

		// other errors are unexpected and should be treated as fatal
		return nil, executionErr
	}

	convertedValue := cadence.ConvertValue(value)

	return &ScriptResult{
		ScriptID: scriptID,
		Value:    convertedValue,
		Error:    nil,
		Logs:     runtimeContext.Logs(),
		Events:   convertedEvents,
	}, nil
}

func convertEvents(events []runtime.Event, txID flow.Identifier) ([]flow.Event, error) {
	flowEvents := make([]flow.Event, len(events))

	for i, event := range events {
		flowEvents[i] = flow.Event{
			Type:          string(event.Type.ID()),
			TransactionID: txID,
			// TODO: include transaction index field
			// TransactionIndex: txIndex,
			EventIndex: i,
			Value:      cadence.ConvertEvent(event),
		}
	}

	return flowEvents, nil
}
