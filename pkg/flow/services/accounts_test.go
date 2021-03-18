package services

import (
	"strings"
	"testing"

	"github.com/onflow/flow-cli/pkg/flow"

	"github.com/onflow/flow-go-sdk/crypto"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/onflow/flow-cli/pkg/flow/util"
	flowsdk "github.com/onflow/flow-go-sdk"

	"github.com/onflow/flow-cli/tests"
)

const (
	serviceAddress = "f8d6e0586b0a20c7"
	serviceName    = "emulator-account"
	pubKey         = "858a7d978b25d61f348841a343f79131f4b9fab341dd8a476a6f4367c25510570bf69b795fc9c3d2b7191327d869bcf848508526a3c1cafd1af34f71c7765117"
	sigAlgo        = "ECDSA_P256"
	hashAlgo       = "SHA3_256"
)

func TestAccounts(t *testing.T) {

	mock := &tests.MockGateway{}

	project, err := flow.InitProject(crypto.ECDSA_P256, crypto.SHA3_256)
	require.NoError(t, err)

	accounts := NewAccounts(mock, project, util.NewStdoutLogger(util.NoneLog))

	t.Run("Get an Account", func(t *testing.T) {
		mock.GetAccountMock = func(address flowsdk.Address) (*flowsdk.Account, error) {
			return tests.NewAccountWithAddress(address.String()), nil
		}

		account, err := accounts.Get(serviceAddress)

		require.NoError(t, err)
		assert.Equal(t, account.Address.String(), serviceAddress)
	})

	t.Run("Create an Account", func(t *testing.T) {
		newAddress := "192440c99cb17282"

		mock.SendTransactionMock = func(tx *flowsdk.Transaction, signer *flow.Account) (*flowsdk.Transaction, error) {
			assert.Equal(t, tx.Authorizers[0].String(), serviceAddress)
			assert.Equal(t, signer.Address().String(), serviceAddress)

			return tests.NewTransaction(), nil
		}

		mock.GetTransactionResultMock = func(tx *flowsdk.Transaction) (*flowsdk.TransactionResult, error) {
			return tests.NewAccountCreateResult(newAddress), nil
		}

		mock.GetAccountMock = func(address flowsdk.Address) (*flowsdk.Account, error) {
			assert.Equal(t, address.String(), newAddress)

			return tests.NewAccountWithAddress(newAddress), nil
		}

		account, err := accounts.Create(serviceName, []string{pubKey}, sigAlgo, hashAlgo, nil)

		require.NoError(t, err)
		assert.Equal(t, len(account.Address), 8)
	})

	t.Run("Create an Account with Contract", func(t *testing.T) {
		newAddress := "192440c99cb17282"

		mock.SendTransactionMock = func(tx *flowsdk.Transaction, signer *flow.Account) (*flowsdk.Transaction, error) {
			assert.Equal(t, tx.Authorizers[0].String(), serviceAddress)
			assert.Equal(t, signer.Address().String(), serviceAddress)
			assert.True(t, strings.Contains(string(tx.Script), "acct.contracts.add"))

			return tests.NewTransaction(), nil
		}

		mock.GetTransactionResultMock = func(tx *flowsdk.Transaction) (*flowsdk.TransactionResult, error) {
			return tests.NewAccountCreateResult(newAddress), nil
		}

		mock.GetAccountMock = func(address flowsdk.Address) (*flowsdk.Account, error) {
			assert.Equal(t, address.String(), newAddress)

			return tests.NewAccountWithAddress(newAddress), nil
		}

		account, err := accounts.Create(serviceName, []string{pubKey}, sigAlgo, hashAlgo, []string{"Hello:../../../tests/Hello.cdc"})

		require.NoError(t, err)
		assert.Equal(t, len(account.Address), 8)
	})

	t.Run("Contract Add for Account", func(t *testing.T) {
		mock.SendTransactionMock = func(tx *flowsdk.Transaction, signer *flow.Account) (*flowsdk.Transaction, error) {
			assert.Equal(t, tx.Authorizers[0].String(), serviceAddress)
			assert.Equal(t, signer.Address().String(), serviceAddress)
			assert.True(t, strings.Contains(string(tx.Script), "signer.contracts.add"))

			return tests.NewTransaction(), nil
		}

		mock.GetTransactionResultMock = func(tx *flowsdk.Transaction) (*flowsdk.TransactionResult, error) {
			return tests.NewTransactionResult(nil), nil
		}

		mock.GetAccountMock = func(address flowsdk.Address) (*flowsdk.Account, error) {
			return tests.NewAccountWithAddress(address.String()), nil
		}

		account, err := accounts.AddContract(serviceName, "Hello", "../../../tests/Hello.cdc", false)

		require.NoError(t, err)
		assert.Equal(t, len(account.Address), 8)
	})

	t.Run("Contract Update for Account", func(t *testing.T) {
		mock.SendTransactionMock = func(tx *flowsdk.Transaction, signer *flow.Account) (*flowsdk.Transaction, error) {
			assert.Equal(t, tx.Authorizers[0].String(), serviceAddress)
			assert.Equal(t, signer.Address().String(), serviceAddress)
			assert.True(t, strings.Contains(string(tx.Script), "signer.contracts.update__experimental"))

			return tests.NewTransaction(), nil
		}

		mock.GetTransactionResultMock = func(tx *flowsdk.Transaction) (*flowsdk.TransactionResult, error) {
			return tests.NewTransactionResult(nil), nil
		}

		mock.GetAccountMock = func(address flowsdk.Address) (*flowsdk.Account, error) {
			return tests.NewAccountWithAddress(address.String()), nil
		}

		account, err := accounts.AddContract(serviceName, "Hello", "../../../tests/Hello.cdc", true)

		require.NoError(t, err)
		assert.Equal(t, len(account.Address), 8)
	})

	t.Run("Contract Remove for Account", func(t *testing.T) {
		mock.SendTransactionMock = func(tx *flowsdk.Transaction, signer *flow.Account) (*flowsdk.Transaction, error) {
			assert.Equal(t, tx.Authorizers[0].String(), serviceAddress)
			assert.Equal(t, signer.Address().String(), serviceAddress)
			assert.True(t, strings.Contains(string(tx.Script), "signer.contracts.remove"))

			return tests.NewTransaction(), nil
		}

		mock.GetTransactionResultMock = func(tx *flowsdk.Transaction) (*flowsdk.TransactionResult, error) {
			return tests.NewTransactionResult(nil), nil
		}

		mock.GetAccountMock = func(address flowsdk.Address) (*flowsdk.Account, error) {
			return tests.NewAccountWithAddress(address.String()), nil
		}

		account, err := accounts.RemoveContract("Hello", serviceName)

		require.NoError(t, err)
		assert.Equal(t, len(account.Address), 8)
	})

}
