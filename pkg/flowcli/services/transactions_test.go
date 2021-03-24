package services

import (
	"testing"

	"github.com/onflow/flow-cli/pkg/flowcli/output"
	"github.com/onflow/flow-cli/pkg/flowcli/project"

	"github.com/onflow/flow-go-sdk"
	"github.com/stretchr/testify/assert"

	"github.com/onflow/flow-cli/tests"
	"github.com/onflow/flow-go-sdk/crypto"
)

func TestTransactions(t *testing.T) {
	mock := &tests.MockGateway{}

	proj, err := project.InitProject(crypto.ECDSA_P256, crypto.SHA3_256)
	assert.NoError(t, err)

	transactions := NewTransactions(mock, proj, output.NewStdoutLogger(output.NoneLog))

	t.Run("Get Transaction", func(t *testing.T) {
		called := 0
		txs := tests.NewTransaction()

		mock.GetTransactionResultMock = func(tx *flow.Transaction) (*flow.TransactionResult, error) {
			called++
			assert.Equal(t, tx.ID().String(), txs.ID().String())
			return tests.NewTransactionResult(nil), nil
		}

		mock.GetTransactionMock = func(id flow.Identifier) (*flow.Transaction, error) {
			called++
			return txs, nil
		}

		_, _, err := transactions.GetStatus(txs.ID().String(), true)

		assert.NoError(t, err)
		assert.Equal(t, called, 2)
	})

	t.Run("Send Transaction args", func(t *testing.T) {
		called := 0

		mock.GetTransactionResultMock = func(tx *flow.Transaction) (*flow.TransactionResult, error) {
			called++
			return tests.NewTransactionResult(nil), nil
		}

		mock.SendTransactionMock = func(tx *flow.Transaction, signer *project.Account) (*flow.Transaction, error) {
			called++
			arg, err := tx.Argument(0)

			assert.NoError(t, err)
			assert.Equal(t, arg.String(), "\"Bar\"")
			assert.Equal(t, signer.Address().String(), serviceAddress)
			assert.Equal(t, len(string(tx.Script)), 209)
			return tests.NewTransaction(), nil
		}

		_, _, err := transactions.Send("../../../tests/transaction.cdc", serviceName, []string{"String:Bar"}, "")

		assert.NoError(t, err)
		assert.Equal(t, called, 2)
	})

	t.Run("Send Transaction JSON args", func(t *testing.T) {
		called := 0

		mock.GetTransactionResultMock = func(tx *flow.Transaction) (*flow.TransactionResult, error) {
			called++
			return tests.NewTransactionResult(nil), nil
		}

		mock.SendTransactionMock = func(tx *flow.Transaction, signer *project.Account) (*flow.Transaction, error) {
			called++
			assert.Equal(t, signer.Address().String(), serviceAddress)
			assert.Equal(t, len(string(tx.Script)), 209)
			return tests.NewTransaction(), nil
		}

		_, _, err := transactions.Send(
			"../../../tests/transaction.cdc",
			serviceName,
			nil,
			"[{\"type\": \"String\", \"value\": \"Bar\"}]",
		)

		assert.NoError(t, err)
		assert.Equal(t, called, 2)
	})

	t.Run("Send Transaction Fails wrong args", func(t *testing.T) {
		_, _, err := transactions.Send("../../../tests/transaction.cdc", serviceName, []string{"Bar"}, "")
		assert.Equal(t, err.Error(), "Argument not passed in correct format, correct format is: Type:Value, got Bar")
	})

	t.Run("Send Transaction Fails wrong filename", func(t *testing.T) {
		_, _, err := transactions.Send("nooo.cdc", serviceName, []string{"Bar"}, "")
		assert.Equal(t, err.Error(), "Failed to load file: nooo.cdc")
	})

	t.Run("Send Transaction Fails wrong args", func(t *testing.T) {
		_, _, err := transactions.Send("../../../tests/transaction.cdc", serviceName, nil, "[{\"Bar\":\"No\"}]")
		assert.Equal(t, err.Error(), "failed to decode value: invalid JSON Cadence structure")
	})
}
