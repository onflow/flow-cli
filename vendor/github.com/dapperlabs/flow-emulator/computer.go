package emulator

import (
	"errors"
	"fmt"

	"github.com/dapperlabs/cadence"
	encoding "github.com/dapperlabs/cadence/encoding/xdr"
	"github.com/dapperlabs/cadence/runtime"
	"github.com/dapperlabs/flow-go-sdk"
	"github.com/dapperlabs/flow-go/crypto"

	"github.com/dapperlabs/flow-emulator/execution"
	"github.com/dapperlabs/flow-emulator/types"
)

// A computer uses a runtime instance to execute transactions and scripts.
type computer struct {
	runtime        runtime.Runtime
	onEventEmitted func(event flow.Event, blockNumber uint64, txHash crypto.Hash)
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
func (c *computer) ExecuteTransaction(ledger *types.LedgerView, tx flow.Transaction) (TransactionResult, error) {
	runtimeContext := execution.NewRuntimeContext(ledger)

	runtimeContext.SetChecker(func(code []byte, location runtime.Location) error {
		return c.runtime.ParseAndCheckProgram(code, runtimeContext, location)
	})
	runtimeContext.SetSigningAccounts(tx.ScriptAccounts)

	location := runtime.TransactionLocation(tx.Hash())

	executionErr := c.runtime.ExecuteTransaction(tx.Script, runtimeContext, location)

	convertedEvents, err := convertEvents(runtimeContext.Events(), tx.Hash())
	if err != nil {
		return TransactionResult{}, err
	}

	if executionErr != nil {
		if errors.As(executionErr, &runtime.Error{}) {
			// runtime errors occur when the execution reverts
			return TransactionResult{
				TransactionHash: tx.Hash(),
				Error:           executionErr,
				Logs:            runtimeContext.Logs(),
				Events:          convertedEvents,
			}, nil
		}

		// other errors are unexpected and should be treated as fatal
		return TransactionResult{}, executionErr
	}

	return TransactionResult{
		TransactionHash: tx.Hash(),
		Error:           nil,
		Logs:            runtimeContext.Logs(),
		Events:          convertedEvents,
	}, nil
}

// ExecuteScript executes a plain script in the runtime.
//
// This function initializes a new runtime context using the provided registers view.
func (c *computer) ExecuteScript(view *types.LedgerView, script []byte) (ScriptResult, error) {
	runtimeContext := execution.NewRuntimeContext(view)

	hasher := crypto.NewSHA3_256()
	scriptHash := hasher.ComputeHash(script)

	location := runtime.ScriptLocation(scriptHash)

	value, executionErr := c.runtime.ExecuteScript(script, runtimeContext, location)

	convertedEvents, err := convertEvents(runtimeContext.Events(), nil)
	if err != nil {
		return ScriptResult{}, err
	}

	if executionErr != nil {
		if errors.As(executionErr, &runtime.Error{}) {
			// runtime errors occur when the execution reverts
			return ScriptResult{
				ScriptHash: scriptHash,
				Value:      nil,
				Error:      executionErr,
				Logs:       runtimeContext.Logs(),
				Events:     convertedEvents,
			}, nil
		}

		// other errors are unexpected and should be treated as fatal
		return ScriptResult{}, executionErr
	}

	convertedValue := cadence.ConvertValue(value)

	return ScriptResult{
		ScriptHash: scriptHash,
		Value:      convertedValue,
		Error:      nil,
		Logs:       runtimeContext.Logs(),
		Events:     convertedEvents,
	}, nil
}

func convertEvents(rtEvents []runtime.Event, txHash crypto.Hash) ([]flow.Event, error) {
	flowEvents := make([]flow.Event, len(rtEvents))

	for i, event := range rtEvents {
		eventValue := cadence.ConvertEvent(event)

		payload, err := encoding.Encode(eventValue)
		if err != nil {
			return nil, fmt.Errorf("failed to encode event: %w", err)
		}

		flowEvents[i] = flow.Event{
			Type:    string(event.Type.ID()),
			TxHash:  txHash,
			Index:   uint(i),
			Payload: payload,
		}
	}

	return flowEvents, nil
}
