package flow

import (
	"github.com/dapperlabs/flow-go/crypto"
)

type Spock []byte

type ExecutionReceipt struct {
	ExecutorID        Identifier
	ExecutionResult   ExecutionResult
	Spocks            []crypto.Signature
	ExecutorSignature crypto.Signature
}

// Body returns the body of the execution receipt.
func (er *ExecutionReceipt) Body() interface{} {
	return struct {
		ExecutorID      Identifier
		ExecutionResult ExecutionResult
		Spocks          []crypto.Signature
	}{
		ExecutorID:      er.ExecutorID,
		ExecutionResult: er.ExecutionResult,
		Spocks:          er.Spocks,
	}
}

// ID returns the canonical ID of the execution receipt.
func (er *ExecutionReceipt) ID() Identifier {
	return MakeID(er.Body())
}

// Checksum returns a checksum for the execution receipt including the signatures.
func (er *ExecutionReceipt) Checksum() Identifier {
	return MakeID(er)
}
