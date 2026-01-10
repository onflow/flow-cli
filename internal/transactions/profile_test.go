/*
 * Flow CLI
 *
 * Copyright Flow Foundation
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

package transactions

import (
	"io"
	"testing"
	"time"

	"github.com/onflow/flow-go-sdk"
	"github.com/onflow/flow-go-sdk/crypto"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/onflow/flow-emulator/convert"
	"github.com/onflow/flow-emulator/emulator"
	"github.com/onflow/flow-emulator/server"
	"github.com/onflow/flowkit/v2"
	"github.com/onflow/flowkit/v2/accounts"
	"github.com/onflow/flowkit/v2/config"
	"github.com/onflow/flowkit/v2/gateway"
	"github.com/onflow/flowkit/v2/output"
	"github.com/onflow/flowkit/v2/tests"

	"github.com/onflow/flow-cli/internal/command"
	"github.com/onflow/flow-cli/internal/util"
)

func Test_Profile_Validation(t *testing.T) {
	srv, state, _ := util.TestMocks(t)

	t.Run("Fail no network specified", func(t *testing.T) {
		profileFlags.Network = ""
		result, err := profile([]string{"0x01"}, command.GlobalFlags{}, util.NoLogger, srv.Mock, state)
		assert.EqualError(t, err, "network must be specified with --network flag")
		assert.Nil(t, result)
	})

	t.Run("Fail network not found", func(t *testing.T) {
		profileFlags.Network = "invalid-network"
		result, err := profile([]string{"0x01"}, command.GlobalFlags{}, util.NoLogger, srv.Mock, state)
		assert.EqualError(t, err, "network \"invalid-network\" not found in flow.json")
		assert.Nil(t, result)
		profileFlags.Network = ""
	})

	t.Run("Fail transaction not found", func(t *testing.T) {
		profileFlags.Network = "testnet"
		srv.GetTransactionByID.Return(nil, nil, assert.AnError)
		result, err := profile([]string{"0x01"}, command.GlobalFlags{}, util.NoLogger, srv.Mock, state)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to get transaction")
		assert.Nil(t, result)
		profileFlags.Network = ""
	})

	t.Run("Fail transaction not sealed", func(t *testing.T) {
		profileFlags.Network = "testnet"
		tx := tests.NewTransaction()
		result := tests.NewTransactionResult(nil)
		result.Status = flow.TransactionStatusPending
		srv.GetTransactionByID.Return(tx, result, nil)

		res, err := profile([]string{"0x01"}, command.GlobalFlags{}, util.NoLogger, srv.Mock, state)
		assert.EqualError(t, err, "transaction is not sealed (status: PENDING)")
		assert.Nil(t, res)
		profileFlags.Network = ""
	})
}

func Test_ProfilingResult(t *testing.T) {
	txID := flow.HexToID("0123456789abcdef")
	tx := tests.NewTransaction()
	txResult := tests.NewTransactionResult(nil)
	txResult.Status = flow.TransactionStatusSealed

	t.Run("Result without profiling data", func(t *testing.T) {
		result := &profilingResult{
			txID:        txID,
			tx:          tx,
			result:      txResult,
			networkName: "testnet",
			blockHeight: 123,
		}

		output := result.String()
		assert.Contains(t, output, txID.String())
		assert.Contains(t, output, "testnet")
		assert.Contains(t, output, "Note: Computation profiling data not available")
	})

	t.Run("Result with profiling data", func(t *testing.T) {
		txReport := &emulator.ProcedureReport{
			ComputationUsed: 42,
			MemoryEstimate:  1024,
			Intensities: map[string]uint64{
				"test1": 10,
				"test2": 20,
			},
		}

		result := &profilingResult{
			txID:        txID,
			tx:          tx,
			result:      txResult,
			txReport:    txReport,
			networkName: "testnet",
			blockHeight: 123,
		}

		output := result.String()
		assert.Contains(t, output, "Computation used: 42")
		assert.Contains(t, output, "Memory estimate:  1024 bytes")

		jsonOutput := result.JSON()
		jsonMap, ok := jsonOutput.(map[string]any)
		require.True(t, ok)
		assert.Equal(t, "testnet", jsonMap["network"])
		assert.Equal(t, uint64(123), jsonMap["block_height"])
		assert.NotNil(t, jsonMap["computation_metrics"])
	})

	t.Run("Oneliner format", func(t *testing.T) {
		result := &profilingResult{
			txID:        txID,
			tx:          tx,
			result:      txResult,
			networkName: "testnet",
			blockHeight: 123,
		}

		oneliner := result.Oneliner()
		assert.Contains(t, oneliner, txID.String())
		assert.Contains(t, oneliner, "SEALED")
		assert.Contains(t, oneliner, "123")
	})

	t.Run("Exit code on success", func(t *testing.T) {
		result := &profilingResult{
			txID:        txID,
			tx:          tx,
			result:      txResult,
			networkName: "testnet",
			blockHeight: 123,
		}
		assert.Equal(t, 0, result.ExitCode())
	})
}

func Test_Profile_Integration_LocalEmulator(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	emulatorHost := "127.0.0.1:3570"
	t.Log("Starting local emulator...")
	emulatorServer, testTxID, testBlockHeight := startEmulatorWithTestTransaction(t, emulatorHost)
	defer emulatorServer.Stop()

	time.Sleep(500 * time.Millisecond)

	t.Run("Profile transaction by forking local emulator", func(t *testing.T) {
		rw, _ := tests.ReaderWriter()
		state, err := flowkit.Init(rw)
		require.NoError(t, err)

		emulatorAccount, err := accounts.NewEmulatorAccount(rw, crypto.ECDSA_P256, crypto.SHA3_256, "")
		require.NoError(t, err)
		state.Accounts().AddOrUpdate(emulatorAccount)

		state.Networks().AddOrUpdate(config.Network{
			Name: "emulator",
			Host: emulatorHost,
		})

		gw, err := gateway.NewGrpcGateway(config.Network{
			Name: "emulator",
			Host: emulatorHost,
		})
		require.NoError(t, err)

		logger := output.NewStdoutLogger(output.InfoLog)
		services := flowkit.NewFlowkit(state, config.Network{Name: "emulator", Host: emulatorHost}, gw, logger)

		profileFlags.Network = "emulator"
		defer func() { profileFlags.Network = "" }()

		t.Logf("Attempting to profile transaction %s...", testTxID.String())

		done := make(chan bool)
		var result interface{}
		var profileErr error

		go func() {
			result, profileErr = profile(
				[]string{testTxID.String()},
				command.GlobalFlags{},
				logger,
				services,
				state,
			)
			done <- true
		}()

		select {
		case <-done:
			if profileErr != nil {
				t.Fatalf("Profile failed: %v", profileErr)
			}

			require.NotNil(t, result)

			profilingResult, ok := result.(*profilingResult)
			require.True(t, ok)

			assert.Equal(t, testTxID, profilingResult.txID)
			assert.Equal(t, "emulator", profilingResult.networkName)
			assert.Equal(t, testBlockHeight, profilingResult.blockHeight)
			assert.NotNil(t, profilingResult.tx)
			assert.NotNil(t, profilingResult.result)

			t.Log("\n" + profilingResult.String())

			assert.NotNil(t, profilingResult.txReport)
			if profilingResult.txReport != nil {
				t.Logf("✓ Computation used: %d", profilingResult.txReport.ComputationUsed)
				t.Logf("✓ Memory estimate: %d bytes", profilingResult.txReport.MemoryEstimate)
			}

			jsonOutput := profilingResult.JSON()
			require.NotNil(t, jsonOutput)
			jsonMap, ok := jsonOutput.(map[string]any)
			require.True(t, ok)
			assert.Equal(t, "emulator", jsonMap["network"])
			assert.Equal(t, testBlockHeight, jsonMap["block_height"])

			t.Log("✓ Successfully profiled transaction!")

		case <-time.After(30 * time.Second):
			t.Fatal("Profile command timed out")
		}
	})
}

func startEmulatorWithTestTransaction(t *testing.T, host string) (*server.EmulatorServer, flow.Identifier, uint64) {
	zlog := zerolog.New(zerolog.ConsoleWriter{Out: io.Discard})

	serverConf := &server.Config{
		GRPCPort:                    3570,
		Host:                        "127.0.0.1",
		ComputationReportingEnabled: true,
		StorageLimitEnabled:         false,
	}

	emulatorServer := server.NewEmulatorServer(&zlog, serverConf)
	go emulatorServer.Start()

	time.Sleep(2 * time.Second)

	blockchain := emulatorServer.Emulator()

	testTx := flow.NewTransaction().
		SetScript([]byte(`
			transaction {
				prepare(signer: &Account) {
					log("Test transaction")
				}
				execute {
					var i = 0
					while i < 10 {
						i = i + 1
					}
				}
			}
		`)).
		SetGasLimit(1000).
		SetProposalKey(
			blockchain.ServiceKey().Address,
			blockchain.ServiceKey().Index,
			blockchain.ServiceKey().SequenceNumber,
		).
		SetPayer(blockchain.ServiceKey().Address).
		AddAuthorizer(blockchain.ServiceKey().Address)

	signer, err := blockchain.ServiceKey().Signer()
	require.NoError(t, err)

	err = testTx.SignEnvelope(
		blockchain.ServiceKey().Address,
		blockchain.ServiceKey().Index,
		signer,
	)
	require.NoError(t, err)

	flowGoTx := convert.SDKTransactionToFlow(*testTx)
	err = blockchain.AddTransaction(*flowGoTx)
	require.NoError(t, err)

	_, _, err = blockchain.ExecuteAndCommitBlock()
	require.NoError(t, err)

	latestBlock, err := blockchain.GetLatestBlock()
	require.NoError(t, err)

	t.Logf("Created test transaction %s at block %d", testTx.ID().String(), latestBlock.Height)

	return emulatorServer, testTx.ID(), latestBlock.Height
}
