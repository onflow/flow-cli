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

	"github.com/stretchr/testify/mock"

	"github.com/onflow/flow-cli/pkg/flowkit"
	"github.com/onflow/flow-go-sdk"

	"github.com/onflow/flow-cli/tests"
	"github.com/stretchr/testify/assert"
)

const gasLimit = 1000

func TestTransactions(t *testing.T) {
	state, _, _ := setup()
	serviceAcc, _ := state.EmulatorServiceAccount()
	serviceAddress := serviceAcc.Address()

	t.Run("Get Transaction", func(t *testing.T) {
		_, s, gw := setup()
		txs := tests.NewTransaction()

		_, _, err := s.Transactions.GetStatus(txs.ID(), true)

		assert.NoError(t, err)
		gw.Mock.AssertNumberOfCalls(t, tests.GetTransactionResultFunc, 1)
		gw.Mock.AssertCalled(t, tests.GetTransactionFunc, txs.ID())
	})

	t.Run("Send Transaction args", func(t *testing.T) {
		_, s, gw := setup()

		var txID flow.Identifier
		gw.SendSignedTransaction.Run(func(args mock.Arguments) {
			tx := args.Get(0).(*flowkit.Transaction)
			arg, err := tx.FlowTransaction().Argument(0)
			assert.NoError(t, err)
			assert.Equal(t, arg.String(), "\"Bar\"")
			assert.Equal(t, tx.Signer().Address(), serviceAddress)
			assert.Equal(t, len(string(tx.FlowTransaction().Script)), 227)

			t := tests.NewTransaction()
			txID = t.ID()
			gw.SendSignedTransaction.Return(t, nil)
		})

		gw.GetTransactionResult.Run(func(args mock.Arguments) {
			assert.Equal(t, args.Get(0).(*flow.Transaction).ID(), txID)
			gw.GetTransactionResult.Return(tests.NewTransactionResult(nil), nil)
		})

		args, _ := flowkit.ParseArgumentsCommaSplit([]string{"String:Bar"})

		_, _, err := s.Transactions.Send(
			serviceAcc,
			tests.TransactionArgString.Source,
			"",
			gasLimit,
			args,
			"",
		)

		assert.NoError(t, err)
		gw.Mock.AssertNumberOfCalls(t, tests.SendSignedTransactionFunc, 1)
		gw.Mock.AssertNumberOfCalls(t, tests.GetTransactionResultFunc, 1)
	})

}
