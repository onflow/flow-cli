package virtualmachine

import (
	"crypto/sha256"
	"fmt"
	"math/big"
	"strings"

	"github.com/dapperlabs/flow-go/model/flow"
)

// A Ledger is the storage interface used by the virtual machine to read and write register values.
type Ledger interface {
	Set(key flow.RegisterID, value flow.RegisterValue)
	Get(key flow.RegisterID) (flow.RegisterValue, error)
	Delete(key flow.RegisterID)
}

// A MapLedger is a naive ledger storage implementation backed by a simple map.
//
// This implementation is designed for testing purposes.
type MapLedger map[string]flow.RegisterValue

func (m MapLedger) Set(key flow.RegisterID, value flow.RegisterValue) {
	m[string(key)] = value
}

func (m MapLedger) Get(key flow.RegisterID) (flow.RegisterValue, error) {
	return m[string(key)], nil
}

func (m MapLedger) Delete(key flow.RegisterID) {
	delete(m, string(key))
}

const (
	keyAddressState   = "account_address_state"
	keyExists         = "exists"
	keyCode           = "code"
	keyPublicKeyCount = "public_key_count"
)

func fullKey(owner, controller, key string) string {
	// https://en.wikipedia.org/wiki/C0_and_C1_control_codes#Field_separators
	return strings.Join([]string{owner, controller, key}, "\x1F")
}
func fullKeyHash(owner, controller, key string) flow.RegisterID {
	h := sha256.New()
	_, _ = h.Write([]byte(fullKey(owner, controller, key)))
	return h.Sum(nil)
}

func keyPublicKey(index uint64) string {
	return fmt.Sprintf("public_key_%d", index)
}

// A LedgerDAL is an abstraction layer used to read and manipulate ledger state in a consistent way.
type LedgerDAL struct {
	Ledger
}

func NewLedgerDAL(ledger Ledger) LedgerDAL {
	return LedgerDAL{Ledger: ledger}
}

func (r *LedgerDAL) CheckAccountExists(accountAddress []byte) error {
	exists, err := r.Get(fullKeyHash(string(accountAddress), "", keyExists))
	if err != nil {
		return err
	}

	if len(exists) != 0 {
		return nil
	}

	return fmt.Errorf("account with ID %x does not exist", accountAddress)
}

func (r *LedgerDAL) GetAccountPublicKeys(accountAddress []byte) (publicKeys []flow.AccountPublicKey, err error) {
	countBytes, err := r.Get(
		fullKeyHash(string(accountAddress), string(accountAddress), keyPublicKeyCount),
	)
	if err != nil {
		return nil, err
	}

	if countBytes == nil {
		return nil, fmt.Errorf("key count not set")
	}

	countInt := new(big.Int).SetBytes(countBytes)
	if !countInt.IsUint64() {
		return nil, fmt.Errorf(
			"retrieved public key account count bytes (hex-encoded): %x do not represent valid uint64",
			countBytes,
		)
	}
	count := countInt.Uint64()

	publicKeys = make([]flow.AccountPublicKey, count)

	for i := uint64(0); i < count; i++ {
		publicKey, err := r.Get(
			fullKeyHash(string(accountAddress), string(accountAddress), keyPublicKey(i)),
		)
		if err != nil {
			return nil, err
		}

		if publicKey == nil {
			return nil, fmt.Errorf("failed to retrieve key from account %s", accountAddress)
		}

		decodedPublicKey, err := flow.DecodeAccountPublicKey(publicKey)
		if err != nil {
			return nil, fmt.Errorf("failed to decode account public key: %w", err)
		}

		publicKeys[i] = decodedPublicKey
	}

	return publicKeys, nil
}

func (r *LedgerDAL) GetAccount(address flow.Address) *flow.Account {
	accountAddress := address.Bytes()

	err := r.CheckAccountExists(accountAddress)
	if err != nil {
		return nil
	}

	code, _ := r.Get(fullKeyHash(string(accountAddress), string(accountAddress), keyCode))

	publicKeys, err := r.GetAccountPublicKeys(accountAddress)
	if err != nil {
		panic(err)
	}

	return &flow.Account{
		Address: address,
		Code:    code,
		Keys:    publicKeys,
	}
}

func (r *LedgerDAL) GetAddressState() (*flow.AddressGenerator, error) {
	stateBytes, err := r.Get(fullKeyHash("", "", keyAddressState))
	if err != nil {
		return nil, err
	}
	state := flow.BytesToAddressState(stateBytes)
	return state, nil
}

func (r *LedgerDAL) SetAddressState(state *flow.AddressGenerator) {
	stateBytes := state.Bytes()
	r.Set(fullKeyHash("", "", keyAddressState), stateBytes)
}

func (r *LedgerDAL) CreateAccount(publicKeys []flow.AccountPublicKey) (flow.Address, error) {
	addressState, err := r.GetAddressState()
	if err != nil {
		return flow.Address{}, err
	}
	// generate the new account address
	newAddress, err := addressState.NextAddress()
	if err != nil {
		return flow.EmptyAddress, err
	}

	err = r.CreateAccountWithAddress(newAddress, publicKeys)
	if err != nil {
		return flow.Address{}, err
	}

	// update the address state
	r.SetAddressState(addressState)

	return newAddress, nil
}

func (r *LedgerDAL) CreateAccountWithAddress(
	addr flow.Address,
	publicKeys []flow.AccountPublicKey,
) error {
	accountID := addr.Bytes()

	// mark that account with this ID exists
	r.Set(fullKeyHash(string(accountID), "", keyExists), []byte{1})

	r.Set(fullKeyHash(string(accountID), string(accountID), keyCode), nil)

	err := r.SetAccountPublicKeys(accountID, publicKeys)
	if err != nil {
		return err
	}

	return nil
}

func (r *LedgerDAL) SetAccountPublicKeys(accountAddress []byte, publicKeys []flow.AccountPublicKey) error {

	var existingCount uint64

	countBytes, err := r.Get(
		fullKeyHash(string(accountAddress), string(accountAddress), keyPublicKeyCount),
	)
	if err != nil {
		return err
	}

	if countBytes != nil {
		countInt := new(big.Int).SetBytes(countBytes)
		if !countInt.IsUint64() {
			return fmt.Errorf(
				"retrieved public key account bytes (hex): %x do not represent valid uint64",
				countBytes,
			)
		}
		existingCount = countInt.Uint64()
	} else {
		existingCount = 0
	}

	newCount := uint64(len(publicKeys)) //len returns int and this won't exceed uint64
	newKeyCount := new(big.Int).SetUint64(newCount)

	r.Set(
		fullKeyHash(string(accountAddress), string(accountAddress), keyPublicKeyCount),
		newKeyCount.Bytes(),
	)

	for i, publicKey := range publicKeys {

		err = publicKey.Validate()
		if err != nil {
			return err
		}

		publicKeyBytes, err := flow.EncodeAccountPublicKey(publicKey)
		if err != nil {
			return fmt.Errorf("cannot encode account public key: %w", err)
		}

		// asserted length of publicKeys so i should always fit into uint64
		r.setAccountPublicKey(accountAddress, uint64(i), publicKeyBytes)
	}

	// delete leftover keys
	for i := newCount; i < existingCount; i++ {
		r.Delete(fullKeyHash(string(accountAddress), string(accountAddress), keyPublicKey(i)))
	}

	return nil
}

func (r *LedgerDAL) setAccountPublicKey(accountAddress []byte, keyID uint64, publicKey []byte) {
	r.Set(
		fullKeyHash(string(accountAddress), string(accountAddress), keyPublicKey(keyID)),
		publicKey,
	)
}
