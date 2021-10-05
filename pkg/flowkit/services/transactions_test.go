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
	"fmt"
	"strings"
	"testing"

	"github.com/onflow/flow-cli/pkg/flowkit/config"

	"github.com/onflow/flow-go-sdk/crypto"

	"github.com/onflow/cadence"
	"github.com/onflow/flow-go-sdk"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/onflow/flow-cli/pkg/flowkit"
	"github.com/onflow/flow-cli/tests"
)

const gasLimit = 1000

func TestTransactions(t *testing.T) {
	t.Parallel()

	state, _, _ := setup()
	serviceAcc, _ := state.EmulatorServiceAccount()
	serviceAddress := serviceAcc.Address()

	t.Run("Get Transaction", func(t *testing.T) {
		t.Parallel()
		_, s, gw := setup()
		txs := tests.NewTransaction()

		_, _, err := s.Transactions.GetStatus(txs.ID(), true)

		assert.NoError(t, err)
		gw.Mock.AssertNumberOfCalls(t, tests.GetTransactionResultFunc, 1)
		gw.Mock.AssertCalled(t, tests.GetTransactionFunc, txs.ID())
	})

	t.Run("Send Transaction args", func(t *testing.T) {
		t.Parallel()
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

		args := []cadence.Value{cadence.String("Bar")}
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

func setupAccounts(state *flowkit.State, s *Services) {
	setupAccount(state, s, tests.Alice())
	setupAccount(state, s, tests.Bob())
	setupAccount(state, s, tests.Charlie())
}

func setupAccount(state *flowkit.State, s *Services, account *flowkit.Account) {
	srv, _ := state.EmulatorServiceAccount()

	key := account.Key()
	pk, _ := key.PrivateKey()
	acc, _ := s.Accounts.Create(srv,
		[]crypto.PublicKey{(*pk).PublicKey()},
		[]int{flow.AccountKeyWeightThreshold},
		key.SigAlgo(),
		key.HashAlgo(),
		nil,
	)

	newAcc := &flowkit.Account{}
	newAcc.SetName(account.Name())
	newAcc.SetAddress(acc.Address)
	newAcc.SetKey(key)

	state.Accounts().AddOrUpdate(newAcc)
}

func TestTransactions_Integration(t *testing.T) {
	t.Parallel()

	t.Run("Build Transaction", func(t *testing.T) {
		t.Parallel()
		state, s := setupIntegration()
		setupAccounts(state, s)

		type txIn struct {
			prop    flow.Address
			auth    []flow.Address
			payer   flow.Address
			index   int
			code    []byte
			file    string
			gas     uint64
			args    []cadence.Value
			network string
		}

		a, _ := state.Accounts().ByName("Alice")
		b, _ := state.Accounts().ByName("Bob")
		c, _ := state.Accounts().ByName("Charlie")

		txIns := []txIn{{
			a.Address(),
			[]flow.Address{a.Address()},
			a.Address(),
			0,
			tests.TransactionSimple.Source,
			tests.TransactionSimple.Filename,
			1000,
			nil,
			"",
		}, {
			c.Address(),
			[]flow.Address{a.Address(), b.Address()},
			c.Address(),
			0,
			tests.TransactionSimple.Source,
			tests.TransactionSimple.Filename,
			1000,
			nil,
			"",
		}}

		for _, i := range txIns {
			tx, err := s.Transactions.Build(i.prop, i.auth, i.payer, i.index, i.code, i.file, i.gas, i.args, i.network)

			assert.NoError(t, err)
			ftx := tx.FlowTransaction()
			assert.Equal(t, ftx.Script, i.code)
			assert.Equal(t, ftx.Payer, i.payer)
			assert.Equal(t, ftx.Authorizers, i.auth)
			assert.Equal(t, ftx.ProposalKey.Address, i.prop)
			assert.Equal(t, ftx.ProposalKey.KeyIndex, i.index)
		}

	})

	t.Run("Build Transaction with Imports", func(t *testing.T) {
		t.Parallel()
		state, s := setupIntegration()
		setupAccounts(state, s)

		srvAcc, _ := state.EmulatorServiceAccount()
		signer := srvAcc.Address()

		// setup
		c := config.Contract{
			Name:    tests.ContractHelloString.Name,
			Source:  tests.ContractHelloString.Filename,
			Network: "emulator",
		}
		state.Contracts().AddOrUpdate(c.Name, c)

		n := config.Network{
			Name: "emulator",
			Host: "127.0.0.1:3569",
		}
		state.Networks().AddOrUpdate(n.Name, n)

		d := config.Deployment{
			Network: n.Name,
			Account: srvAcc.Name(),
			Contracts: []config.ContractDeployment{{
				Name: c.Name,
				Args: nil,
			}},
		}
		state.Deployments().AddOrUpdate(d)
		_, _ = s.Accounts.AddContract(srvAcc, tests.ContractHelloString.Name, tests.ContractHelloString.Source, false)

		tx, err := s.Transactions.Build(
			signer,
			[]flow.Address{signer},
			signer,
			srvAcc.Key().Index(),
			tests.TransactionImports.Source,
			tests.TransactionImports.Filename,
			1000,
			nil,
			n.Name,
		)

		assert.NoError(t, err)
		ftx := tx.FlowTransaction()
		assert.Equal(t,
			string(ftx.Script),
			strings.ReplaceAll(
				string(tests.TransactionImports.Source),
				"import Hello from \"./contractHello.cdc\"",
				fmt.Sprintf("import Hello from 0x%s", srvAcc.Address()),
			),
		)
	})

}
