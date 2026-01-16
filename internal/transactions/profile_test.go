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
	"context"
	"fmt"
	"io"
	"net"
	"testing"
	"time"

	"github.com/onflow/flow-emulator/convert"
	"github.com/onflow/flow-emulator/emulator"
	"github.com/onflow/flow-emulator/server"
	flow "github.com/onflow/flow-go-sdk"
	"github.com/onflow/flow-go-sdk/crypto"
	"github.com/onflow/flow-go-sdk/templates"
	flowgo "github.com/onflow/flow-go/model/flow"
	"github.com/onflow/flowkit/v2"
	"github.com/onflow/flowkit/v2/accounts"
	"github.com/onflow/flowkit/v2/config"
	"github.com/onflow/flowkit/v2/gateway"
	"github.com/onflow/flowkit/v2/output"
	"github.com/onflow/flowkit/v2/tests"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/onflow/flow-cli/internal/command"
	"github.com/onflow/flow-cli/internal/util"
)

const (
	emulatorStartupWait = 2 * time.Second
	emulatorStableWait  = 500 * time.Millisecond
	profileTestTimeout  = 15 * time.Second
	initialBlockCount   = 3
	transactionGasLimit = 1000
)

func Test_Profile_Validation(t *testing.T) {
	t.Parallel()

	srv, state, _ := util.TestMocks(t)

	t.Run("Fail no network specified", func(t *testing.T) {
		t.Parallel()
		result, err := profile([]string{"0x01"}, command.GlobalFlags{}, util.NoLogger, srv.Mock, state)
		assert.EqualError(t, err, "network must be specified with --network flag")
		assert.Nil(t, result)
	})

	t.Run("Fail network not found", func(t *testing.T) {
		t.Parallel()
		result, err := profile([]string{"0x01"}, command.GlobalFlags{Network: "invalid-network"}, util.NoLogger, srv.Mock, state)
		assert.EqualError(t, err, "network \"invalid-network\" not found in flow.json")
		assert.Nil(t, result)
	})

	t.Run("Fail transaction not found", func(t *testing.T) {
		t.Parallel()
		srv.GetTransactionByID.Return(nil, nil, assert.AnError)
		result, err := profile([]string{"0x01"}, command.GlobalFlags{Network: "testnet"}, util.NoLogger, srv.Mock, state)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to get transaction")
		assert.Nil(t, result)
	})

	t.Run("Fail transaction not sealed", func(t *testing.T) {
		t.Parallel()
		tx := tests.NewTransaction()
		result := tests.NewTransactionResult(nil)
		result.Status = flow.TransactionStatusPending
		srv.GetTransactionByID.Return(tx, result, nil)

		res, err := profile([]string{"0x01"}, command.GlobalFlags{Network: "testnet"}, util.NoLogger, srv.Mock, state)
		assert.EqualError(t, err, "transaction is not sealed (status: PENDING)")
		assert.Nil(t, res)
	})
}

func Test_ProfilingResult(t *testing.T) {
	t.Parallel()

	txID := flow.HexToID("0123456789abcdef")
	tx := tests.NewTransaction()
	txResult := tests.NewTransactionResult(nil)
	txResult.Status = flow.TransactionStatusSealed

	t.Run("Result with profile file", func(t *testing.T) {
		t.Parallel()
		result := &profilingResult{
			txID:        txID,
			tx:          tx,
			result:      txResult,
			networkName: "testnet",
			blockHeight: 123,
			profileFile: "test-profile.pb.gz",
		}

		output := result.String()
		assert.Contains(t, output, txID.String())
		assert.Contains(t, output, "testnet")
		assert.Contains(t, output, "Profile saved: test-profile.pb.gz")
		assert.Contains(t, output, "go tool pprof")

		jsonOutput := result.JSON()
		jsonMap, ok := jsonOutput.(map[string]any)
		require.True(t, ok)
		assert.Equal(t, "testnet", jsonMap["network"])
		assert.Equal(t, uint64(123), jsonMap["block_height"])
		assert.Equal(t, "test-profile.pb.gz", jsonMap["profileFile"])
	})

	t.Run("Oneliner format", func(t *testing.T) {
		t.Parallel()
		result := &profilingResult{
			txID:        txID,
			tx:          tx,
			result:      txResult,
			networkName: "testnet",
			blockHeight: 123,
		}

		oneliner := result.Oneliner()
		assert.Contains(t, oneliner, txID.String()[:8])
		assert.Contains(t, oneliner, "profiled successfully")
	})
}

func Test_Profile_Integration_LocalEmulator(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	t.Run("Profile user transaction", func(t *testing.T) {
		t.Parallel()

		port := getFreePort(t)
		emulatorHost := fmt.Sprintf("127.0.0.1:%d", port)
		emulatorServer, testTxID, testBlockHeight := startEmulatorWithTestTransaction(t, emulatorHost, port)
		defer emulatorServer.Stop()

		time.Sleep(emulatorStableWait)

		runProfileTest(t, emulatorHost, testTxID, testBlockHeight)
	})

	t.Run("Profile failed transaction", func(t *testing.T) {
		t.Parallel()

		port := getFreePort(t)
		emulatorHost := fmt.Sprintf("127.0.0.1:%d", port)
		emulatorServer, failedTxID, testBlockHeight := startEmulatorWithFailedTransaction(t, emulatorHost, port)
		defer emulatorServer.Stop()

		time.Sleep(emulatorStableWait)

		runProfileTest(t, emulatorHost, failedTxID, testBlockHeight)
	})

	t.Run("Profile transaction with multiple prior transactions", func(t *testing.T) {
		t.Parallel()

		port := getFreePort(t)
		emulatorHost := fmt.Sprintf("127.0.0.1:%d", port)
		emulatorServer, targetTxID, testBlockHeight := startEmulatorWithMultipleTransactions(t, emulatorHost, port, 5)
		defer emulatorServer.Stop()

		time.Sleep(emulatorStableWait)

		runProfileTest(t, emulatorHost, targetTxID, testBlockHeight)
	})

	t.Run("Profile system transaction", func(t *testing.T) {
		t.Skip("System transactions via gRPC not supported in local emulator - tested manually on mainnet")
		t.Parallel()

		port := getFreePort(t)
		emulatorHost := fmt.Sprintf("127.0.0.1:%d", port)
		emulatorServer, systemTxID, testBlockHeight := startEmulatorWithScheduledTransaction(t, emulatorHost, port)
		defer emulatorServer.Stop()

		time.Sleep(emulatorStableWait)

		require.NotEqual(t, flow.Identifier{}, systemTxID, "System transaction should be found")
		require.Greater(t, testBlockHeight, uint64(0), "Block height should be valid")

		runProfileTest(t, emulatorHost, systemTxID, testBlockHeight)
	})
}

func getFreePort(t *testing.T) int {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		if t != nil {
			require.NoError(t, err)
		} else {
			panic(fmt.Sprintf("failed to get free port: %v", err))
		}
	}
	defer listener.Close()
	return listener.Addr().(*net.TCPAddr).Port
}

func runProfileTest(t *testing.T, emulatorHost string, testTxID flow.Identifier, testBlockHeight uint64) {
	rw, _ := tests.ReaderWriter()

	state, err := flowkit.Init(rw)
	require.NoError(t, err)

	emulatorAccount, err := accounts.NewEmulatorAccount(rw, crypto.ECDSA_P256, crypto.SHA3_256, "")
	require.NoError(t, err)
	state.Accounts().AddOrUpdate(emulatorAccount)

	network := config.Network{Name: "emulator", Host: emulatorHost}
	state.Networks().AddOrUpdate(network)

	gw, err := gateway.NewGrpcGateway(network)
	require.NoError(t, err)

	logger := output.NewStdoutLogger(output.InfoLog)
	services := flowkit.NewFlowkit(state, network, gw, logger)

	result, err := profile(
		[]string{testTxID.String()},
		command.GlobalFlags{Network: "emulator"},
		logger,
		services,
		state,
	)
	require.NoError(t, err)
	require.NotNil(t, result)

	profilingResult, ok := result.(*profilingResult)
	require.True(t, ok)

	assert.Equal(t, testTxID, profilingResult.txID)
	assert.Equal(t, "emulator", profilingResult.networkName)
	assert.Equal(t, testBlockHeight, profilingResult.blockHeight)
	assert.NotNil(t, profilingResult.tx)
	assert.NotNil(t, profilingResult.result)
	assert.NotEmpty(t, profilingResult.profileFile)
	assert.Equal(t, testTxID, profilingResult.tx.ID())

	jsonOutput := profilingResult.JSON()
	require.NotNil(t, jsonOutput)
	jsonMap, ok := jsonOutput.(map[string]any)
	require.True(t, ok)
	assert.Equal(t, "emulator", jsonMap["network"])
	assert.Equal(t, testBlockHeight, jsonMap["block_height"])
}

func createEmulatorServer(port int) *server.EmulatorServer {
	zlog := zerolog.New(zerolog.ConsoleWriter{Out: io.Discard})

	restPort := getFreePort(nil)
	adminPort := getFreePort(nil)
	debuggerPort := getFreePort(nil)

	serverConf := &server.Config{
		GRPCPort:                     port,
		RESTPort:                     restPort,
		AdminPort:                    adminPort,
		DebuggerPort:                 debuggerPort,
		Host:                         "127.0.0.1",
		ComputationReportingEnabled:  true,
		StorageLimitEnabled:          false,
		WithContracts:                true,
		ScheduledTransactionsEnabled: true,
		ChainID:                      "flow-emulator",
	}

	emulatorServer := server.NewEmulatorServer(&zlog, serverConf)
	go emulatorServer.Start()

	// Wait for gRPC server to be ready
	maxWait := 5 * time.Second
	start := time.Now()
	for time.Since(start) < maxWait {
		conn, err := grpc.NewClient(
			fmt.Sprintf("%s:%d", serverConf.Host, serverConf.GRPCPort),
			grpc.WithTransportCredentials(insecure.NewCredentials()),
		)
		if err == nil {
			conn.Close()
			break
		}
		time.Sleep(50 * time.Millisecond)
	}

	return emulatorServer
}

func createInitialBlocks(t *testing.T, blockchain emulator.Emulator) {
	for i := 0; i < initialBlockCount; i++ {
		_, _, err := blockchain.ExecuteAndCommitBlock()
		require.NoError(t, err)
	}
}

func buildTransaction(t *testing.T, script string, blockchain emulator.Emulator) *flow.Transaction {
	latestBlock, err := blockchain.GetLatestBlock()
	require.NoError(t, err)

	serviceKey := blockchain.ServiceKey()

	return flow.NewTransaction().
		SetScript([]byte(script)).
		SetComputeLimit(transactionGasLimit).
		SetProposalKey(
			serviceKey.Address,
			serviceKey.Index,
			serviceKey.SequenceNumber,
		).
		SetReferenceBlockID(convert.FlowIdentifierToSDK(latestBlock.ID())).
		SetPayer(serviceKey.Address).
		AddAuthorizer(serviceKey.Address)
}

func submitAndCommitTransaction(t *testing.T, tx *flow.Transaction, blockchain emulator.Emulator) {
	err := blockchain.AddTransaction(*convert.SDKTransactionToFlow(*tx))
	require.NoError(t, err)

	_, _, err = blockchain.ExecuteAndCommitBlock()
	require.NoError(t, err)
}

func startEmulatorWithTestTransaction(t *testing.T, host string, port int) (*server.EmulatorServer, flow.Identifier, uint64) {
	emulatorServer := createEmulatorServer(port)
	blockchain := emulatorServer.Emulator()

	createInitialBlocks(t, blockchain)

	testTx := buildTransaction(t, `
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
	`, blockchain)

	submitAndCommitTransaction(t, testTx, blockchain)

	latestBlock, err := blockchain.GetLatestBlock()
	require.NoError(t, err)

	return emulatorServer, testTx.ID(), latestBlock.Height
}

func startEmulatorWithFailedTransaction(t *testing.T, host string, port int) (*server.EmulatorServer, flow.Identifier, uint64) {
	emulatorServer := createEmulatorServer(port)
	blockchain := emulatorServer.Emulator()

	createInitialBlocks(t, blockchain)

	failTx := buildTransaction(t, `
		transaction {
			prepare(signer: &Account) {
				log("About to fail")
			}
			execute {
				panic("Intentional failure for testing")
			}
		}
	`, blockchain)

	submitAndCommitTransaction(t, failTx, blockchain)

	block, err := blockchain.GetLatestBlock()
	require.NoError(t, err)

	return emulatorServer, failTx.ID(), block.Height
}

func startEmulatorWithMultipleTransactions(t *testing.T, host string, port int, count int) (*server.EmulatorServer, flow.Identifier, uint64) {
	emulatorServer := createEmulatorServer(port)
	blockchain := emulatorServer.Emulator()

	createInitialBlocks(t, blockchain)

	var lastTxID flow.Identifier
	serviceKey := blockchain.ServiceKey()

	for i := 0; i < count; i++ {
		tx := buildTransaction(t, fmt.Sprintf(`
			transaction {
				prepare(signer: &Account) {
					log("Transaction %d")
				}
				execute {
					var sum = 0
					var i = 0
					while i < 10 {
						sum = sum + i
						i = i + 1
					}
				}
			}
		`, i), blockchain)

		submitAndCommitTransaction(t, tx, blockchain)

		lastTxID = tx.ID()
		serviceKey.SequenceNumber++
	}

	block, err := blockchain.GetLatestBlock()
	require.NoError(t, err)

	return emulatorServer, lastTxID, block.Height
}

func startEmulatorWithScheduledTransaction(t *testing.T, host string, port int) (*server.EmulatorServer, flow.Identifier, uint64) {
	emulatorServer := createEmulatorServer(port)

	blockchain := emulatorServer.Emulator()
	serviceAddress := blockchain.ServiceKey().Address

	contractCode := `
import FlowTransactionScheduler from 0xf8d6e0586b0a20c7

access(all) contract TestHandler {
	access(all) resource Handler: FlowTransactionScheduler.TransactionHandler {
		access(FlowTransactionScheduler.Execute) fun executeTransaction(id: UInt64, data: AnyStruct?) {
			log("Handler executed with ID: ".concat(id.toString()))
			var sum = 0
			var i = 0
			while i < 100 {
				sum = sum + i
				i = i + 1
			}
		}
	}
	access(all) fun createHandler(): @Handler {
		return <- create Handler()
	}
}`

	latestBlock, err := blockchain.GetLatestBlock()
	require.NoError(t, err)

	deployTx := templates.AddAccountContract(serviceAddress, templates.Contract{Name: "TestHandler", Source: contractCode})
	deployTx.SetComputeLimit(flowgo.DefaultMaxTransactionGasLimit).
		SetReferenceBlockID(convert.FlowIdentifierToSDK(latestBlock.ID())).
		SetProposalKey(serviceAddress, blockchain.ServiceKey().Index, blockchain.ServiceKey().SequenceNumber).
		SetPayer(serviceAddress)

	signer, err := blockchain.ServiceKey().Signer()
	require.NoError(t, err)

	err = deployTx.SignEnvelope(serviceAddress, blockchain.ServiceKey().Index, signer)
	require.NoError(t, err)

	err = blockchain.SendTransaction(convert.SDKTransactionToFlow(*deployTx))
	require.NoError(t, err)

	_, _, err = blockchain.ExecuteAndCommitBlock()
	require.NoError(t, err)

	scheduleScript := `
import FlowTransactionScheduler from 0xf8d6e0586b0a20c7
import TestHandler from 0xf8d6e0586b0a20c7
import FungibleToken from 0xee82856bf20e2aa6
import FlowToken from 0x0ae53cb6e3f42a79

transaction {
	prepare(acct: auth(Storage, Capabilities, FungibleToken.Withdraw) &Account) {
		let handler <- TestHandler.createHandler()
		acct.storage.save(<-handler, to: /storage/testHandler)
		let issued = acct.capabilities.storage.issue<auth(FlowTransactionScheduler.Execute) &{FlowTransactionScheduler.TransactionHandler}>(/storage/testHandler)

		let adminRef = acct.storage.borrow<&FlowToken.Administrator>(from: /storage/flowTokenAdmin) ?? panic("missing admin")
		let minter <- adminRef.createNewMinter(allowedAmount: 10.0)
		let minted <- minter.mintTokens(amount: 1.0)
		let receiver = acct.capabilities.borrow<&{FungibleToken.Receiver}>(/public/flowTokenReceiver) ?? panic("missing receiver")
		receiver.deposit(from: <-minted)
		destroy minter

		let vaultRef = acct.storage.borrow<auth(FungibleToken.Withdraw) &FlowToken.Vault>(from: /storage/flowTokenVault) ?? panic("missing vault")
		let fees <- (vaultRef.withdraw(amount: 0.001) as! @FlowToken.Vault)

		destroy <- FlowTransactionScheduler.schedule(
			handlerCap: issued,
			data: nil,
			timestamp: getCurrentBlock().timestamp + 1.0,
			priority: FlowTransactionScheduler.Priority.High,
			executionEffort: UInt64(5000),
			fees: <-fees
		)
	}
}`

	latestBlock, err = blockchain.GetLatestBlock()
	require.NoError(t, err)

	scheduleTx := flow.NewTransaction().
		SetScript([]byte(scheduleScript)).
		SetComputeLimit(flowgo.DefaultMaxTransactionGasLimit).
		SetProposalKey(serviceAddress, blockchain.ServiceKey().Index, blockchain.ServiceKey().SequenceNumber).
		SetPayer(serviceAddress).
		AddAuthorizer(serviceAddress).
		SetReferenceBlockID(convert.FlowIdentifierToSDK(latestBlock.ID()))

	signer, err = blockchain.ServiceKey().Signer()
	require.NoError(t, err)

	err = scheduleTx.SignEnvelope(serviceAddress, blockchain.ServiceKey().Index, signer)
	require.NoError(t, err)

	err = blockchain.SendTransaction(convert.SDKTransactionToFlow(*scheduleTx))
	require.NoError(t, err)

	_, _, err = blockchain.ExecuteAndCommitBlock()
	require.NoError(t, err)

	scheduleBlockHeight := latestBlock.Height

	// Commit multiple blocks to ensure scheduled transaction is processed
	for i := 0; i < 5; i++ {
		time.Sleep(1500 * time.Millisecond)
		_, _, err = blockchain.ExecuteAndCommitBlock()
		require.NoError(t, err)
	}

	gw, err := gateway.NewGrpcGateway(config.Network{Host: host})
	require.NoError(t, err)

	rw, _ := tests.ReaderWriter()

	state, err := flowkit.Init(rw)
	require.NoError(t, err)

	services := flowkit.NewFlowkit(state, config.Network{Name: "emulator", Host: host}, gw, output.NewStdoutLogger(output.NoneLog))

	latestBlock, err = blockchain.GetLatestBlock()
	require.NoError(t, err)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Look for system transactions in blocks after scheduling
	for height := scheduleBlockHeight + 1; height <= latestBlock.Height; height++ {
		block, err := services.GetBlock(ctx, flowkit.BlockQuery{Height: height})
		if err != nil {
			t.Logf("Failed to get block at height %d: %v", height, err)
			continue
		}

		// GetTransactionsByBlockID on emulator DOES include system transactions
		txs, _, err := services.GetTransactionsByBlockID(ctx, block.ID)
		if err != nil {
			t.Logf("Failed to get transactions for block %d: %v", height, err)
			continue
		}

		t.Logf("Block %d has %d transactions", height, len(txs))
		for _, tx := range txs {
			t.Logf("  Transaction %s, Payer: %s", tx.ID().String()[:8], tx.Payer.String())
			if tx.Payer == flow.EmptyAddress {
				t.Logf("Found system transaction: %s at height %d", tx.ID(), height)
				return emulatorServer, tx.ID(), height
			}
		}
	}

	t.Fatalf("No system transaction found after scheduled transaction (searched heights %d to %d)", scheduleBlockHeight+1, latestBlock.Height)
	return emulatorServer, flow.Identifier{}, 0
}
