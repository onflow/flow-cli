package execution

import (
	"errors"
	"fmt"
	"math/big"
	"strings"

	"github.com/onflow/cadence"
	"github.com/onflow/cadence/runtime"
	"github.com/onflow/flow-go-sdk"

	"github.com/dapperlabs/flow-emulator/types"
)

type CheckerFunc func([]byte, runtime.Location) error

// RuntimeContext implements host functionality required by the Cadence runtime.
//
// A context is short-lived and is intended to be used when executing a single transaction.
//
// The logic in this runtime context is specific to the emulator and is designed to be
// used with a Blockchain instance.
type RuntimeContext struct {
	ledger          *types.LedgerView
	signingAccounts []runtime.Address
	checker         CheckerFunc
	logs            []string
	events          []runtime.Event
}

// NewRuntimeContext returns a new RuntimeContext instance.
func NewRuntimeContext(ledger *types.LedgerView) *RuntimeContext {
	return &RuntimeContext{
		ledger:  ledger,
		checker: func([]byte, runtime.Location) error { return nil },
		events:  make([]runtime.Event, 0),
	}
}

// SetSigningAccounts sets the signing accounts for this context.
//
// Signing accounts are the accounts that signed the transaction executing
// inside this context.
func (r *RuntimeContext) SetSigningAccounts(accounts []flow.Address) {
	signingAccounts := make([]runtime.Address, len(accounts))

	for i, account := range accounts {
		signingAccounts[i] = runtime.Address(cadence.NewAddress(account))
	}

	r.signingAccounts = signingAccounts
}

// GetSigningAccounts gets the signing accounts for this context.
//
// Signing accounts are the accounts that signed the transaction executing
// inside this context.
func (r *RuntimeContext) GetSigningAccounts() []runtime.Address {
	return r.signingAccounts
}

// SetChecker sets the semantic checker function for this context.
func (r *RuntimeContext) SetChecker(checker CheckerFunc) {
	r.checker = checker
}

// Events returns all events emitted by the runtime to this context.
func (r *RuntimeContext) Events() []runtime.Event {
	return r.events
}

// Logs returns all logs emitted by the runtime to this context.
func (r *RuntimeContext) Logs() []string {
	return r.logs
}

// GetValue gets a register value from the world state.
func (r *RuntimeContext) GetValue(owner, controller, key []byte) ([]byte, error) {
	v, _ := r.ledger.Get(fullKey(string(owner), string(controller), string(key)))
	return v, nil
}

// SetValue sets a register value in the world state.
func (r *RuntimeContext) SetValue(owner, controller, key, value []byte) error {
	r.ledger.Set(fullKey(string(owner), string(controller), string(key)), value)
	return nil
}

// CreateAccount creates a new account and inserts it into the world state.
//
// This function returns an error if the input is invalid.
//
// After creating the account, this function calls the onAccountCreated callback registered
// with this context.
func (r *RuntimeContext) CreateAccount(publicKeys [][]byte) (runtime.Address, error) {
	latestAccountID, _ := r.ledger.Get(keyLatestAccount)

	accountIDInt := big.NewInt(0).SetBytes(latestAccountID)
	accountIDBytes := accountIDInt.Add(accountIDInt, big.NewInt(1)).Bytes()

	accountAddress := flow.BytesToAddress(accountIDBytes)

	accountID := accountAddress[:]

	// mark that account with this ID exists
	r.ledger.Set(fullKey(string(accountID), "", keyExists), []byte{1})

	r.ledger.Set(fullKey(string(accountID), "", keyBalance), big.NewInt(0).Bytes())

	accountKeys := make([]accountKey, len(publicKeys))
	for i, publicKey := range publicKeys {
		accountKeys[i] = accountKey{
			publicKey: publicKey,
			// initial sequence number is zero
			sequenceNumber: 0,
		}
	}

	err := r.setAccountKeys(accountID, accountKeys)
	if err != nil {
		return runtime.Address{}, err
	}

	r.ledger.Set(keyLatestAccount, accountID)

	r.Log("Creating new account\n")
	r.Log(fmt.Sprintf("Address: %s", accountAddress))

	return runtime.Address(cadence.NewAddress(accountAddress)), nil
}

// AddAccountKey adds a public key to an existing account.
//
// This function returns an error if the specified account does not exist or
// if the key insertion fails.
func (r *RuntimeContext) AddAccountKey(address runtime.Address, publicKey []byte) error {
	accountID := address[:]

	err := r.checkAccountExists(accountID)
	if err != nil {
		return err
	}

	accountKeys, err := r.getAccountKeys(accountID)
	if err != nil {
		return err
	}

	newAccountKey := accountKey{
		publicKey: publicKey,
		// initial sequence number is zero
		sequenceNumber: 0,
	}

	accountKeys = append(accountKeys, newAccountKey)

	return r.setAccountKeys(accountID, accountKeys)
}

// RemoveAccountKey removes a public key by index from an existing account.
//
// This function returns an error if the specified account does not exist, the
// provided key is invalid, or if key deletion fails.
func (r *RuntimeContext) RemoveAccountKey(address runtime.Address, index int) (publicKey []byte, err error) {
	accountID := address[:]

	err = r.checkAccountExists(accountID)
	if err != nil {
		return nil, err
	}

	accountKeys, err := r.getAccountKeys(accountID)
	if err != nil {
		return publicKey, err
	}

	if index < 0 || index > len(accountKeys)-1 {
		return publicKey, fmt.Errorf("invalid key index %d, account has %d keys", index, len(accountKeys))
	}

	removedKey := accountKeys[index]

	// remove key from list
	accountKeys = append(accountKeys[:index], accountKeys[index+1:]...)

	err = r.setAccountKeys(accountID, accountKeys)
	if err != nil {
		return publicKey, err
	}

	return removedKey.publicKey, nil
}

type accountKey struct {
	publicKey      []byte
	sequenceNumber uint64
}

func (r *RuntimeContext) getAccountKeys(accountID []byte) (accountKeys []accountKey, err error) {
	countBytes, err := r.ledger.Get(
		fullKey(string(accountID), string(accountID), keyPublicKeyCount),
	)
	if err != nil {
		return nil, err
	}

	if countBytes == nil {
		return nil, fmt.Errorf("key count not set")
	}

	count := int(big.NewInt(0).SetBytes(countBytes).Int64())

	accountKeys = make([]accountKey, count)

	for i := 0; i < count; i++ {
		publicKey, err := r.ledger.Get(
			fullKey(string(accountID), string(accountID), keyPublicKey(i)),
		)
		if err != nil {
			return nil, err
		}

		if publicKey == nil {
			return nil, fmt.Errorf("failed to retrieve key from account %s", accountID)
		}

		seqNumBytes, err := r.ledger.Get(
			fullKey(string(accountID), string(accountID), keyPublicKeySequenceNumber(i)),
		)
		if err != nil {
			return nil, err
		}

		seqNum := big.NewInt(0).SetBytes(seqNumBytes).Uint64()

		accountKeys[i] = accountKey{
			publicKey:      publicKey,
			sequenceNumber: seqNum,
		}
	}

	return accountKeys, nil
}

func (r *RuntimeContext) setAccountKeys(accountID []byte, accountKeys []accountKey) error {
	var existingCount int

	countBytes, err := r.ledger.Get(
		fullKey(string(accountID), string(accountID), keyPublicKeyCount),
	)
	if err != nil {
		return err
	}

	if countBytes != nil {
		existingCount = int(big.NewInt(0).SetBytes(countBytes).Int64())
	} else {
		existingCount = 0
	}

	newCount := len(accountKeys)

	r.ledger.Set(
		fullKey(string(accountID), string(accountID), keyPublicKeyCount),
		big.NewInt(int64(newCount)).Bytes(),
	)

	for i, accountKey := range accountKeys {
		accountPublicKey, err := flow.DecodeAccountKey(accountKey.publicKey)
		if err != nil {
			return err
		}

		err = accountPublicKey.Validate()
		if err != nil {
			return err
		}

		r.ledger.Set(
			fullKey(string(accountID), string(accountID), keyPublicKey(i)),
			accountKey.publicKey,
		)

		seqNumBytes := big.NewInt(0).SetUint64(accountKey.sequenceNumber).Bytes()

		r.ledger.Set(
			fullKey(string(accountID), string(accountID), keyPublicKeySequenceNumber(i)),
			seqNumBytes,
		)
	}

	// delete leftover keys
	for i := newCount; i < existingCount; i++ {
		r.ledger.Delete(fullKey(string(accountID), string(accountID), keyPublicKey(i)))
		r.ledger.Delete(fullKey(string(accountID), string(accountID), keyPublicKeySequenceNumber(i)))
	}

	return nil
}

// CheckCode checks the code for its validity.
func (r *RuntimeContext) CheckCode(address runtime.Address, code []byte) (err error) {
	return r.checkProgram(code, address)
}

// UpdateAccountCode updates the deployed code on an existing account.
//
// This function returns an error if the specified account does not exist or is
// not a valid signing account.
func (r *RuntimeContext) UpdateAccountCode(address runtime.Address, code []byte, checkPermission bool) (err error) {
	accountID := address[:]

	if checkPermission && !r.isValidSigningAccount(address) {
		return fmt.Errorf("not permitted to update account with ID %s", accountID)
	}

	err = r.checkAccountExists(accountID)
	if err != nil {
		return err
	}

	r.ledger.Set(fullKey(string(accountID), string(accountID), keyCode), code)

	return nil
}

// GetAccount gets an account by address.
//
// The function returns nil if the specified account does not exist.
func (r *RuntimeContext) GetAccount(address flow.Address) *flow.Account {
	accountID := address.Bytes()

	err := r.checkAccountExists(accountID)
	if err != nil {
		return nil
	}

	balanceBytes, _ := r.ledger.Get(fullKey(string(accountID), "", keyBalance))
	balanceInt := big.NewInt(0).SetBytes(balanceBytes)

	code, _ := r.ledger.Get(fullKey(string(accountID), string(accountID), keyCode))

	accountKeys, err := r.getAccountKeys(accountID)
	if err != nil {
		panic(err)
	}

	accountPublicKeys := make([]*flow.AccountKey, len(accountKeys))
	for i, accountKey := range accountKeys {
		accountPublicKey, err := flow.DecodeAccountKey(accountKey.publicKey)
		if err != nil {
			panic(err)
		}

		// include sequence number
		accountPublicKey.SequenceNumber = accountKey.sequenceNumber
		accountPublicKey.ID = i

		accountPublicKeys[i] = accountPublicKey
	}

	return &flow.Account{
		Address: address,
		Balance: balanceInt.Uint64(),
		Code:    code,
		Keys:    accountPublicKeys,
	}
}

func (r *RuntimeContext) checkAccountExists(accountID []byte) error {
	exists, err := r.ledger.Get(fullKey(string(accountID), "", keyExists))
	if err != nil {
		return err
	}

	if len(exists) == 0 {
		return fmt.Errorf("account with ID %s does not exist", accountID)
	}

	return nil
}

// CheckAndIncrementSequenceNumber validates and increments a sequence number for with an account key.
//
// This function first checks that the provided sequence number matches the version stored on-chain.
// If they are equal, the on-chain sequence number is incremented.
// If they are not equal, the on-chain sequence number is not incremented.
//
// This function returns a boolean flag indicating validity as well as the updated sequence number value.
// This function returns an error if the sequence number cannot be read from storage.
func (r *RuntimeContext) CheckAndIncrementSequenceNumber(
	address flow.Address,
	keyID int,
	sequenceNumber uint64,
) (bool, uint64, error) {
	accountID := address.Bytes()

	seqNumKey := keyPublicKeySequenceNumber(keyID)

	storedSeqNumBytes, err := r.ledger.Get(
		fullKey(string(accountID), string(accountID), seqNumKey),
	)
	if err != nil {
		return false, 0, err
	}

	storedSeqNum := big.NewInt(0).SetBytes(storedSeqNumBytes).Uint64()

	valid := storedSeqNum == sequenceNumber

	if valid {
		storedSeqNum++
		storedSeqNumBytes = big.NewInt(0).SetUint64(storedSeqNum).Bytes()

		r.ledger.Set(
			fullKey(string(accountID), string(accountID), seqNumKey),
			storedSeqNumBytes,
		)
	}

	return valid, storedSeqNum, nil
}

// ResolveImport imports code for the provided import location.
//
// This function returns an error if the import location is not an account address,
// or if there is no code deployed at the specified address.
func (r *RuntimeContext) ResolveImport(location runtime.Location) ([]byte, error) {
	addressLocation, ok := location.(runtime.AddressLocation)
	if !ok {
		return nil, errors.New("import location must be an account address")
	}

	address := flow.BytesToAddress(addressLocation)

	accountID := address.Bytes()

	code, err := r.ledger.Get(fullKey(string(accountID), string(accountID), keyCode))
	if err != nil {
		return nil, err
	}

	if code == nil {
		return nil, fmt.Errorf("no code deployed at address %s", accountID)
	}

	return code, nil
}

// Log captures a log message from the runtime.
func (r *RuntimeContext) Log(message string) {
	r.logs = append(r.logs, message)
}

// EmitEvent is called when an event is emitted by the runtime.
func (r *RuntimeContext) EmitEvent(event runtime.Event) {
	r.events = append(r.events, event)
}

func (r *RuntimeContext) isValidSigningAccount(address runtime.Address) bool {
	for _, accountAddress := range r.GetSigningAccounts() {
		if accountAddress == address {
			return true
		}
	}

	return false
}

// checkProgram checks the given code for syntactic and semantic correctness.
func (r *RuntimeContext) checkProgram(code []byte, address runtime.Address) error {
	if code == nil {
		return nil
	}

	location := runtime.AddressLocation(address[:])

	return r.checker(code, location)
}

const (
	keyLatestAccount  = "latest_account"
	keyExists         = "exists"
	keyBalance        = "balance"
	keyCode           = "code"
	keyPublicKeyCount = "public_key_count"
)

func fullKey(owner, controller, key string) string {
	return strings.Join([]string{owner, controller, key}, "__")
}

func keyPublicKey(index int) string {
	return fmt.Sprintf("public_key_%d", index)
}

func keyPublicKeySequenceNumber(index int) string {
	return fmt.Sprintf("public_key_%d_seq_num", index)
}
