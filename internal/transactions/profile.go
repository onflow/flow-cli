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
	"bytes"
	"context"
	"fmt"
	"strings"

	"github.com/onflow/cadence/runtime"
	"github.com/onflow/flow-go/fvm"
	"github.com/onflow/flow-go/fvm/environment"
	reusableRuntime "github.com/onflow/flow-go/fvm/runtime"
	fvmStorage "github.com/onflow/flow-go/fvm/storage"
	fvmState "github.com/onflow/flow-go/fvm/storage/state"
	flowgo "github.com/onflow/flow-go/model/flow"
	"github.com/rs/zerolog"
	"github.com/spf13/cobra"

	"github.com/onflow/flow-emulator/convert"
	"github.com/onflow/flow-emulator/storage/remote"
	"github.com/onflow/flow-emulator/storage/sqlite"
	flowsdk "github.com/onflow/flow-go-sdk"

	"github.com/onflow/flowkit/v2"
	"github.com/onflow/flowkit/v2/config"
	"github.com/onflow/flowkit/v2/output"

	"github.com/onflow/flow-cli/internal/command"
	"github.com/onflow/flow-cli/internal/util"
)

const (
	minProfileableBlockHeight = 2 // Cannot fork from genesis (0) or block 1
	profileFilePrefix         = "profile-"
	profileFileSuffix         = ".pb.gz"
	txIDDisplayLength         = 8
)

type flagsProfile struct {
	Output string `default:"" flag:"output,o" info:"Output file path for profile data (default: profile-{tx_id}.pb.gz)"`
}

type profilingResult struct {
	txID            flowsdk.Identifier
	tx              *flowsdk.Transaction
	result          *flowsdk.TransactionResult
	networkName     string
	blockHeight     uint64
	profileFile     string
	computationUsed uint64
}

var profileFlags = flagsProfile{}

var profileCommand = &command.Command{
	Cmd: &cobra.Command{
		Use:     "profile <tx_id>",
		Short:   "Profile a transaction's execution",
		Example: "flow transactions profile 07a8...b433 -n mainnet",
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
	inputTxID := flowsdk.HexToID(strings.TrimPrefix(args[0], "0x"))

	networkName := globalFlags.Network
	if networkName == "" {
		return nil, fmt.Errorf("network must be specified with --network flag")
	}

	network, err := state.Networks().ByName(networkName)
	if err != nil {
		return nil, fmt.Errorf("network %q not found in flow.json", networkName)
	}

	logger.StartProgress(fmt.Sprintf("Fetching transaction %s from %s...", inputTxID.String(), networkName))

	tx, result, err := flow.GetTransactionByID(context.Background(), inputTxID, true)
	if err != nil {
		logger.StopProgress()
		return nil, fmt.Errorf("failed to get transaction %s: %w", inputTxID.String(), err)
	}

	txID := tx.ID()

	if result.Status != flowsdk.TransactionStatusSealed {
		logger.StopProgress()
		return nil, fmt.Errorf("transaction is not sealed (status: %s)", result.Status)
	}

	logger.Info(fmt.Sprintf("✓ Transaction found in block %d", result.BlockHeight))

	block, err := flow.GetBlock(context.Background(), flowkit.BlockQuery{Height: result.BlockHeight})
	if err != nil {
		logger.StopProgress()
		return nil, fmt.Errorf("failed to get block at height %d: %w", result.BlockHeight, err)
	}

	allTxs, _, err := flow.GetTransactionsByBlockID(context.Background(), block.ID)
	if err != nil {
		logger.StopProgress()
		return nil, fmt.Errorf("failed to get transactions for block %s: %w", block.ID.String(), err)
	}

	targetIdx := findTransactionIndex(allTxs, txID)
	if targetIdx == -1 {
		logger.StopProgress()
		return nil, fmt.Errorf("target transaction %s not found in block %d", txID.String()[:txIDDisplayLength], block.Height)
	}

	targetTx := allTxs[targetIdx]
	isSystemTx := isSystemTransaction(targetTx)
	priorUserTxs, priorSystemTxs := separateTransactionsByType(allTxs[:targetIdx])

	totalPrior := len(priorUserTxs) + len(priorSystemTxs)
	if totalPrior > 0 {
		logger.StartProgress(fmt.Sprintf("Forking state from block %d and replaying %d transactions...", block.Height-1, totalPrior))
	} else {
		logger.StartProgress(fmt.Sprintf("Forking state from block %d...", block.Height-1))
	}

	profile, computationUsed, err := profileTransactionWithFVM(
		state,
		network,
		block,
		priorUserTxs,
		priorSystemTxs,
		targetTx,
		isSystemTx,
		logger,
	)
	if err != nil {
		logger.StopProgress()
		return nil, err
	}

	logger.StopProgress()
	logger.Info("✓ Transaction profiled successfully")

	outputPath := profileFlags.Output
	if outputPath == "" {
		outputPath = fmt.Sprintf("%s%s%s", profileFilePrefix, txID.String()[:txIDDisplayLength], profileFileSuffix)
	}

	if err := writePprofBinary(profile, outputPath, state.ReaderWriter()); err != nil {
		return nil, fmt.Errorf("failed to write profile: %w", err)
	}

	return &profilingResult{
		txID:            txID,
		tx:              tx,
		result:          result,
		networkName:     networkName,
		blockHeight:     result.BlockHeight,
		profileFile:     outputPath,
		computationUsed: computationUsed,
	}, nil
}

func (r *profilingResult) JSON() any {
	return map[string]any{
		"transactionId":   r.txID.String(),
		"network":         r.networkName,
		"block_height":    r.blockHeight,
		"status":          r.result.Status.String(),
		"events":          len(r.result.Events),
		"profileFile":     r.profileFile,
		"computationUsed": r.computationUsed,
	}
}

func (r *profilingResult) String() string {
	var b strings.Builder

	b.WriteString("Transaction Profiling Report\n")
	b.WriteString("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n\n")
	b.WriteString(fmt.Sprintf("Transaction ID:  %s\n", r.txID.String()))
	b.WriteString(fmt.Sprintf("Network:         %s\n", r.networkName))
	b.WriteString(fmt.Sprintf("Block Height:    %d\n", r.blockHeight))
	b.WriteString(fmt.Sprintf("Status:          %s\n", r.result.Status.String()))
	b.WriteString(fmt.Sprintf("Events emitted:  %d\n", len(r.result.Events)))
	b.WriteString(fmt.Sprintf("Computation:     %d\n\n", r.computationUsed))

	b.WriteString(fmt.Sprintf("Profile saved: %s\n\n", r.profileFile))
	b.WriteString("Analyze with:\n")
	b.WriteString(fmt.Sprintf("  go tool pprof -http=:8080 %s\n", r.profileFile))

	return b.String()
}

func (r *profilingResult) Oneliner() string {
	return fmt.Sprintf("Transaction %s profiled successfully", r.txID.String()[:txIDDisplayLength])
}

func profileTransactionWithFVM(
	state *flowkit.State,
	network *config.Network,
	block *flowsdk.Block,
	priorUserTxs []*flowsdk.Transaction,
	priorSystemTxs []*flowsdk.Transaction,
	targetTx *flowsdk.Transaction,
	isSystemTx bool,
	logger output.Logger,
) (*runtime.ComputationProfile, uint64, error) {
	chainID, err := util.GetChainIDFromHost(network.Host)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get chain ID from host %s: %w", network.Host, err)
	}

	blockHeight := block.Height
	forkHeight := blockHeight - 1
	if blockHeight < minProfileableBlockHeight {
		return nil, 0, fmt.Errorf("cannot profile transactions in genesis or block 1 (no prior state to fork from)")
	}

	nopLogger := zerolog.Nop()
	baseStore, err := sqlite.New(sqlite.InMemory)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to create storage: %w", err)
	}

	store, err := remote.New(baseStore, &nopLogger,
		remote.WithForkHost(network.Host),
		remote.WithForkHeight(forkHeight),
	)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to create forked storage at height %d: %w", forkHeight, err)
	}

	ctx := context.Background()
	baseLedger, err := store.LedgerByHeight(ctx, forkHeight)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get ledger at height %d: %w", forkHeight, err)
	}

	execState := fvmState.NewExecutionState(baseLedger, fvmState.DefaultParameters())

	computationProfile := runtime.NewComputationProfile()
	executionEffortWeights := environment.MainnetExecutionEffortWeights
	computationProfile.WithComputationWeights(executionEffortWeights)

	runtimeConfig := runtime.Config{
		ComputationProfile: computationProfile,
	}
	customRuntimePool := reusableRuntime.NewCustomReusableCadenceRuntimePool(
		1,
		runtimeConfig,
		func(cfg runtime.Config) runtime.Runtime {
			return runtime.NewRuntime(cfg)
		},
	)

	vm := fvm.NewVirtualMachine()

	// Create block header for FVM context (enables getCurrentBlock() for scheduled transactions)
	blockHeader := &flowgo.Header{
		HeaderBody: flowgo.HeaderBody{
			ChainID:   chainID,
			ParentID:  flowgo.Identifier(block.ParentID),
			Height:    block.Height,
			Timestamp: uint64(block.Timestamp.UnixMilli()),
		},
		PayloadHash: flowgo.Identifier(block.ID),
	}

	baseFvmOptions := []fvm.Option{
		fvm.WithLogger(nopLogger),
		fvm.WithChain(chainID.Chain()),
		fvm.WithBlockHeader(blockHeader),
		fvm.WithContractDeploymentRestricted(false),
		fvm.WithComputationLimit(flowgo.DefaultMaxTransactionGasLimit),
		fvm.WithReusableCadenceRuntimePool(customRuntimePool),
	}

	userCtx := fvm.NewContext(append(baseFvmOptions,
		fvm.WithTransactionFeesEnabled(true),
		fvm.WithAuthorizationChecksEnabled(true),
		fvm.WithSequenceNumberCheckAndIncrementEnabled(true),
	)...)

	systemCtx := fvm.NewContext(append(baseFvmOptions,
		fvm.WithTransactionFeesEnabled(false),
		fvm.WithAuthorizationChecksEnabled(false),
		fvm.WithSequenceNumberCheckAndIncrementEnabled(false),
	)...)

	// Execute prior transactions to recreate state
	txIndex := 0
	if len(priorUserTxs) > 0 {
		if err := executeTransactions(vm, userCtx, execState, priorUserTxs, txIndex, logger); err != nil {
			return nil, 0, fmt.Errorf("failed to execute prior user transactions: %w", err)
		}
		txIndex += len(priorUserTxs)
	}

	if len(priorSystemTxs) > 0 {
		if err := executeTransactions(vm, systemCtx, execState, priorSystemTxs, txIndex, logger); err != nil {
			return nil, 0, fmt.Errorf("failed to execute prior system transactions: %w", err)
		}
	}

	computationProfile.Reset()

	targetFlowTx := convert.SDKTransactionToFlow(*targetTx)

	targetCtx := userCtx
	if isSystemTx {
		targetCtx = systemCtx
	}

	blockDB := fvmStorage.NewBlockDatabase(execState, 0, nil)
	txn, err := blockDB.NewTransaction(0, fvmState.DefaultParameters())
	if err != nil {
		return nil, 0, fmt.Errorf("failed to create transaction context: %w", err)
	}

	targetTxIndex := uint32(len(priorUserTxs) + len(priorSystemTxs))
	txProc := fvm.Transaction(targetFlowTx, targetTxIndex)
	_, output, err := vm.Run(targetCtx, txProc, txn)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to execute target transaction: %w", err)
	}

	if output.Err != nil {
		logger.Info(fmt.Sprintf("⚠️  Transaction failed during execution: %s", output.Err.Error()))
	}

	return computationProfile, output.ComputationUsed, nil
}

// findTransactionIndex returns the index of a transaction in a slice, or -1 if not found
func findTransactionIndex(txs []*flowsdk.Transaction, txID flowsdk.Identifier) int {
	for i, tx := range txs {
		if tx.ID() == txID {
			return i
		}
	}
	return -1
}

// isSystemTransaction returns true if the transaction is a system transaction
func isSystemTransaction(tx *flowsdk.Transaction) bool {
	return tx.Payer == flowsdk.EmptyAddress
}

// separateTransactionsByType separates transactions into user and system transactions
func separateTransactionsByType(txs []*flowsdk.Transaction) (user, system []*flowsdk.Transaction) {
	user = make([]*flowsdk.Transaction, 0, len(txs))
	system = make([]*flowsdk.Transaction, 0, len(txs))

	for _, tx := range txs {
		if isSystemTransaction(tx) {
			system = append(system, tx)
		} else {
			user = append(user, tx)
		}
	}
	return user, system
}

// executeTransactions executes a list of transactions and updates the execution state
func executeTransactions(
	vm *fvm.VirtualMachine,
	ctx fvm.Context,
	execState *fvmState.ExecutionState,
	txs []*flowsdk.Transaction,
	startIndex int,
	logger output.Logger,
) error {
	for i, tx := range txs {
		flowTx := convert.SDKTransactionToFlow(*tx)

		blockDB := fvmStorage.NewBlockDatabase(execState, 0, nil)
		txn, err := blockDB.NewTransaction(0, fvmState.DefaultParameters())
		if err != nil {
			return fmt.Errorf("failed to create transaction context for tx %d: %w", startIndex+i, err)
		}

		txProc := fvm.Transaction(flowTx, uint32(startIndex+i))
		executionSnapshot, _, err := vm.Run(ctx, txProc, txn)
		if err != nil {
			return fmt.Errorf("failed to execute transaction %d (%s): %w", startIndex+i, tx.ID().String()[:txIDDisplayLength], err)
		}

		if err := execState.Merge(executionSnapshot); err != nil {
			return fmt.Errorf("failed to merge execution snapshot for tx %d: %w", startIndex+i, err)
		}
	}

	return nil
}

// writePprofBinary writes a computation profile to a pprof binary file
func writePprofBinary(profile *runtime.ComputationProfile, outputPath string, rw flowkit.ReaderWriter) error {
	if profile == nil {
		return fmt.Errorf("no profiling data available: profile is nil")
	}

	exporter := runtime.NewPProfExporter(profile)
	pprofData, err := exporter.Export()
	if err != nil {
		return fmt.Errorf("failed to export pprof data: %w", err)
	}

	if pprofData == nil {
		return fmt.Errorf("pprof data is nil after export")
	}

	var buf bytes.Buffer
	if err := pprofData.Write(&buf); err != nil {
		return fmt.Errorf("failed to write pprof data: %w", err)
	}

	if err := rw.WriteFile(outputPath, buf.Bytes(), 0644); err != nil {
		return fmt.Errorf("failed to create output file %s: %w", outputPath, err)
	}

	return nil
}
