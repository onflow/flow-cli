package evm

import (
	"context"
	_ "embed"
	"fmt"
	"math/big"
	"net/http"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/spf13/cobra"

	"github.com/onflow/flow-cli/flowkit"
	"github.com/onflow/flow-cli/flowkit/output"
	"github.com/onflow/flow-cli/internal/command"
)

//go:embed run.cdc
var callCode []byte

type flagsRPC struct{}

var rpcFlags = flagsRPC{}

var rpcCommand = &command.Command{
	Cmd: &cobra.Command{
		Use:     "rpc",
		Short:   "Start the RPC ethereum server",
		Args:    cobra.ExactArgs(0),
		Example: "flow rpc",
	},
	Flags: &rpcFlags,
	RunS:  rpcRun,
}

// todo only for demo, super hacky now

func rpcRun(
	args []string,
	_ command.GlobalFlags,
	_ output.Logger,
	flow flowkit.Services,
	state *flowkit.State,
) (command.Result, error) {

	server := rpc.NewServer()
	err := server.RegisterName("eth", &ethAPI{flow})
	if err != nil {
		return nil, err
	}

	http.Handle("/", server)
	err = http.ListenAndServe(":9000", nil)
	if err != nil {
		return nil, err
	}

	server.Stop()
	return nil, nil
}

type ethAPI struct {
	flow flowkit.Services
}

func (e *ethAPI) Call(
	ctx context.Context,
	args TransactionArgs,
	blockNumberOrHash *rpc.BlockNumberOrHash,
	overrides *StateOverride,
	blockOverrides *BlockOverrides,
) (hexutil.Bytes, error) {
	val, err := executeCall(
		e.flow,
		strings.ReplaceAll(args.To.String(), "0x", ""),
		"f8d6e0586b0a20c7",
		*args.Data,
	)

	fmt.Println("result", val, err)
	return val, err
}

func (e *ethAPI) Ping() (int, error) {
	return 1, nil
}

func (e *ethAPI) BlockNumber() hexutil.Uint64 {
	return hexutil.Uint64(65848272)
}

func (s *ethAPI) GetBlockByNumber(
	ctx context.Context,
	blockNumber rpc.BlockNumber,
	fullTx bool,
) (map[string]interface{}, error) {
	return map[string]interface{}{}, nil
}

func (s *ethAPI) GasPrice(ctx context.Context) (*hexutil.Big, error) {
	return (*hexutil.Big)(big.NewInt(1)), nil
}

type TransactionArgs struct {
	From                 *common.Address `json:"from"`
	To                   *common.Address `json:"to"`
	Gas                  *hexutil.Uint64 `json:"gas"`
	GasPrice             *hexutil.Big    `json:"gasPrice"`
	MaxFeePerGas         *hexutil.Big    `json:"maxFeePerGas"`
	MaxPriorityFeePerGas *hexutil.Big    `json:"maxPriorityFeePerGas"`
	Value                *hexutil.Big    `json:"value"`
	Nonce                *hexutil.Uint64 `json:"nonce"`

	// We accept "data" and "input" for backwards-compatibility reasons.
	// "input" is the newer name and should be preferred by clients.
	// Issue detail: https://github.com/ethereum/go-ethereum/issues/15628
	Data  *hexutil.Bytes `json:"data"`
	Input *hexutil.Bytes `json:"input"`

	// Introduced by AccessListTxType transaction.
	AccessList *types.AccessList `json:"accessList,omitempty"`
	ChainID    *hexutil.Big      `json:"chainId,omitempty"`
}

type StateOverride map[common.Address]OverrideAccount

type BlockOverrides struct {
	Number      *hexutil.Big
	Difficulty  *hexutil.Big
	Time        *hexutil.Uint64
	GasLimit    *hexutil.Uint64
	Coinbase    *common.Address
	Random      *common.Hash
	BaseFee     *hexutil.Big
	BlobBaseFee *hexutil.Big
}

type OverrideAccount struct {
	Nonce     *hexutil.Uint64              `json:"nonce"`
	Code      *hexutil.Bytes               `json:"code"`
	Balance   **hexutil.Big                `json:"balance"`
	State     *map[common.Hash]common.Hash `json:"state"`
	StateDiff *map[common.Hash]common.Hash `json:"stateDiff"`
}
