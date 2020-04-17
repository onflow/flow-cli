package emulator

import (
	"fmt"
	"strings"

	"github.com/dapperlabs/flow-go-sdk"
	"github.com/dapperlabs/flow-go/crypto"
)

type ErrNotFound interface {
	isNotFoundError()
}

type ErrBlockNotFound interface {
	isBlockNotFoundError()
}

// ErrBlockNotFoundByHeight indicates that a block could not be found at the specified height.
type ErrBlockNotFoundByHeight struct {
	Height uint64
}

func (e *ErrBlockNotFoundByHeight) isNotFoundError()      {}
func (e *ErrBlockNotFoundByHeight) isBlockNotFoundError() {}

func (e *ErrBlockNotFoundByHeight) Error() string {
	return fmt.Sprintf("could not find block at height %d", e.Height)
}

// ErrBlockNotFoundByID indicates that a block with the specified ID could not be found.
type ErrBlockNotFoundByID struct {
	ID flow.Identifier
}

func (e *ErrBlockNotFoundByID) isNotFoundError()      {}
func (e *ErrBlockNotFoundByID) isBlockNotFoundError() {}

func (e *ErrBlockNotFoundByID) Error() string {
	return fmt.Sprintf("could not find block with ID %s", e.ID)
}

// ErrCollectionNotFound indicates that a collection could not be found.
type ErrCollectionNotFound struct {
	ID flow.Identifier
}

func (e *ErrCollectionNotFound) isNotFoundError() {}

func (e *ErrCollectionNotFound) Error() string {
	return fmt.Sprintf("could not find collection with ID %s", e.ID)
}

// ErrTransactionNotFound indicates that a transaction could not be found.
type ErrTransactionNotFound struct {
	ID flow.Identifier
}

func (e *ErrTransactionNotFound) isNotFoundError() {}

func (e *ErrTransactionNotFound) Error() string {
	return fmt.Sprintf("could not find transaction with ID %s", e.ID)
}

// ErrAccountNotFound indicates that an account could not be found.
type ErrAccountNotFound struct {
	Address flow.Address
}

func (e *ErrAccountNotFound) isNotFoundError() {}

func (e *ErrAccountNotFound) Error() string {
	return fmt.Sprintf("could not find account with address %s", e.Address)
}

// ErrDuplicateTransaction indicates that a transaction has already been submitted.
type ErrDuplicateTransaction struct {
	TxID flow.Identifier
}

func (e *ErrDuplicateTransaction) Error() string {
	return fmt.Sprintf("Transaction with ID %s has already been submitted", e.TxID)
}

// ErrMissingSignature indicates that a transaction is missing a required signature.
type ErrMissingSignature struct {
	Account flow.Address
}

func (e *ErrMissingSignature) Error() string {
	return fmt.Sprintf("Account %s does not have sufficient signatures", e.Account)
}

// ErrInvalidSignaturePublicKey indicates that signature uses an invalid public key.
type ErrInvalidSignaturePublicKey struct {
	Account flow.Address
	KeyID   int
}

func (e *ErrInvalidSignaturePublicKey) Error() string {
	return fmt.Sprintf("invalid signature for key %d on account %s", e.KeyID, e.Account)
}

// ErrInvalidSignatureAccount indicates that a signature references a nonexistent account.
type ErrInvalidSignatureAccount struct {
	Address flow.Address
}

func (e *ErrInvalidSignatureAccount) Error() string {
	return fmt.Sprintf("Account with address %s does not exist", e.Address)
}

// ErrInvalidTransaction indicates that a submitted transaction is invalid (missing required fields).
type ErrInvalidTransaction struct {
	TxID          flow.Identifier
	MissingFields []string
}

func (e *ErrInvalidTransaction) Error() string {
	return fmt.Sprintf(
		"Transaction with ID %s is invalid (missing required fields): %s",
		e.TxID,
		strings.Join(e.MissingFields, ", "),
	)
}

// ErrInvalidStateVersion indicates that a state version hash provided is invalid.
type ErrInvalidStateVersion struct {
	Version crypto.Hash
}

func (e *ErrInvalidStateVersion) Error() string {
	return fmt.Sprintf("Execution state with version hash %x is invalid", e.Version)
}

// ErrPendingBlockCommitBeforeExecution indicates that the current pending block has not been executed (cannot commit).
type ErrPendingBlockCommitBeforeExecution struct {
	BlockID flow.Identifier
}

func (e *ErrPendingBlockCommitBeforeExecution) Error() string {
	return fmt.Sprintf("Pending block with ID %s cannot be commited before execution", e.BlockID)
}

// ErrPendingBlockMidExecution indicates that the current pending block is mid-execution.
type ErrPendingBlockMidExecution struct {
	BlockID flow.Identifier
}

func (e *ErrPendingBlockMidExecution) Error() string {
	return fmt.Sprintf("Pending block with ID %s is currently being executed", e.BlockID)
}

// ErrPendingBlockTransactionsExhausted indicates that the current pending block has finished executing (no more transactions to execute).
type ErrPendingBlockTransactionsExhausted struct {
	BlockID flow.Identifier
}

func (e *ErrPendingBlockTransactionsExhausted) Error() string {
	return fmt.Sprintf("Pending block with ID %s contains no more transactions to execute", e.BlockID)
}

// ErrStorage indicates that an error occurred in the storage provider.
type ErrStorage struct {
	inner error
}

func (e *ErrStorage) Error() string {
	return fmt.Sprintf("storage failure: %v", e.inner)
}

func (e *ErrStorage) Unwrap() error {
	return e.inner
}
