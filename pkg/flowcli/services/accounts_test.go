package services

import (
	"strings"
	"testing"

	"github.com/onflow/flow-cli/pkg/flowcli/output"
	"github.com/onflow/flow-cli/pkg/flowcli/project"

	"github.com/onflow/flow-go-sdk/crypto"

	"github.com/onflow/flow-go-sdk"
	"github.com/stretchr/testify/assert"

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

	proj, err := project.InitProject(crypto.ECDSA_P256, crypto.SHA3_256)
	assert.NoError(t, err)

	accounts := NewAccounts(mock, proj, output.NewStdoutLogger(output.NoneLog))

	t.Run("Get an Account", func(t *testing.T) {
		mock.GetAccountMock = func(address flow.Address) (*flow.Account, error) {
			return tests.NewAccountWithAddress(address.String()), nil
		}

		account, err := accounts.Get(serviceAddress)

		assert.NoError(t, err)
		assert.Equal(t, account.Address.String(), serviceAddress)
	})

	t.Run("Create an Account", func(t *testing.T) {
		newAddress := "192440c99cb17282"

		mock.SendTransactionMock = func(tx *flow.Transaction, signer *project.Account) (*flow.Transaction, error) {
			assert.Equal(t, tx.Authorizers[0].String(), serviceAddress)
			assert.Equal(t, signer.Address().String(), serviceAddress)

			return tests.NewTransaction(), nil
		}

		mock.GetTransactionResultMock = func(tx *flow.Transaction) (*flow.TransactionResult, error) {
			return tests.NewAccountCreateResult(newAddress), nil
		}

		mock.GetAccountMock = func(address flow.Address) (*flow.Account, error) {
			assert.Equal(t, address.String(), newAddress)

			return tests.NewAccountWithAddress(newAddress), nil
		}

		a, err := accounts.Create(serviceName, []string{pubKey}, sigAlgo, hashAlgo, nil)

		assert.NotNil(t, a)
		assert.NoError(t, err)
		assert.Equal(t, len(a.Address), 8)
	})

	t.Run("Create an Account with Contract", func(t *testing.T) {
		newAddress := "192440c99cb17282"

		mock.SendTransactionMock = func(tx *flow.Transaction, signer *project.Account) (*flow.Transaction, error) {
			assert.Equal(t, tx.Authorizers[0].String(), serviceAddress)
			assert.Equal(t, signer.Address().String(), serviceAddress)
			assert.True(t, strings.Contains(string(tx.Script), "acct.contracts.add"))

			return tests.NewTransaction(), nil
		}

		mock.GetTransactionResultMock = func(tx *flow.Transaction) (*flow.TransactionResult, error) {
			return tests.NewAccountCreateResult(newAddress), nil
		}

		mock.GetAccountMock = func(address flow.Address) (*flow.Account, error) {
			assert.Equal(t, address.String(), newAddress)

			return tests.NewAccountWithAddress(newAddress), nil
		}

		a, err := accounts.Create(serviceName, []string{pubKey}, sigAlgo, hashAlgo, []string{"Hello:../../../tests/Hello.cdc"})

		assert.NotNil(t, a)
		assert.NoError(t, err)
		assert.Equal(t, len(a.Address), 8)
	})

	t.Run("Contract Add for Account", func(t *testing.T) {
		mock.SendTransactionMock = func(tx *flow.Transaction, signer *project.Account) (*flow.Transaction, error) {
			assert.Equal(t, tx.Authorizers[0].String(), serviceAddress)
			assert.Equal(t, signer.Address().String(), serviceAddress)
			assert.True(t, strings.Contains(string(tx.Script), "signer.contracts.add"))

			return tests.NewTransaction(), nil
		}

		mock.GetTransactionResultMock = func(tx *flow.Transaction) (*flow.TransactionResult, error) {
			return tests.NewTransactionResult(nil), nil
		}

		mock.GetAccountMock = func(address flow.Address) (*flow.Account, error) {
			return tests.NewAccountWithAddress(address.String()), nil
		}

		a, err := accounts.AddContract(serviceName, "Hello", "../../../tests/Hello.cdc", false)

		assert.NotNil(t, a)
		assert.NoError(t, err)
		assert.Equal(t, len(a.Address), 8)
	})

	t.Run("Contract Update for Account", func(t *testing.T) {
		mock.SendTransactionMock = func(tx *flow.Transaction, signer *project.Account) (*flow.Transaction, error) {
			assert.Equal(t, tx.Authorizers[0].String(), serviceAddress)
			assert.Equal(t, signer.Address().String(), serviceAddress)
			assert.True(t, strings.Contains(string(tx.Script), "signer.contracts.update__experimental"))

			return tests.NewTransaction(), nil
		}

		mock.GetTransactionResultMock = func(tx *flow.Transaction) (*flow.TransactionResult, error) {
			return tests.NewTransactionResult(nil), nil
		}

		mock.GetAccountMock = func(address flow.Address) (*flow.Account, error) {
			return tests.NewAccountWithAddress(address.String()), nil
		}

		account, err := accounts.AddContract(serviceName, "Hello", "../../../tests/Hello.cdc", true)

		assert.NotNil(t, account)
		assert.Equal(t, len(account.Address), 8)
		assert.NoError(t, err)
	})

	t.Run("Contract Remove for Account", func(t *testing.T) {
		mock.SendTransactionMock = func(tx *flow.Transaction, signer *project.Account) (*flow.Transaction, error) {
			assert.Equal(t, tx.Authorizers[0].String(), serviceAddress)
			assert.Equal(t, signer.Address().String(), serviceAddress)
			assert.True(t, strings.Contains(string(tx.Script), "signer.contracts.remove"))

			return tests.NewTransaction(), nil
		}

		mock.GetTransactionResultMock = func(tx *flow.Transaction) (*flow.TransactionResult, error) {
			return tests.NewTransactionResult(nil), nil
		}

		mock.GetAccountMock = func(address flow.Address) (*flow.Account, error) {
			return tests.NewAccountWithAddress(address.String()), nil
		}

		account, err := accounts.RemoveContract("Hello", serviceName)

		assert.NotNil(t, account)
		assert.Equal(t, len(account.Address), 8)
		assert.NoError(t, err)
	})

}
