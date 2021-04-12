/*
 * Flow CLI
 *
 * Copyright 2019-2021 Dapper Labs, Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *   http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package services

import (
	"testing"

	"github.com/onflow/flow-go-sdk"
	"github.com/onflow/flow-go-sdk/crypto"
	"github.com/stretchr/testify/assert"

	"github.com/onflow/flow-cli/pkg/flowcli/output"
	"github.com/onflow/flow-cli/pkg/flowcli/project"
	"github.com/onflow/flow-cli/tests"
)

func TestTransactions(t *testing.T) {
	mock := &tests.MockGateway{}

	// default implementations
	mock.PrepareTransactionPayloadMock = func(tx *project.Transaction) (*project.Transaction, error) {
		return tx, nil
	}

	mock.SendSignedTransactionMock = func(tx *project.Transaction) (*flow.Transaction, error) {
		return tx.FlowTransaction(), nil
	}

	proj, err := project.Init(crypto.ECDSA_P256, crypto.SHA3_256)
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

		mock.SendSignedTransactionMock = func(tx *project.Transaction) (*flow.Transaction, error) {
			called++
			arg, err := tx.FlowTransaction().Argument(0)

			assert.NoError(t, err)
			assert.Equal(t, arg.String(), "\"Bar\"")
			assert.Equal(t, tx.Signer().Address().String(), serviceAddress)
			assert.Equal(t, len(string(tx.FlowTransaction().Script)), 209)
			return tests.NewTransaction(), nil
		}

		_, _, err := transactions.Send(
			"../../../tests/transaction.cdc",
			"",
			serviceName,
			[]string{"String:Bar"},
			"",
		)

		assert.NoError(t, err)
		assert.Equal(t, called, 2)
	})

	t.Run("Send Transaction JSON args", func(t *testing.T) {
		called := 0

		mock.GetTransactionResultMock = func(tx *flow.Transaction) (*flow.TransactionResult, error) {
			called++
			return tests.NewTransactionResult(nil), nil
		}

		mock.SendSignedTransactionMock = func(tx *project.Transaction) (*flow.Transaction, error) {
			called++
			assert.Equal(t, tx.Signer().Address().String(), serviceAddress)
			assert.Equal(t, len(string(tx.FlowTransaction().Script)), 209)
			return tests.NewTransaction(), nil
		}

		_, _, err := transactions.Send(
			"../../../tests/transaction.cdc",
			"",
			serviceName,
			nil,
			"[{\"type\": \"String\", \"value\": \"Bar\"}]",
		)

		assert.NoError(t, err)
		assert.Equal(t, called, 2)
	})

	t.Run("Send Transaction Fails wrong args", func(t *testing.T) {
		_, _, err := transactions.Send(
			"../../../tests/transaction.cdc",
			"",
			serviceName,
			[]string{"Bar"},
			"",
		)
		assert.Equal(t, err.Error(), "argument not passed in correct format, correct format is: Type:Value, got Bar")
	})

	t.Run("Send Transaction Fails wrong filename", func(t *testing.T) {
		_, _, err := transactions.Send(
			"nooo.cdc",
			"",
			serviceName,
			[]string{"Bar"},
			"",
		)
		assert.Equal(t, err.Error(), "Failed to load file: nooo.cdc")
	})

	t.Run("Send Transaction Fails wrong args", func(t *testing.T) {
		_, _, err := transactions.Send(
			"../../../tests/transaction.cdc",
			"",
			serviceName,
			nil,
			"[{\"Bar\":\"No\"}]",
		)
		assert.Equal(t, err.Error(), "failed to decode value: invalid JSON Cadence structure")
	})
}
