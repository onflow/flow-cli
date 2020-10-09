package emulator

import (
	"fmt"

	"github.com/onflow/flow-go/access"
	flowgo "github.com/onflow/flow-go/model/flow"
	"github.com/onflow/flow-go-sdk"
	"github.com/onflow/flow-go-sdk/crypto"
)

// A NotFoundError indicates that an entity could not be found.
type NotFoundError interface {
	isNotFoundError()
}

// A BlockNotFoundError indicates that a block could not be found.
type BlockNotFoundError interface {
	isBlockNotFoundError()
}

// A BlockNotFoundByHeightError indicates that a block could not be found at the specified height.
type BlockNotFoundByHeightError struct {
	Height uint64
}

func (e *BlockNotFoundByHeightError) isNotFoundError()      {}
func (e *BlockNotFoundByHeightError) isBlockNotFoundError() {}

func (e *BlockNotFoundByHeightError) Error() string {
	return fmt.Sprintf("could not find block at height %d", e.Height)
}

// A BlockNotFoundByIDError indicates that a block with the specified ID could not be found.
type BlockNotFoundByIDError struct {
	ID flow.Identifier
}

func (e *BlockNotFoundByIDError) isNotFoundError()      {}
func (e *BlockNotFoundByIDError) isBlockNotFoundError() {}

func (e *BlockNotFoundByIDError) Error() string {
	return fmt.Sprintf("could not find block with ID %s", e.ID)
}

// A CollectionNotFoundError indicates that a collection could not be found.
type CollectionNotFoundError struct {
	ID flow.Identifier
}

func (e *CollectionNotFoundError) isNotFoundError() {}

func (e *CollectionNotFoundError) Error() string {
	return fmt.Sprintf("could not find collection with ID %s", e.ID)
}

// A TransactionNotFoundError indicates that a transaction could not be found.
type TransactionNotFoundError struct {
	ID flowgo.Identifier
}

func (e *TransactionNotFoundError) isNotFoundError() {}

func (e *TransactionNotFoundError) Error() string {
	return fmt.Sprintf("could not find transaction with ID %s", e.ID)
}

// An AccountNotFoundError indicates that an account could not be found.
type AccountNotFoundError struct {
	Address flowgo.Address
}

func (e *AccountNotFoundError) isNotFoundError() {}

func (e *AccountNotFoundError) Error() string {
	return fmt.Sprintf("could not find account with address %s", e.Address)
}

// A TransactionValidationError indicates that a submitted transaction is invalid.
type TransactionValidationError interface {
	isTransactionValidationError()
}

// A DuplicateTransactionError indicates that a transaction has already been submitted.
type DuplicateTransactionError struct {
	TxID flowgo.Identifier
}

func (e *DuplicateTransactionError) isTransactionValidationError() {}

func (e *DuplicateTransactionError) Error() string {
	return fmt.Sprintf("transaction with ID %s has already been submitted", e.TxID)
}

// IncompleteTransactionError indicates that a transaction is missing one or more required fields.
type IncompleteTransactionError struct {
	MissingFields []string
}

func (e *IncompleteTransactionError) isTransactionValidationError() {}

func (e *IncompleteTransactionError) Error() string {
	return fmt.Sprintf("transaction is missing required fields: %s", e.MissingFields)
}

// ExpiredTransactionError indicates that a transaction has expired.
type ExpiredTransactionError struct {
	RefHeight, FinalHeight uint64
}

func (e *ExpiredTransactionError) isTransactionValidationError() {}

func (e *ExpiredTransactionError) Error() string {
	return fmt.Sprintf("transaction is expired: ref_height=%d final_height=%d", e.RefHeight, e.FinalHeight)
}

// InvalidTransactionScriptError indicates that a transaction contains an invalid Cadence script.
type InvalidTransactionScriptError struct {
	ParserErr error
}

func (e *InvalidTransactionScriptError) isTransactionValidationError() {}

func (e *InvalidTransactionScriptError) Error() string {
	return fmt.Sprintf("failed to parse transaction Cadence script: %s", e.ParserErr)
}

func (e *InvalidTransactionScriptError) Unwrap() error {
	return e.ParserErr
}

// InvalidTransactionGasLimitError indicates that a transaction specifies a gas limit that exceeds the maximum.
type InvalidTransactionGasLimitError struct {
	Maximum uint64
	Actual  uint64
}

func (e *InvalidTransactionGasLimitError) isTransactionValidationError() {}

func (e *InvalidTransactionGasLimitError) Error() string {
	return fmt.Sprintf("transaction gas limit (%d) exceeds the maximum gas limit (%d)", e.Actual, e.Maximum)
}

// An InvalidStateVersionError indicates that a state version hash provided is invalid.
type InvalidStateVersionError struct {
	Version crypto.Hash
}

func (e *InvalidStateVersionError) Error() string {
	return fmt.Sprintf("execution state with version hash %x is invalid", e.Version)
}

// A PendingBlockCommitBeforeExecutionError indicates that the current pending block has not been executed (cannot commit).
type PendingBlockCommitBeforeExecutionError struct {
	BlockID flowgo.Identifier
}

func (e *PendingBlockCommitBeforeExecutionError) Error() string {
	return fmt.Sprintf("pending block with ID %s cannot be commited before execution", e.BlockID)
}

// A PendingBlockMidExecutionError indicates that the current pending block is mid-execution.
type PendingBlockMidExecutionError struct {
	BlockID flowgo.Identifier
}

func (e *PendingBlockMidExecutionError) Error() string {
	return fmt.Sprintf("pending block with ID %s is currently being executed", e.BlockID)
}

// A PendingBlockTransactionsExhaustedError indicates that the current pending block has finished executing (no more transactions to execute).
type PendingBlockTransactionsExhaustedError struct {
	BlockID flowgo.Identifier
}

func (e *PendingBlockTransactionsExhaustedError) Error() string {
	return fmt.Sprintf("pending block with ID %s contains no more transactions to execute", e.BlockID)
}

// A StorageError indicates that an error occurred in the storage provider.
type StorageError struct {
	inner error
}

func (e *StorageError) Error() string {
	return fmt.Sprintf("storage failure: %v", e.inner)
}

func (e *StorageError) Unwrap() error {
	return e.inner
}

// An ExecutionError occurs when a transaction fails to execute.
type ExecutionError struct {
	Code    int
	Message string
}

func (e *ExecutionError) Error() string {
	return fmt.Sprintf("execution error code %d: %s", e.Code, e.Message)
}

func convertAccessError(err error) error {
	switch typedErr := err.(type) {
	case access.IncompleteTransactionError:
		return &IncompleteTransactionError{MissingFields: typedErr.MissingFields}
	case access.ExpiredTransactionError:
		return &ExpiredTransactionError{RefHeight: typedErr.RefHeight, FinalHeight: typedErr.FinalHeight}
	case access.InvalidGasLimitError:
		return &InvalidTransactionGasLimitError{Maximum: typedErr.Maximum, Actual: typedErr.Actual}
	case access.InvalidScriptError:
		return &InvalidTransactionScriptError{ParserErr: typedErr.ParserErr}
	}

	return err
}
