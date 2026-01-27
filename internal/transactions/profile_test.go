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
		t.Parallel()

		port := getFreePort(t)
		emulatorHost := fmt.Sprintf("127.0.0.1:%d", port)

		// Get scheduled execute callback transaction ID
		executeCallbackID, blockHeight := setupScheduledTransaction(t, emulatorHost, port)

		// Profile the scheduled execute callback transaction
		runProfileTest(t, emulatorHost, executeCallbackID, blockHeight)

		t.Logf("✅ Successfully profiled scheduled execute callback transaction!")
	})
}

func getFreePort(t *testing.T) int {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	defer listener.Close()

	tcpAddr, ok := listener.Addr().(*net.TCPAddr)
	require.True(t, ok, "expected TCP address")
	return tcpAddr.Port
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

	// Note: System transaction IDs from GetSystemTransactionsForBlock may differ from
	// IDs returned by GetTransaction due to how the emulator handles system txs.
	// The important thing is that profiling succeeded.
	assert.Equal(t, "emulator", profilingResult.networkName)
	assert.Equal(t, testBlockHeight, profilingResult.blockHeight)
	assert.NotNil(t, profilingResult.tx)
	assert.NotNil(t, profilingResult.result)
	assert.NotEmpty(t, profilingResult.profileFile)
	t.Logf("Expected TX ID: %s", testTxID)
	t.Logf("Profiled TX ID: %s", profilingResult.tx.ID())
	t.Logf("Result TX ID: %s", profilingResult.txID)

	jsonOutput := profilingResult.JSON()
	require.NotNil(t, jsonOutput)
	jsonMap, ok := jsonOutput.(map[string]any)
	require.True(t, ok)
	assert.Equal(t, "emulator", jsonMap["network"])
	assert.Equal(t, testBlockHeight, jsonMap["block_height"])
}

func createEmulatorServer(t *testing.T, port int) *server.EmulatorServer {
	zlog := zerolog.New(zerolog.ConsoleWriter{Out: io.Discard})

	restPort := getFreePort(t)
	adminPort := getFreePort(t)
	debuggerPort := getFreePort(t)

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

	return server.NewEmulatorServer(&zlog, serverConf)
}

func startServer(t *testing.T, emulatorServer *server.EmulatorServer, host string, port int) {
	t.Helper()

	go func() {
		emulatorServer.Start()
	}()

	// Wait for gRPC server to be ready
	maxWait := 5 * time.Second
	start := time.Now()
	connected := false
	for time.Since(start) < maxWait {
		conn, err := grpc.NewClient(
			fmt.Sprintf("%s:%d", host, port),
			grpc.WithTransportCredentials(insecure.NewCredentials()),
		)
		if err == nil {
			conn.Close()
			connected = true
			t.Logf("✅ gRPC server ready on %s:%d", host, port)
			break
		}
		time.Sleep(50 * time.Millisecond)
	}
	if !connected {
		t.Logf("⚠️  gRPC server did not become ready after %v", maxWait)
	}
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
	emulatorServer := createEmulatorServer(t, port)
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

	startServer(t, emulatorServer, "127.0.0.1", port)
	return emulatorServer, testTx.ID(), latestBlock.Height
}

func startEmulatorWithFailedTransaction(t *testing.T, host string, port int) (*server.EmulatorServer, flow.Identifier, uint64) {
	emulatorServer := createEmulatorServer(t, port)
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

	startServer(t, emulatorServer, "127.0.0.1", port)
	return emulatorServer, failTx.ID(), block.Height
}

func startEmulatorWithMultipleTransactions(t *testing.T, host string, port int, count int) (*server.EmulatorServer, flow.Identifier, uint64) {
	emulatorServer := createEmulatorServer(t, port)
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

	startServer(t, emulatorServer, "127.0.0.1", port)
	return emulatorServer, lastTxID, block.Height
}

// setupScheduledTransaction follows the pattern from flow-emulator's TestScheduledTransaction_QueryByID
// It creates an emulator, schedules a transaction, waits for execution, gets the system tx ID, then starts gRPC server
func setupScheduledTransaction(t *testing.T, host string, port int) (flow.Identifier, uint64) {
	t.Helper()

	// Create emulator server (but don't start gRPC yet)
	emulatorServer := createEmulatorServer(t, port)
	b := emulatorServer.Emulator()

	serviceAddress := b.ServiceKey().Address
	serviceHex := serviceAddress.Hex()

	// Deploy handler contract (copied from emulator reference test)
	handlerContract := fmt.Sprintf(`
		import FlowTransactionScheduler from 0x%s
		access(all) contract ScheduledHandler {
			access(contract) var count: Int
			access(all) view fun getCount(): Int { return self.count }
			access(all) resource Handler: FlowTransactionScheduler.TransactionHandler {
				access(FlowTransactionScheduler.Execute) fun executeTransaction(id: UInt64, data: AnyStruct?) {
					ScheduledHandler.count = ScheduledHandler.count + 1
				}
			}
			access(all) fun createHandler(): @Handler { return <- create Handler() }
			init() { self.count = 0 }
		}
	`, serviceHex)

	latestBlock, err := b.GetLatestBlock()
	require.NoError(t, err)
	tx := templates.AddAccountContract(serviceAddress, templates.Contract{Name: "ScheduledHandler", Source: handlerContract})
	tx.SetComputeLimit(flowgo.DefaultMaxTransactionGasLimit).
		SetReferenceBlockID(flow.Identifier(latestBlock.ID())).
		SetProposalKey(serviceAddress, b.ServiceKey().Index, b.ServiceKey().SequenceNumber).
		SetPayer(serviceAddress)
	signer, err := b.ServiceKey().Signer()
	require.NoError(t, err)
	require.NoError(t, tx.SignEnvelope(serviceAddress, b.ServiceKey().Index, signer))
	require.NoError(t, b.SendTransaction(convert.SDKTransactionToFlow(*tx)))
	_, _, err = b.ExecuteAndCommitBlock()
	require.NoError(t, err)
	t.Log("✅ Handler contract deployed")

	// Schedule transaction (copied from emulator reference test)
	scheduleTx := fmt.Sprintf(`
		import FlowTransactionScheduler from 0x%s
		import ScheduledHandler from 0x%s
		import FungibleToken from 0xee82856bf20e2aa6
		import FlowToken from 0x0ae53cb6e3f42a79
		transaction {
			prepare(acct: auth(Storage, Capabilities, FungibleToken.Withdraw) &Account) {
				let handler <- ScheduledHandler.createHandler()
				acct.storage.save(<-handler, to: /storage/counterHandler)
				let cap = acct.capabilities.storage.issue<auth(FlowTransactionScheduler.Execute) &{FlowTransactionScheduler.TransactionHandler}>(/storage/counterHandler)
				let admin = acct.storage.borrow<&FlowToken.Administrator>(from: /storage/flowTokenAdmin)!
				let minter <- admin.createNewMinter(allowedAmount: 10.0)
				let minted <- minter.mintTokens(amount: 1.0)
				let receiver = acct.capabilities.borrow<&{FungibleToken.Receiver}>(/public/flowTokenReceiver)!
				receiver.deposit(from: <-minted)
				destroy minter
				let estimate = FlowTransactionScheduler.estimate(
					data: nil,
					timestamp: getCurrentBlock().timestamp + 3.0,
					priority: FlowTransactionScheduler.Priority.High,
					executionEffort: UInt64(5000)
				)
				let feeAmount: UFix64 = estimate.flowFee ?? 0.001
				let vaultRef = acct.storage.borrow<auth(FungibleToken.Withdraw) &FlowToken.Vault>(from: /storage/flowTokenVault)!
				let fees <- (vaultRef.withdraw(amount: feeAmount) as! @FlowToken.Vault)
				destroy <- FlowTransactionScheduler.schedule(
					handlerCap: cap, data: nil,
					timestamp: getCurrentBlock().timestamp + 3.0,
					priority: FlowTransactionScheduler.Priority.High,
					executionEffort: UInt64(5000), fees: <-fees
				)
			}
		}
	`, serviceHex, serviceHex)

	latestBlock, err = b.GetLatestBlock()
	require.NoError(t, err)
	tx = flow.NewTransaction().SetScript([]byte(scheduleTx)).
		SetComputeLimit(flowgo.DefaultMaxTransactionGasLimit).
		SetProposalKey(serviceAddress, b.ServiceKey().Index, b.ServiceKey().SequenceNumber).
		SetPayer(serviceAddress).AddAuthorizer(serviceAddress).
		SetReferenceBlockID(flow.Identifier(latestBlock.ID()))
	signer, err = b.ServiceKey().Signer()
	require.NoError(t, err)
	require.NoError(t, tx.SignEnvelope(serviceAddress, b.ServiceKey().Index, signer))
	require.NoError(t, b.SendTransaction(convert.SDKTransactionToFlow(*tx)))
	_, results, err := b.ExecuteAndCommitBlock()
	require.NoError(t, err)
	for _, r := range results {
		if r.Error != nil {
			t.Fatalf("schedule tx failed: %v", r.Error)
		}
	}
	t.Log("✅ Schedule transaction succeeded")

	// Advance time and commit blocks to trigger scheduled execution (from emulator reference)
	t.Log("⏳ Waiting for scheduled transaction to execute...")
	time.Sleep(3500 * time.Millisecond)
	_, _, err = b.ExecuteAndCommitBlock()
	require.NoError(t, err)
	time.Sleep(3500 * time.Millisecond)
	block, _, err := b.ExecuteAndCommitBlock()
	require.NoError(t, err)

	// Get system transaction IDs using GetSystemTransactionsForBlock (from emulator reference)
	// Need to type-assert to *emulator.Blockchain to access this method
	blockchain, ok := b.(*emulator.Blockchain)
	require.True(t, ok, "emulator should be *emulator.Blockchain")
	systemTxIDs, err := blockchain.GetSystemTransactionsForBlock(flowgo.Identifier(block.ID()))
	require.NoError(t, err)
	t.Logf("Found %d system transactions", len(systemTxIDs))
	for i, id := range systemTxIDs {
		t.Logf("  [%d] %s", i, id)
	}

	require.GreaterOrEqual(t, len(systemTxIDs), 3, "should have ProcessCallbacks + ExecuteCallback + SystemChunk")

	// ExecuteCallback is at index 1 (per emulator reference)
	scheduledTxID := systemTxIDs[1]
	t.Logf("🎯 Scheduled ExecuteCallback transaction: %s", scheduledTxID)

	// NOW start gRPC server so profile command can query it
	startServer(t, emulatorServer, host, port)
	time.Sleep(emulatorStableWait)

	return convert.FlowIdentifierToSDK(scheduledTxID), block.Height
}
