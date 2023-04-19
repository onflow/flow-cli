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
	"context"
	accounts2 "github.com/onflow/flow-cli/flowkit/accounts"
	"testing"

	"github.com/onflow/cadence"
	flowsdk "github.com/onflow/flow-go-sdk"
	"github.com/onflow/flow-go-sdk/crypto"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/onflow/flow-cli/flowkit"
	"github.com/onflow/flow-cli/flowkit/config"
	"github.com/onflow/flow-cli/flowkit/gateway"
	"github.com/onflow/flow-cli/flowkit/output"
)

const testAccountName = "test-account"

func initTestnet(t *testing.T) (gateway.Gateway, *flowkit.State, flowkit.Services, flowkit.ReaderWriter, afero.Fs) {
	readerWriter, mockFs := ReaderWriter()

	state, err := flowkit.Init(readerWriter, crypto.ECDSA_P256, crypto.SHA3_256)
	require.NoError(t, err)

	gw, err := gateway.NewGrpcGateway(config.TestnetNetwork)
	require.NoError(t, err)

	logger := output.NewStdoutLogger(output.NoneLog)
	flow := flowkit.NewFlowkit(state, config.TestnetNetwork, gw, logger)

	key, err := crypto.DecodePrivateKeyHex(crypto.ECDSA_P256, "4b2b6442fcbef2209bc1182af15d203a6195346cc8d95ebb433d3df1acb3910c")
	require.NoError(t, err)

	funderKey := accounts2.NewHexKeyFromPrivateKey(0, crypto.SHA3_256, key)
	funder := &accounts2.Account{
		Name:    "funder",
		Address: flowsdk.HexToAddress("0x72ddb3d2cec14114"),
		Key:     funderKey,
	}

	testKey, err := flow.GenerateKey(context.Background(), crypto.ECDSA_P256, "")
	require.NoError(t, err)

	flowAccount, _, err := flow.CreateAccount(
		context.Background(),
		funder,
		[]flowkit.AccountPublicKey{{
			Public:   testKey.PublicKey(),
			Weight:   1000,
			SigAlgo:  crypto.ECDSA_P256,
			HashAlgo: crypto.SHA3_256,
		}},
	)
	require.NoError(t, err)

	testAccount := &accounts2.Account{
		Name:    testAccountName,
		Address: flowAccount.Address,
		Key:     accounts2.NewHexKeyFromPrivateKey(0, crypto.SHA3_256, testKey),
	}
	state.Accounts().AddOrUpdate(testAccount)

	// fund the account
	// TODO(sideninja) refactor core contracts lib to offer the template
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
	_, _, err = flow.SendTransaction(
		context.Background(),
		flowkit.NewTransactionSingleAccountRole(*funder),
		flowkit.Script{
			Code: transferTx,
			Args: []cadence.Value{amount, cadence.NewAddress(testAccount.Address)},
		},
		flowsdk.DefaultTransactionGasLimit,
	)
	require.NoError(t, err)

	return gw, state, flow, readerWriter, mockFs
}

var testnet = config.TestnetNetwork.Name

func Test_Foo(t *testing.T) {
	_, st, _, rw, _ := initTestnet(t)

	rw.WriteFile("test", []byte("foo"), 0644)

	out, _ := rw.ReadFile("test")
	assert.Equal(t, out, []byte("foo"))

	rw.WriteFile("test", []byte("bar"), 0644)
	out, _ = st.ReadFile("test")
	assert.Equal(t, out, []byte("bar"))
}

func Test_Testnet_ProjectDeploy(t *testing.T) {
	_, state, flow, rw, _ := initTestnet(t)

	state.Contracts().AddOrUpdate(config.Contract{
		Name:     ContractA.Name,
		Location: ContractA.Filename,
	})

	state.Contracts().AddOrUpdate(config.Contract{
		Name:     ContractB.Name,
		Location: ContractB.Filename,
	})

	state.Contracts().AddOrUpdate(config.Contract{
		Name:     ContractC.Name,
		Location: ContractC.Filename,
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

	contracts, err := flow.DeployProject(context.Background(), flowkit.UpdateExistingContract(false))
	assert.NoError(t, err)
	assert.Len(t, contracts, 3)
	assert.Equal(t, ContractA.Name, contracts[0].Name)
	assert.Equal(t, ContractB.Name, contracts[1].Name)
	assert.Equal(t, ContractC.Name, contracts[2].Name)

	// make a change
	updated := []byte(`
		import "ContractA"
		pub contract ContractB {
			pub init() {}
		}
	`)
	ContractB.Source = updated
	err = rw.WriteFile(ContractB.Filename, ContractB.Source, 0644)
	require.NoError(t, err)

	contracts, err = flow.DeployProject(context.Background(), flowkit.UpdateExistingContract(true))
	assert.NoError(t, err)
	assert.Len(t, contracts, 3)
	assert.Equal(t, ContractA.Name, contracts[0].Name)
	assert.Equal(t, ContractB.Name, contracts[1].Name)
	assert.Equal(t, ContractB.Source, updated)
	assert.Equal(t, ContractC.Name, contracts[2].Name)
}
