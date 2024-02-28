package evm

import (
	"context"
	"fmt"
	"math/big"
	"os"
	"os/signal"
	"syscall"

	gethCommon "github.com/ethereum/go-ethereum/common"
	"github.com/onflow/flow-evm-gateway/bootstrap"
	"github.com/onflow/flow-evm-gateway/config"
	flowsdk "github.com/onflow/flow-go-sdk"
	"github.com/onflow/flow-go-sdk/crypto"
	"github.com/onflow/flow-go/fvm/evm/emulator"
	"github.com/onflow/flowkit/v2"
	"github.com/onflow/flowkit/v2/output"
	"github.com/spf13/cobra"

	"github.com/onflow/flow-cli/internal/command"
)

var Cmd = &cobra.Command{
	Use:              "evm",
	Short:            "EVM related commands",
	TraverseChildren: true,
}

func init() {
	gatewayCommand.AddToParent(Cmd)
}

type gatewayFlag struct {
	DatabaseDir        string `flag:"database-dir" default:"./db" info:"path to the directory for the database"`
	RPCHost            string `flag:"rpc-host" default:"localhost" info:"host for the RPC API server"`
	RPCPort            int    `flag:"rpc-port" default:"3000" info:"port for the RPC API server"`
	AccessNodeGRPCHost string `flag:"access-node-grpc-host" default:"localhost:3569" info:"host to the flow access node gRPC API"`
	InitCadenceHeight  uint64 `flag:"init-cadence-height" default:"0" info:"init cadence block height from where the event ingestion will start. WARNING: you should only provide this if there are no existing values in the database, otherwise an error will be thrown"`
	EVMNetworkID       string `flag:"evm-network-id" default:"testnet" info:"EVM network ID (testnet, mainnet)"`
	FlowNetworkID      string `flag:"flow-network-id" default:"emulator" info:"EVM network ID (emulator, previewnet)"`
	Coinbase           string `flag:"coinbase" default:"" info:"coinbase address to use for fee collection"`
	GasPrice           string `flag:"gas-price" default:"1" info:"static gas price used for EVM transactions"`
	COAAddress         string `flag:"coa-address" default:"" info:"Flow address that holds COA account used for submitting transactions"`
	COAKey             string `flag:"coa-key" default:"" info:"WARNING: do not use this flag in production! private key value for the COA address used for submitting transactions"`
	CreateCOAResource  bool   `flag:"coa-resource-create" default:"false" info:"auto-create the COA resource in the Flow COA account provided if one doesn't exist"`
}

var flagGateway = gatewayFlag{}

var gatewayCommand = &command.Command{
	Cmd: &cobra.Command{
		Use:     "gateway",
		Short:   "Start the EVM gateway that exposes EVM RPC APIs",
		Example: "flow evm gateway",
	},
	Flags: &flagGateway,
	Run: func(
		args []string,
		globalFlags command.GlobalFlags,
		logger output.Logger,
		readerWriter flowkit.ReaderWriter,
		flow flowkit.Services,
	) (command.Result, error) {
		cfg := &config.Config{
			DatabaseDir:        flagGateway.DatabaseDir,
			AccessNodeGRPCHost: flagGateway.AccessNodeGRPCHost,
			RPCPort:            flagGateway.RPCPort,
			RPCHost:            flagGateway.RPCHost,
			InitCadenceHeight:  flagGateway.InitCadenceHeight,
			CreateCOAResource:  flagGateway.CreateCOAResource,
		}

		if flagGateway.Coinbase == "" {
			return nil, fmt.Errorf("coinbase EVM address required")
		}
		cfg.Coinbase = gethCommon.HexToAddress(flagGateway.Coinbase)
		if g, ok := new(big.Int).SetString(flagGateway.GasPrice, 10); ok {
			cfg.GasPrice = g
		}

		cfg.COAAddress = flowsdk.HexToAddress(flagGateway.COAAddress)
		if cfg.COAAddress == flowsdk.EmptyAddress {
			return nil, fmt.Errorf("invalid COA address value")
		}

		pkey, err := crypto.DecodePrivateKeyHex(crypto.ECDSA_P256, flagGateway.COAKey)
		if err != nil {
			return nil, fmt.Errorf("invalid COA key: %w", err)
		}
		cfg.COAKey = pkey

		cfg.FlowNetworkID = flagGateway.FlowNetworkID

		switch flagGateway.EVMNetworkID {
		case "testnet":
			cfg.EVMNetworkID = emulator.FlowEVMTestnetChainID
		case "mainnet":
			cfg.EVMNetworkID = emulator.FlowEVMMainnetChainID
		default:
			return nil, fmt.Errorf("EVM network ID not supported")
		}

		if cfg.FlowNetworkID != "previewnet" && cfg.FlowNetworkID != "emulator" {
			return nil, fmt.Errorf("flow network ID is invalid, only allowed to set 'emulator' and 'previewnet'")
		}

		ctx, cancel := context.WithCancel(context.Background())

		err = bootstrap.Start(ctx, cfg)
		if err != nil {
			panic(err)
		}

		osSig := make(chan os.Signal, 1)
		signal.Notify(osSig, syscall.SIGINT, syscall.SIGTERM)

		<-osSig
		fmt.Println("OS Signal to shutdown received, shutting down")
		cancel()

		return nil, nil
	},
}
