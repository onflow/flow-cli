package tests

import (
	"testing"

	"github.com/onflow/flow-cli/flow/lib"

	"github.com/onflow/flow-go-sdk"

	"github.com/onflow/flow-go/utils/io"

	"github.com/onflow/flow-cli/flow/services"

	"github.com/onflow/flow-cli/flow/gateway"
	"github.com/onflow/flow-cli/flow/util"
	"github.com/stretchr/testify/require"
	"gotest.tools/assert"
)

const (
	serviceAddress  = "f8d6e0586b0a20c7"
	contractPath    = "./Hello.cdc"
	emulatorAccount = "emulator-account"
	host            = "127.0.0.1:3569"
	conf            = "./flow.json"
)

var logger = util.NewStdoutLogger(util.NoneLog)

func TestAccount(t *testing.T) {
	helloContract, _ := io.ReadFile(contractPath)

	gw, err := gateway.NewGrpcGateway(host)
	require.NoError(t, err)

	project, err := lib.LoadProject([]string{conf})
	require.NoError(t, err)

	accounts := services.NewAccounts(gw, project, logger)

	/*
		t.Run("Address Test", func(t *testing.T) {
			tx := services.NewTransactions(gw, project, logger)
			_, tr, _ := tx.GetStatus("c0ff9c817f54526d69d381ba1e22e2721e95308b2e88f9107a543b7f233fce05", false)
			fmt.Println(tr.Events[0].Value)

		})
	*/

	t.Run("Get an Account", func(t *testing.T) {
		account, err := accounts.Get(serviceAddress)

		require.NoError(t, err)
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

		require.NoError(t, err)
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

		require.NoError(t, err)
		assert.Equal(t, string(acc.Contracts["Hello"]), string(helloContract))
	})

	t.Run("Account Update Contract", func(t *testing.T) {
		acc, err := accounts.AddContract(
			emulatorAccount,
			"Hello",
			contractPath,
			true,
		)

		require.NoError(t, err)
		assert.Equal(t, string(acc.Contracts["Hello"]), string(helloContract))
	})

	t.Run("Account Update Contract", func(t *testing.T) {
		acc, err := accounts.AddContract(
			emulatorAccount,
			"Hello",
			contractPath,
			true,
		)

		require.NoError(t, err)
		assert.Equal(t, string(acc.Contracts["Hello"]), string(helloContract))
	})

	t.Run("Account Remove Contract", func(t *testing.T) {
		acc, err := accounts.RemoveContract("Hello", emulatorAccount)

		require.NoError(t, err)
		assert.Equal(t, string(acc.Contracts["Hello"]), "")
	})
}

func TestEvents(t *testing.T) {
	gateway, err := gateway.NewGrpcGateway(host)
	require.NoError(t, err)

	project, err := lib.LoadProject([]string{conf})
	require.NoError(t, err)

	events := services.NewEvents(gateway, project, logger)

	t.Run("Get Event", func(t *testing.T) {
		event, err := events.Get("flow.createAccount", "0", "100")

		require.NoError(t, err)
		require.Greater(t, len(event), 0)
	})
}

func TestKeys(t *testing.T) {
	gateway, err := gateway.NewGrpcGateway(host)
	require.NoError(t, err)

	project, err := lib.LoadProject([]string{conf})
	require.NoError(t, err)

	keys := services.NewKeys(gateway, project, logger)

	t.Run("Generate keys", func(t *testing.T) {
		key, err := keys.Generate("", "ECDSA_P256")

		require.NoError(t, err)
		assert.Equal(t, key.Algorithm().String(), "ECDSA_P256")
		assert.Equal(t, len(key.PublicKey().String()), 130)
	})
}

func TestProject(t *testing.T) {
	gateway, err := gateway.NewGrpcGateway(host)
	require.NoError(t, err)

	project, err := lib.LoadProject([]string{conf})
	require.NoError(t, err)

	projects := services.NewProject(gateway, project, logger)

	t.Run("Deploy project", func(t *testing.T) {
		contracts, err := projects.Deploy("emulator", true)

		require.NoError(t, err)
		assert.Equal(t, contracts[0].Name(), "NonFungibleToken")
		assert.Equal(t, contracts[1].Name(), "Foo")
		assert.Equal(t, contracts[1].Dependencies()["./NonFungibleToken.cdc"].Target(), contracts[0].Target())
		assert.Equal(t, len(contracts), 2)
	})
}

func TestScripts(t *testing.T) {
	gateway, err := gateway.NewGrpcGateway(host)
	require.NoError(t, err)

	project, err := lib.LoadProject([]string{conf})
	require.NoError(t, err)

	scripts := services.NewScripts(gateway, project, logger)

	t.Run("Test Script", func(t *testing.T) {
		val, err := scripts.Execute("./script.cdc", []string{"String:Mr G"}, "")

		require.NoError(t, err)
		assert.Equal(t, val.String(), `"Hello Mr G"`)
	})

	t.Run("Test Script JSON args", func(t *testing.T) {
		val, err := scripts.Execute("./script.cdc", []string{}, "[{\"type\": \"String\", \"value\": \"Mr G\"}]")

		require.NoError(t, err)
		assert.Equal(t, val.String(), `"Hello Mr G"`)
	})
}

func TestTransactions(t *testing.T) {
	gateway, err := gateway.NewGrpcGateway(host)
	require.NoError(t, err)

	project, err := lib.LoadProject([]string{conf})
	require.NoError(t, err)

	transactions := services.NewTransactions(gateway, project, logger)
	var txID1 flow.Identifier

	t.Run("Test Transactions", func(t *testing.T) {
		tx, tr, err := transactions.Send("./transaction.cdc", emulatorAccount, []string{"String:Hello"}, "")
		txID1 = tx.ID()

		require.NoError(t, err)
		assert.Equal(t, tx.Payer.String(), serviceAddress)
		assert.Equal(t, tr.Status.String(), "SEALED")
	})

	t.Run("Test Failed Transactions", func(t *testing.T) {
		tx, tr, err := transactions.Send("./transactionErr.cdc", emulatorAccount, []string{}, "")

		require.NoError(t, err)
		assert.Equal(t, tx.Payer.String(), serviceAddress)
		assert.Equal(t, tr.Status.String(), "SEALED")
		require.Greater(t, len(tr.Error.Error()), 100)
	})

	t.Run("Get Transaction", func(t *testing.T) {
		tx, tr, err := transactions.GetStatus(txID1.Hex(), true)

		require.NoError(t, err)
		assert.Equal(t, tx.Payer.String(), serviceAddress)
		assert.Equal(t, tr.Status.String(), "SEALED")
	})
}
