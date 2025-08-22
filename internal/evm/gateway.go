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

package evm

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/rs/zerolog/log"

	gethCommon "github.com/ethereum/go-ethereum/common"
	"github.com/onflow/flow-evm-gateway/bootstrap"
	"github.com/onflow/flow-evm-gateway/config"
	flowsdk "github.com/onflow/flow-go-sdk"
	"github.com/onflow/flow-go-sdk/crypto"
	"github.com/onflow/flow-go/fvm/evm/types"
	flowGo "github.com/onflow/flow-go/model/flow"
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
	DatabaseDir       string `flag:"database-dir" default:"./db" info:"path to the directory for the database"`
	RPCHost           string `flag:"rpc-host" default:"localhost" info:"host for the RPC API server"`
	RPCPort           int    `flag:"rpc-port" default:"3000" info:"port for the RPC API server"`
	AccessNodeHost    string `flag:"access-node-host" default:"localhost:3569" info:"host to the flow access node gRPC API"`
	InitCadenceHeight uint64 `flag:"init-cadence-height" default:"0" info:"init cadence block height from where the event ingestion will start. WARNING: you should only provide this if there are no existing values in the database, otherwise an error will be thrown"`
	EVMNetworkID      string `flag:"evm-network-id" default:"testnet" info:"EVM network ID (testnet, mainnet)"`
	FlowNetworkID     string `flag:"flow-network-id" default:"emulator" info:"EVM network ID (emulator, testnet, mainnet)"`
	Coinbase          string `flag:"coinbase" default:"" info:"coinbase address to use for fee collection"`
	GasPrice          string `flag:"gas-price" default:"1" info:"static gas price used for EVM transactions"`
	COAAddress        string `flag:"coa-address" default:"" info:"Flow address that holds COA account used for submitting transactions"`
	COAKey            string `flag:"coa-key" default:"" info:"WARNING: do not use this flag in production! private key value for the COA address used for submitting transactions"`
	CreateCOAResource bool   `flag:"coa-resource-create" default:"false" info:"auto-create the COA resource in the Flow COA account provided if one doesn't exist"`
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
		cfg := config.Config{
			DatabaseDir:       flagGateway.DatabaseDir,
			AccessNodeHost:    flagGateway.AccessNodeHost,
			RPCPort:           flagGateway.RPCPort,
			RPCHost:           flagGateway.RPCHost,
			InitCadenceHeight: flagGateway.InitCadenceHeight,
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

		switch flagGateway.FlowNetworkID {
		case "mainnet":
			cfg.FlowNetworkID = flowGo.Mainnet
		case "testnet":
			cfg.FlowNetworkID = flowGo.Testnet
		case "emulator":
			cfg.FlowNetworkID = flowGo.Emulator
		default:
			return nil, fmt.Errorf("flow network ID not supported, only possible to specify emulator, testnet, mainnet")
		}

		switch flagGateway.EVMNetworkID {
		case "testnet":
			cfg.EVMNetworkID = types.FlowEVMTestNetChainID
		case "mainnet":
			cfg.EVMNetworkID = types.FlowEVMMainNetChainID
		default:
			return nil, fmt.Errorf("EVM network ID not supported")
		}

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		done := make(chan struct{})
		ready := make(chan struct{})
		once := sync.Once{}
		closeReady := func() {
			once.Do(func() {
				close(ready)
			})
		}
		go func() {
			defer close(done)
			// In case an error happens before ready is called we need to close the ready channel
			defer closeReady()

			err := bootstrap.Run(
				ctx,
				cfg,
				closeReady,
			)
			if err != nil && !errors.Is(err, context.Canceled) {
				log.Err(err).Msg("Gateway runtime error")
			}
		}()

		<-ready

		osSig := make(chan os.Signal, 1)
		signal.Notify(osSig, syscall.SIGINT, syscall.SIGTERM)

		// wait for gateway to exit or for a shutdown signal
		select {
		case <-osSig:
			log.Info().Msg("OS Signal to shutdown received, shutting down")
			cancel()
		case <-done:
			log.Info().Msg("done, shutting down")
		}

		// Wait for the gateway to completely stop
		<-done

		return nil, nil
	},
}
