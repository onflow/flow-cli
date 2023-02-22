/*
 * Flow CLI
 *
 * Copyright 2019 Dapper Labs, Inc.
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

package tests

import (
	"testing"

	"github.com/onflow/cadence"
	"github.com/onflow/flow-go-sdk"
	"github.com/onflow/flow-go-sdk/crypto"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/onflow/flow-cli/pkg/flowkit"
	"github.com/onflow/flow-cli/pkg/flowkit/config"
	"github.com/onflow/flow-cli/pkg/flowkit/gateway"
	"github.com/onflow/flow-cli/pkg/flowkit/output"
	"github.com/onflow/flow-cli/pkg/flowkit/services"
)

const testAccountName = "test-account"

func initTestnet(t *testing.T) (gateway.Gateway, *flowkit.State, *services.Services, afero.Fs) {
	readerWriter, mockFs := ReaderWriter()

	state, err := flowkit.Init(readerWriter, crypto.ECDSA_P256, crypto.SHA3_256)
	require.NoError(t, err)

	gw, err := gateway.NewGrpcGateway(config.DefaultTestnetNetwork().Host)
	require.NoError(t, err)

	logger := output.NewStdoutLogger(output.NoneLog)
	srv := services.NewServices(gw, state, logger)

	key, err := crypto.DecodePrivateKeyHex(crypto.ECDSA_P256, "4b2b6442fcbef2209bc1182af15d203a6195346cc8d95ebb433d3df1acb3910c")
	require.NoError(t, err)

	funderKey := flowkit.NewHexAccountKeyFromPrivateKey(0, crypto.SHA3_256, key)
	funder := flowkit.NewAccount("funder").SetKey(funderKey).SetAddress(flow.HexToAddress("0x72ddb3d2cec14114"))

	testKey, err := srv.Keys.Generate("", crypto.ECDSA_P256)
	require.NoError(t, err)

	flowAccount, err := srv.Accounts.Create(
		funder,
		[]crypto.PublicKey{testKey.PublicKey()},
		[]int{1000},
		[]crypto.SignatureAlgorithm{crypto.ECDSA_P256},
		[]crypto.HashAlgorithm{crypto.SHA3_256},
		nil,
	)
	require.NoError(t, err)

	testAccount := flowkit.
		NewAccount(testAccountName).
		SetKey(flowkit.NewHexAccountKeyFromPrivateKey(0, crypto.SHA3_256, testKey)).
		SetAddress(flowAccount.Address)

	state.Accounts().AddOrUpdate(testAccount)

	// fund the account
	// todo refactor core contracts lib to offer the template
	transferTx := []byte(`
	import FungibleToken from 0x9a0766d93b6608b7
	import FlowToken from 0x7e60df042a9c0868
	
	transaction(amount: UFix64, to: Address) {
	
		// The Vault resource that holds the tokens that are being transferred
		let sentVault: @FungibleToken.Vault
	
		prepare(signer: AuthAccount) {
	
			// Get a reference to the signer's stored vault
			let vaultRef = signer.borrow<&FlowToken.Vault>(from: /storage/flowTokenVault)
				?? panic("Could not borrow reference to the owner's Vault!")
	
			// Withdraw tokens from the signer's stored vault
			self.sentVault <- vaultRef.withdraw(amount: amount)
		}
	
		execute {
	
			// Get a reference to the recipient's Receiver
			let receiverRef =  getAccount(to)
				.getCapability(/public/flowTokenReceiver)
				.borrow<&{FungibleToken.Receiver}>()
				?? panic("Could not borrow receiver reference to the recipient's Vault")
	
			// Deposit the withdrawn tokens in the recipient's receiver
			receiverRef.deposit(from: <-self.sentVault)
		}
	}`)

	amount, _ := cadence.NewUFix64("0.01")
	_, _, err = srv.Transactions.Send(
		services.NewSingleTransactionAccount(funder),
		flowkit.NewScript(transferTx, []cadence.Value{amount, cadence.NewAddress(testAccount.Address())}, ""),
		flow.DefaultTransactionGasLimit,
		testnet,
	)
	require.NoError(t, err)

	return gw, state, srv, mockFs
}

var testnet = config.DefaultTestnetNetwork().Name

func Test_Testnet_ProjectDeploy(t *testing.T) {
	_, state, srv, mockFs := initTestnet(t)

	state.Contracts().AddOrUpdate(config.Contract{
		Name:     ContractA.Name,
		Location: ContractA.Filename,
		Network:  testnet,
	})

	state.Contracts().AddOrUpdate(ContractB.Name, config.Contract{
		Name:     ContractB.Name,
		Location: ContractB.Filename,
		Network:  testnet,
	})

	state.Contracts().AddOrUpdate(ContractC.Name, config.Contract{
		Name:     ContractC.Name,
		Location: ContractC.Filename,
		Network:  testnet,
	})

	initArg, _ := cadence.NewString("foo")
	state.Deployments().AddOrUpdate(config.Deployment{
		Network: testnet,
		Account: testAccountName,
		Contracts: []config.ContractDeployment{
			{Name: ContractA.Name},
			{Name: ContractB.Name},
			{Name: ContractC.Name, Args: []cadence.Value{initArg}},
		},
	})

	contracts, err := srv.Project.Deploy(testnet, true)
	assert.NoError(t, err)
	assert.Len(t, contracts, 3)
	assert.Equal(t, ContractA.Name, contracts[0].Name)
	assert.Equal(t, ContractB.Name, contracts[1].Name)
	assert.Equal(t, ContractC.Name, contracts[2].Name)

	// make a change
	ContractA.Source = []byte(`pub contract ContractA { init() {} }`)
	_ = afero.WriteFile(mockFs, ContractA.Filename, ContractA.Source, 0644)

	contracts, err = srv.Project.Deploy(testnet, true)
	assert.NoError(t, err)
	assert.Len(t, contracts, 3)
	assert.Equal(t, ContractA.Name, contracts[0].Name)
	assert.Equal(t, ContractB.Name, contracts[1].Name)
	assert.Equal(t, ContractC.Name, contracts[2].Name)
}
