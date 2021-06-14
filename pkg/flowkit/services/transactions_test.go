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

	"github.com/onflow/flow-cli/pkg/flowkit"

	"github.com/onflow/flow-go-sdk"
	"github.com/onflow/flow-go-sdk/crypto"
	"github.com/stretchr/testify/assert"

	"github.com/onflow/flow-cli/pkg/flowkit/output"
	"github.com/onflow/flow-cli/tests"
)

const gasLimit = 1000

func TestTransactions(t *testing.T) {
	mock := tests.DefaultMockGateway()
	readerWriter := tests.ReaderWriter()

	proj, err := flowkit.Init(readerWriter, crypto.ECDSA_P256, crypto.SHA3_256)
	assert.NoError(t, err)

	serviceAcc, _ := proj.EmulatorServiceAccount()
	serviceAddress := serviceAcc.Address()
	transactions := NewTransactions(mock, proj, output.NewStdoutLogger(output.NoneLog))

	t.Run("Get Transaction", func(t *testing.T) {
		txs := tests.NewTransaction()

		mock.GetTransactionResultMock = func(tx *flow.Transaction) (*flow.TransactionResult, error) {
			assert.Equal(t, tx.ID(), txs.ID())
			return tests.NewTransactionResult(nil), nil
		}

		mock.GetTransactionMock = func(id flow.Identifier) (*flow.Transaction, error) {
			return txs, nil
		}

		_, _, err := transactions.GetStatus(txs.ID(), true)

		mock.AssertFuncsCalled(t, true, mock.GetTransactionResult, mock.GetTransaction)
		assert.NoError(t, err)
	})

	t.Run("Send Transaction args", func(t *testing.T) {
		var txID flow.Identifier
		mock.SendSignedTransactionMock = func(tx *flowkit.Transaction) (*flow.Transaction, error) {
			arg, err := tx.FlowTransaction().Argument(0)
			assert.NoError(t, err)
			assert.Equal(t, arg.String(), "\"Bar\"")
			assert.Equal(t, tx.Signer().Address(), serviceAddress)
			assert.Equal(t, len(string(tx.FlowTransaction().Script)), 227)

			t := tests.NewTransaction()
			txID = t.ID()
			return t, nil
		}

		mock.GetTransactionResultMock = func(tx *flow.Transaction) (*flow.TransactionResult, error) {
			assert.Equal(t, tx.ID(), txID)
			return tests.NewTransactionResult(nil), nil
		}

		args, _ := flowkit.ParseArgumentsCommaSplit([]string{"String:Bar"})

		_, _, err := transactions.Send(
			serviceAcc,
			tests.TransactionArgString.Source,
			"",
			gasLimit,
			args,
			"",
		)

		mock.AssertFuncsCalled(t, true, mock.GetTransactionResult, mock.SendSignedTransaction)
		assert.NoError(t, err)
	})

}
