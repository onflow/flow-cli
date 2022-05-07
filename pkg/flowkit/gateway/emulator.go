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

package gateway

import (
	"context"

	"github.com/onflow/cadence"
	emulator "github.com/onflow/flow-emulator"
	"github.com/onflow/flow-emulator/convert/sdk"

	"github.com/onflow/flow-emulator/server/backend"
	"github.com/onflow/flow-go-sdk"
	"github.com/onflow/flow-go-sdk/client"
	"github.com/onflow/flow-go-sdk/client/convert"
	flowGo "github.com/onflow/flow-go/model/flow"
	"github.com/sirupsen/logrus"

	"github.com/onflow/flow-cli/pkg/flowkit"
	"github.com/onflow/flow-cli/pkg/flowkit/config"
)

type EmulatorGateway struct {
	emulator *emulator.Blockchain
	backend  *backend.Backend
	ctx      context.Context
}

func NewEmulatorGateway(serviceAccount *flowkit.Account) *EmulatorGateway {
	return NewEmulatorGatewayWithLogger(logrus.New(), serviceAccount)
}

func NewEmulatorGatewayWithLogger(logger *logrus.Logger, serviceAccount *flowkit.Account) *EmulatorGateway {
	emulator := newEmulator(serviceAccount)
	backend := backend.New(logger, emulator)
	backend.EnableAutoMine()

	return &EmulatorGateway{
		emulator: emulator,
		backend:  backend,
		ctx:      context.Background(),
	}
}

func (g *EmulatorGateway) SetContext(ctx context.Context) {
	g.ctx = ctx
}

func newEmulator(serviceAccount *flowkit.Account) *emulator.Blockchain {
	var opts []emulator.Option
	if serviceAccount != nil && serviceAccount.Key().Type() == config.KeyTypeHex {
		privKey, _ := serviceAccount.Key().PrivateKey()

		opts = append(opts, emulator.WithServicePublicKey(
			(*privKey).PublicKey(),
			serviceAccount.Key().SigAlgo(),
			serviceAccount.Key().HashAlgo(),
		))
	}

	b, err := emulator.NewBlockchain(opts...)
	if err != nil {
		panic(err)
	}

	return b
}

func (g *EmulatorGateway) GetAccount(address flow.Address) (*flow.Account, error) {
	return g.backend.GetAccount(g.ctx, address)
}

func (g *EmulatorGateway) SendSignedTransaction(tx *flowkit.Transaction) (*flow.Transaction, error) {
	err := g.backend.SendTransaction(context.Background(), *tx.FlowTransaction())
	if err != nil {
		return nil, err
	}
	return tx.FlowTransaction(), nil
}

func (g *EmulatorGateway) GetTransactionResult(tx *flow.Transaction, waitSeal bool) (*flow.TransactionResult, error) {
	return g.backend.GetTransactionResult(g.ctx, tx.ID())
}

func (g *EmulatorGateway) GetTransaction(id flow.Identifier) (*flow.Transaction, error) {
	return g.backend.GetTransaction(g.ctx, id)
}

func (g *EmulatorGateway) Ping() error {
	return g.backend.Ping(g.ctx)
}

func (g *EmulatorGateway) ExecuteScript(script []byte, arguments []cadence.Value) (cadence.Value, error) {

	args, err := convert.CadenceValuesToMessages(arguments)
	if err != nil {
		return nil, err
	}

	result, err := g.backend.ExecuteScriptAtLatestBlock(g.ctx, script, args)
	if err != nil {
		return nil, err
	}

	value, err := convert.MessageToCadenceValue(result)
	if err != nil {
		return nil, err
	}

	return value, nil
}

func (g *EmulatorGateway) GetLatestBlock() (*flow.Block, error) {
	block, err := g.backend.GetLatestBlock(g.ctx, true)
	if err != nil {
		return nil, err
	}

	return convertBlock(block), nil
}

func convertBlock(block *flowGo.Block) *flow.Block {
	return &flow.Block{
		BlockHeader: flow.BlockHeader{
			ID:        flow.Identifier(block.Header.ID()),
			ParentID:  flow.Identifier(block.Header.ParentID),
			Height:    block.Header.Height,
			Timestamp: block.Header.Timestamp,
		},
		BlockPayload: flow.BlockPayload{
			CollectionGuarantees: nil,
			Seals:                nil,
		},
	}
}

func (g *EmulatorGateway) GetEvents(
	eventType string,
	startHeight uint64,
	endHeight uint64,
) ([]client.BlockEvents, error) {
	events := make([]client.BlockEvents, 0)

	for height := startHeight; height <= endHeight; height++ {
		events = append(events, g.getBlockEvent(height, eventType))
	}

	return events, nil
}

func (g *EmulatorGateway) getBlockEvent(height uint64, eventType string) client.BlockEvents {
	block, _ := g.backend.GetBlockByHeight(g.ctx, height)
	events, _ := g.backend.GetEventsForBlockIDs(g.ctx, eventType, []flow.Identifier{flow.Identifier(block.ID())})

	result := client.BlockEvents{
		BlockID:        flow.Identifier(block.ID()),
		Height:         uint64(block.Header.Height),
		BlockTimestamp: block.Header.Timestamp,
		Events:         []flow.Event{},
	}

	for _, e := range events {
		if e.BlockID == block.ID() {
			result.Events, _ = sdk.FlowEventsToSDK(e.Events) //convertEvents(e.Events)
			return result
		}
	}

	return result
}

func (g *EmulatorGateway) GetCollection(id flow.Identifier) (*flow.Collection, error) {
	return g.backend.GetCollectionByID(g.ctx, id)
}

func (g *EmulatorGateway) GetBlockByID(id flow.Identifier) (*flow.Block, error) {
	block, err := g.backend.GetBlockByID(g.ctx, id)
	if err != nil {
		return nil, err
	}
	return convertBlock(block), nil
}

func (g *EmulatorGateway) GetBlockByHeight(height uint64) (*flow.Block, error) {
	block, err := g.backend.GetBlockByHeight(g.ctx, height)
	if err != nil {
		return nil, err
	}
	return convertBlock(block), nil
}

func (g *EmulatorGateway) GetLatestProtocolStateSnapshot() ([]byte, error) {
	return g.backend.GetLatestProtocolStateSnapshot(g.ctx)
}

// SecureConnection placeholder func to complete gateway interface implementation
func (g *EmulatorGateway) SecureConnection() bool {
	return false
}
