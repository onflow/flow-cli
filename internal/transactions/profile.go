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
	"strings"
	"time"

	"github.com/rs/zerolog"
	"github.com/spf13/cobra"

	"github.com/onflow/flow-emulator/convert"
	"github.com/onflow/flow-emulator/emulator"
	"github.com/onflow/flow-emulator/server"
	flowsdk "github.com/onflow/flow-go-sdk"
	flowgo "github.com/onflow/flow-go/model/flow"

	"github.com/onflow/flowkit/v2"
	"github.com/onflow/flowkit/v2/config"
	"github.com/onflow/flowkit/v2/output"

	"github.com/onflow/flow-cli/internal/command"
)

type flagsProfile struct {
	Network string `default:"" flag:"network" info:"Network to profile transaction from"`
}

var profileFlags = flagsProfile{}

var profileCommand = &command.Command{
	Cmd: &cobra.Command{
		Use:     "profile <tx_id>",
		Short:   "Profile a transaction by replaying it on a forked emulator",
		Example: "flow transactions profile 07a8b433... --network mainnet",
		Args:    cobra.ExactArgs(1),
	},
	Flags: &profileFlags,
	RunS:  profile,
}

func profile(
	args []string,
	globalFlags command.GlobalFlags,
	logger output.Logger,
	flow flowkit.Services,
	state *flowkit.State,
) (command.Result, error) {
	txID := flowsdk.HexToID(strings.TrimPrefix(args[0], "0x"))

	networkName := profileFlags.Network
	if networkName == "" {
		networkName = globalFlags.Network
	}
	if networkName == "" {
		return nil, fmt.Errorf("network must be specified with --network flag")
	}

	network, err := state.Networks().ByName(networkName)
	if err != nil {
		return nil, fmt.Errorf("network %q not found in flow.json", networkName)
	}

	logger.StartProgress(fmt.Sprintf("Fetching transaction %s from %s...", txID.String(), networkName))

	tx, result, err := flow.GetTransactionByID(context.Background(), txID, true)
	if err != nil {
		logger.StopProgress()
		return nil, fmt.Errorf("failed to get transaction: %w", err)
	}

	if result.Status != flowsdk.TransactionStatusSealed {
		logger.StopProgress()
		return nil, fmt.Errorf("transaction is not sealed (status: %s)", result.Status)
	}

	logger.Info(fmt.Sprintf("✓ Transaction found in block %d", result.BlockHeight))
	logger.Info("Fetching block and transactions...")

	block, err := flow.GetBlock(context.Background(), flowkit.BlockQuery{Height: result.BlockHeight})
	if err != nil {
		logger.StopProgress()
		return nil, fmt.Errorf("failed to get block: %w", err)
	}

	allTxs, _, err := flow.GetTransactionsByBlockID(context.Background(), block.ID)
	if err != nil {
		logger.StopProgress()
		return nil, fmt.Errorf("failed to get block transactions: %w", err)
	}

	logger.Info(fmt.Sprintf("✓ Found %d transactions in block", len(allTxs)))

	targetIdx := findTransactionIndex(allTxs, txID)
	if targetIdx == -1 {
		logger.StopProgress()
		return nil, fmt.Errorf("transaction not found in block")
	}

	logger.Info("Creating forked emulator blockchain...")
	logger.StopProgress()

	blockchain, cleanup, err := createForkedEmulator(state, network, result.BlockHeight)
	if err != nil {
		return nil, err
	}
	defer cleanup()

	logger.StartProgress("Replaying transactions with profiling...")

	txReport, err := replayTransactions(blockchain, allTxs, targetIdx, logger)
	if err != nil {
		logger.StopProgress()
		return nil, err
	}

	logger.StopProgress()
	logger.Info("✓ Transaction profiled successfully")

	return &profilingResult{
		txID:        txID,
		tx:          tx,
		result:      result,
		txReport:    txReport,
		networkName: networkName,
		blockHeight: result.BlockHeight,
	}, nil
}

func findTransactionIndex(txs []*flowsdk.Transaction, targetID flowsdk.Identifier) int {
	for i, tx := range txs {
		if tx.ID() == targetID {
			return i
		}
	}
	return -1
}

func createForkedEmulator(
	state *flowkit.State,
	network *config.Network,
	blockHeight uint64,
) (emulator.Emulator, func(), error) {
	serviceAccount, err := state.EmulatorServiceAccount()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get emulator service account: %w", err)
	}

	privateKey, err := serviceAccount.Key.PrivateKey()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get service account private key: %w", err)
	}

	chainID := detectChainID(network.Host)
	forkHeight := calculateForkHeight(blockHeight)

	zlog := zerolog.Nop()
	serverConf := &server.Config{
		ForkHost:                    network.Host,
		ForkHeight:                  forkHeight,
		ChainID:                     chainID,
		ComputationReportingEnabled: true,
		ServicePrivateKey:           *privateKey,
		ServiceKeySigAlgo:           serviceAccount.Key.SigAlgo(),
		ServiceKeyHashAlgo:          serviceAccount.Key.HashAlgo(),
		SkipTransactionValidation:   true,
		StorageLimitEnabled:         false,
		Host:                        "127.0.0.1",
		GRPCPort:                    3571,
		RESTPort:                    8889,
		AdminPort:                   8081,
	}

	emulatorServer := server.NewEmulatorServer(&zlog, serverConf)
	if emulatorServer == nil {
		return nil, nil, fmt.Errorf("failed to create emulator server")
	}

	go func() {
		defer func() {
			if r := recover(); r != nil {
				// Silently recover from emulator panics
			}
		}()
		emulatorServer.Start()
	}()

	cleanup := func() {
		if emulatorServer != nil {
			emulatorServer.Stop()
		}
	}

	time.Sleep(2 * time.Second)

	blockchain := emulatorServer.Emulator()
	if blockchain == nil {
		cleanup()
		return nil, nil, fmt.Errorf("failed to get emulator blockchain instance")
	}

	return blockchain, cleanup, nil
}

func detectChainID(host string) flowgo.ChainID {
	if strings.Contains(host, "testnet") || strings.Contains(host, "devnet") {
		return flowgo.Testnet
	}
	if strings.Contains(host, "127.0.0.1") || strings.Contains(host, "localhost") {
		return flowgo.Emulator
	}
	return flowgo.Mainnet
}

func calculateForkHeight(blockHeight uint64) uint64 {
	if blockHeight > 1 {
		return blockHeight - 1
	}
	return 1
}

func replayTransactions(
	blockchain emulator.Emulator,
	txs []*flowsdk.Transaction,
	targetIdx int,
	logger output.Logger,
) (*emulator.ProcedureReport, error) {
	var txReport *emulator.ProcedureReport

	for i := 0; i <= targetIdx; i++ {
		tx := txs[i]
		logger.Info(fmt.Sprintf("  [%d/%d] Replaying transaction %s...", i+1, targetIdx+1, tx.ID().String()[:8]))

		flowGoTx := convert.SDKTransactionToFlow(*tx)
		if err := blockchain.AddTransaction(*flowGoTx); err != nil {
			return nil, fmt.Errorf("failed to add transaction %s: %w", tx.ID(), err)
		}

		_, txResults, err := blockchain.ExecuteAndCommitBlock()
		if err != nil {
			return nil, fmt.Errorf("failed to execute block with transaction %s: %w", tx.ID(), err)
		}

		if i == targetIdx && len(txResults) > 0 {
			computationReport := blockchain.ComputationReport()
			if computationReport != nil {
				if report, ok := computationReport.Transactions[tx.ID().String()]; ok {
					txReport = &report
				}
			}
		}
	}

	return txReport, nil
}

type profilingResult struct {
	txID        flowsdk.Identifier
	tx          *flowsdk.Transaction
	result      *flowsdk.TransactionResult
	txReport    *emulator.ProcedureReport
	networkName string
	blockHeight uint64
}

func (r *profilingResult) JSON() any {
	result := map[string]any{
		"transaction_id": r.txID.String(),
		"network":        r.networkName,
		"block_height":   r.blockHeight,
		"status":         r.result.Status.String(),
		"events_count":   len(r.result.Events),
	}

	if r.result.Error != nil {
		result["error"] = r.result.Error.Error()
	}

	if r.txReport != nil {
		result["computation_metrics"] = map[string]any{
			"computation_used": r.txReport.ComputationUsed,
			"intensities":      r.txReport.Intensities,
			"memory_estimate":  r.txReport.MemoryEstimate,
		}
	}

	return result
}

func (r *profilingResult) String() string {
	var sb strings.Builder

	sb.WriteString("Transaction Profiling Report\n")
	sb.WriteString("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n\n")
	sb.WriteString(fmt.Sprintf("Transaction ID:  %s\n", r.txID.String()))
	sb.WriteString(fmt.Sprintf("Network:         %s\n", r.networkName))
	sb.WriteString(fmt.Sprintf("Block Height:    %d\n", r.blockHeight))
	sb.WriteString(fmt.Sprintf("Status:          %s\n", r.result.Status.String()))

	if r.result.Error != nil {
		sb.WriteString(fmt.Sprintf("\nError: %s\n", r.result.Error.Error()))
	}

	sb.WriteString("\nProfiling Metrics:\n")
	sb.WriteString("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
	sb.WriteString(fmt.Sprintf("Events emitted:  %d\n", len(r.result.Events)))

	if r.txReport != nil {
		sb.WriteString("\nComputation Metrics:\n")
		sb.WriteString(fmt.Sprintf("  Computation used: %d\n", r.txReport.ComputationUsed))
		sb.WriteString(fmt.Sprintf("  Memory estimate:  %d bytes\n", r.txReport.MemoryEstimate))

		if len(r.txReport.Intensities) > 0 {
			sb.WriteString("\nIntensity Breakdown:\n")
			count := 0
			for kind, value := range r.txReport.Intensities {
				if count >= 10 {
					break
				}
				sb.WriteString(fmt.Sprintf("  - %s: %d\n", kind, value))
				count++
			}
		}
	} else {
		sb.WriteString("\nNote: Computation profiling data not available.\n")
	}

	return sb.String()
}

func (r *profilingResult) Oneliner() string {
	return fmt.Sprintf("%s: %s (block %d)", r.txID.String(), r.result.Status.String(), r.blockHeight)
}

func (r *profilingResult) ExitCode() int {
	if r.result.Error != nil {
		return 1
	}
	return 0
}
