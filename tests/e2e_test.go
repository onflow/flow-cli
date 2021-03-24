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

package tests

import (
	"os"
	"testing"

	"github.com/onflow/flow-cli/pkg/flow/output"
	"github.com/onflow/flow-cli/pkg/flow/project"

	"github.com/onflow/flow-cli/pkg/flow/gateway"
	"github.com/onflow/flow-cli/pkg/flow/services"
	flowsdk "github.com/onflow/flow-go-sdk"
	"github.com/onflow/flow-go/utils/io"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	serviceAddress  = "f8d6e0586b0a20c7"
	contractPath    = "./Hello.cdc"
	emulatorAccount = "emulator-account"
	host            = "127.0.0.1:3569"
	conf            = "./flow.json"
)

var logger = output.NewStdoutLogger(output.NoneLog)
var e2e = os.Getenv("E2E")

func TestAccount(t *testing.T) {
	if e2e == "" {
		t.Skip("Skipping end-to-end tests")
	}

	helloContract, _ := io.ReadFile(contractPath)

	gw, err := gateway.NewGrpcGateway(host)
	assert.NoError(t, err)

	project, err := project.LoadProject([]string{conf})
	assert.NoError(t, err)

	accounts := services.NewAccounts(gw, project, logger)

	t.Run("Get an Account", func(t *testing.T) {
		account, err := accounts.Get(serviceAddress)

		assert.NoError(t, err)
		assert.Equal(t, account.Address.String(), serviceAddress)
	})

	t.Run("Creates an Account", func(t *testing.T) {
		keys := []string{"0x640a5a359bf3536d15192f18d872d57c98a96cb871b92b70cecb0739c2d5c37b4be12548d3526933c2cda9b0b9c69412f45ffb6b85b6840d8569d969fe84e5b7"}
		account, err := accounts.Create(
			emulatorAccount,
			keys,
			"ECDSA_P256",
			"SHA3_256",
			[]string{},
		)

		assert.NoError(t, err)
		assert.Equal(t, account.Keys[0].PublicKey.String(), keys[0])
		assert.Equal(t, string(account.Code), "")
	})

	t.Run("Account Add Contract", func(t *testing.T) {
		acc, err := accounts.AddContract(
			emulatorAccount,
			"Hello",
			contractPath,
			false,
		)

		assert.NoError(t, err)
		assert.Equal(t, string(acc.Contracts["Hello"]), string(helloContract))
	})

	t.Run("Account Update Contract", func(t *testing.T) {
		acc, err := accounts.AddContract(
			emulatorAccount,
			"Hello",
			contractPath,
			true,
		)

		assert.NoError(t, err)
		assert.Equal(t, string(acc.Contracts["Hello"]), string(helloContract))
	})

	t.Run("Account Update Contract", func(t *testing.T) {
		acc, err := accounts.AddContract(
			emulatorAccount,
			"Hello",
			contractPath,
			true,
		)

		assert.NoError(t, err)
		assert.Equal(t, string(acc.Contracts["Hello"]), string(helloContract))
	})

	t.Run("Account Remove Contract", func(t *testing.T) {
		acc, err := accounts.RemoveContract("Hello", emulatorAccount)

		assert.NoError(t, err)
		assert.Equal(t, string(acc.Contracts["Hello"]), "")
	})
}

func TestEvents(t *testing.T) {
	if e2e == "" {
		t.Skip("Skipping end-to-end tests")
	}

	gw, err := gateway.NewGrpcGateway(host)
	assert.NoError(t, err)

	project, err := project.LoadProject([]string{conf})
	assert.NoError(t, err)

	events := services.NewEvents(gw, project, logger)

	t.Run("Get Event", func(t *testing.T) {
		event, err := events.Get("flow.createAccount", "0", "100")

		assert.NoError(t, err)
		require.Greater(t, len(event), 0)
	})
}

func TestKeys(t *testing.T) {
	if e2e == "" {
		t.Skip("Skipping end-to-end tests")
	}

	gw, err := gateway.NewGrpcGateway(host)
	assert.NoError(t, err)

	proj, err := project.LoadProject([]string{conf})
	assert.NoError(t, err)

	keys := services.NewKeys(gw, proj, logger)

	t.Run("Generate keys", func(t *testing.T) {
		key, err := keys.Generate("", "ECDSA_P256")

		assert.NoError(t, err)
		assert.Equal(t, key.Algorithm().String(), "ECDSA_P256")
		assert.Equal(t, len(key.PublicKey().String()), 130)
	})
}

func TestProject(t *testing.T) {
	if e2e == "" {
		t.Skip("Skipping end-to-end tests")
	}

	gw, err := gateway.NewGrpcGateway(host)
	assert.NoError(t, err)

	project, err := project.LoadProject([]string{conf})
	assert.NoError(t, err)

	projects := services.NewProject(gw, project, logger)

	t.Run("Deploy project", func(t *testing.T) {
		contracts, err := projects.Deploy("emulator", true)

		assert.NoError(t, err)
		assert.Equal(t, contracts[0].Name(), "NonFungibleToken")
		assert.Equal(t, contracts[1].Name(), "Foo")
		assert.Equal(t, contracts[1].Dependencies()["./NonFungibleToken.cdc"].Target(), contracts[0].Target())
		assert.Equal(t, len(contracts), 2)
	})
}

func TestScripts(t *testing.T) {
	if e2e == "" {
		t.Skip("Skipping end-to-end tests")
	}

	gateway, err := gateway.NewGrpcGateway(host)
	assert.NoError(t, err)

	project, err := project.LoadProject([]string{conf})
	assert.NoError(t, err)

	scripts := services.NewScripts(gateway, project, logger)

	t.Run("Test Script", func(t *testing.T) {
		val, err := scripts.Execute("./script.cdc", []string{"String:Mr G"}, "")

		assert.NoError(t, err)
		assert.Equal(t, val.String(), `"Hello Mr G"`)
	})

	t.Run("Test Script JSON args", func(t *testing.T) {
		val, err := scripts.Execute("./script.cdc", []string{}, "[{\"type\": \"String\", \"value\": \"Mr G\"}]")

		assert.NoError(t, err)
		assert.Equal(t, val.String(), `"Hello Mr G"`)
	})
}

func TestTransactions(t *testing.T) {
	if e2e == "" {
		t.Skip("Skipping end-to-end tests")
	}

	gw, err := gateway.NewGrpcGateway(host)
	assert.NoError(t, err)

	project, err := project.LoadProject([]string{conf})
	assert.NoError(t, err)

	transactions := services.NewTransactions(gw, project, logger)
	var txID1 flowsdk.Identifier

	t.Run("Test Transactions", func(t *testing.T) {
		tx, tr, err := transactions.Send("./transaction.cdc", emulatorAccount, []string{"String:Hello"}, "")
		txID1 = tx.ID()

		assert.NoError(t, err)
		assert.Equal(t, tx.Payer.String(), serviceAddress)
		assert.Equal(t, tr.Status.String(), "SEALED")
	})

	t.Run("Test Failed Transactions", func(t *testing.T) {
		tx, tr, err := transactions.Send("./transactionErr.cdc", emulatorAccount, []string{}, "")

		assert.NoError(t, err)
		assert.Equal(t, tx.Payer.String(), serviceAddress)
		assert.Equal(t, tr.Status.String(), "SEALED")
		require.Greater(t, len(tr.Error.Error()), 100)
	})

	t.Run("Get Transaction", func(t *testing.T) {
		tx, tr, err := transactions.GetStatus(txID1.Hex(), true)

		assert.NoError(t, err)
		assert.Equal(t, tx.Payer.String(), serviceAddress)
		assert.Equal(t, tr.Status.String(), "SEALED")
	})
}
