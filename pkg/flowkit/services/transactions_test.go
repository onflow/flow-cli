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

	"github.com/spf13/afero"

	"github.com/onflow/flow-cli/pkg/flowkit"

	"github.com/onflow/flow-go-sdk"
	"github.com/onflow/flow-go-sdk/crypto"
	"github.com/stretchr/testify/assert"

	"github.com/onflow/flow-cli/pkg/flowkit/output"
	"github.com/onflow/flow-cli/tests"
)

const gasLimit = 1000

var transactionCode = []byte(`
	transaction(greeting: String) {
	  let guest: Address
	
	  prepare(authorizer: AuthAccount) {
		self.guest = authorizer.address
	  }
	
	  execute {
		log(greeting.concat(",").concat(self.guest.toString()))
	  }
	}
`)

func TestTransactions(t *testing.T) {
	mock := tests.DefaultMockGateway()

	af := afero.Afero{afero.NewMemMapFs()}
	proj, err := flowkit.Init(af, crypto.ECDSA_P256, crypto.SHA3_256)
	assert.NoError(t, err)
	serviceAcc := proj.Accounts().ByName(serviceName)
	transactions := NewTransactions(mock, proj, output.NewStdoutLogger(output.NoneLog))

	t.Run("Get Transaction", func(t *testing.T) {
		called := 0
		txs := tests.NewTransaction()

		mock.GetTransactionResultMock = func(tx *flow.Transaction) (*flow.TransactionResult, error) {
			called++
			assert.Equal(t, tx.ID(), txs.ID())
			return tests.NewTransactionResult(nil), nil
		}

		mock.GetTransactionMock = func(id flow.Identifier) (*flow.Transaction, error) {
			called++
			return txs, nil
		}

		_, _, err := transactions.GetStatus(txs.ID(), true)

		assert.NoError(t, err)
		assert.Equal(t, called, 2)
	})

	t.Run("Send Transaction args", func(t *testing.T) {
		called := 0

		mock.GetTransactionResultMock = func(tx *flow.Transaction) (*flow.TransactionResult, error) {
			called++
			return tests.NewTransactionResult(nil), nil
		}

		mock.SendSignedTransactionMock = func(tx *flowkit.Transaction) (*flow.Transaction, error) {
			called++
			arg, err := tx.FlowTransaction().Argument(0)

			assert.NoError(t, err)
			assert.Equal(t, arg.String(), "\"Bar\"")
			assert.Equal(t, tx.Signer().Address(), serviceAddress)
			assert.Equal(t, len(string(tx.FlowTransaction().Script)), 216)
			return tests.NewTransaction(), nil
		}

		args, _ := flowkit.ParseArgumentsCommaSplit([]string{"String:Bar"})

		_, _, err := transactions.Send(
			serviceAcc,
			transactionCode,
			"",
			gasLimit,
			args,
			"",
		)

		assert.NoError(t, err)
		assert.Equal(t, called, 2)
	})

}
