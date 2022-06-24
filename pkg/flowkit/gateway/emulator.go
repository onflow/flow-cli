/*
 * Flow CLI
 *
 * Copyright 2019 Dapper Labs, Inc.
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
	"fmt"

	"github.com/onflow/cadence"
	jsoncdc "github.com/onflow/cadence/encoding/json"
	emulator "github.com/onflow/flow-emulator"
	"github.com/onflow/flow-emulator/convert/sdk"
	"github.com/rs/zerolog"
	"google.golang.org/grpc/status"

	"github.com/onflow/flow-emulator/server/backend"
	"github.com/onflow/flow-go-sdk"
	flowGo "github.com/onflow/flow-go/model/flow"
	"github.com/sirupsen/logrus"

	"github.com/onflow/flow-cli/pkg/flowkit"
	"github.com/onflow/flow-cli/pkg/flowkit/config"
)

type EmulatorGateway struct {
	emulator      *emulator.Blockchain
	backend       *backend.Backend
	ctx           context.Context
	logger        *logrus.Logger
	emultorLogger *zerolog.Logger
}

func UnwrapStatusError(err error) error {
	return fmt.Errorf(status.Convert(err).Message())
}

func NewEmulatorGateway(serviceAccount *flowkit.Account) *EmulatorGateway {
	return NewEmulatorGatewayWithOpts(serviceAccount)
}

func NewEmulatorGatewayWithOpts(serviceAccount *flowkit.Account, opts ...func(*EmulatorGateway)) *EmulatorGateway {

	defaultEmulatorLogger := zerolog.Nop()
	gateway := &EmulatorGateway{
		ctx:           context.Background(),
		logger:        logrus.New(),
		emultorLogger: &defaultEmulatorLogger,
	}
	for _, opt := range opts {
		opt(gateway)
	}
	gateway.emulator = newEmulator(serviceAccount, gateway.emultorLogger)
	gateway.backend = backend.New(gateway.logger, gateway.emulator)
	gateway.backend.EnableAutoMine()

	return gateway
}

func WithLogger(logger *logrus.Logger) func(g *EmulatorGateway) {
	return func(g *EmulatorGateway) {
		g.logger = logger
	}
}

func WithEmulatorLogger(logger *zerolog.Logger) func(g *EmulatorGateway) {
	return func(g *EmulatorGateway) {
		g.emultorLogger = logger
	}
}

func WithContext(ctx context.Context) func(g *EmulatorGateway) {
	return func(g *EmulatorGateway) {
		g.ctx = ctx
	}
}

func (g *EmulatorGateway) SetContext(ctx context.Context) {
	g.ctx = ctx
}

func newEmulator(serviceAccount *flowkit.Account, emulatorLogger *zerolog.Logger) *emulator.Blockchain {
	var opts []emulator.Option
	if serviceAccount != nil && serviceAccount.Key().Type() == config.KeyTypeHex {
		privKey, _ := serviceAccount.Key().PrivateKey()

		opts = append(opts, emulator.WithServicePublicKey(
			(*privKey).PublicKey(),
			serviceAccount.Key().SigAlgo(),
			serviceAccount.Key().HashAlgo(),
		), emulator.WithLogger(*emulatorLogger))
	}

	b, err := emulator.NewBlockchain(opts...)
	if err != nil {
		panic(err)
	}

	return b
}

func (g *EmulatorGateway) GetAccount(address flow.Address) (*flow.Account, error) {
	account, err := g.backend.GetAccount(g.ctx, address)
	if err != nil {
		return nil, UnwrapStatusError(err)
	}
	return account, nil
}

func (g *EmulatorGateway) SendSignedTransaction(tx *flowkit.Transaction) (*flow.Transaction, error) {
	err := g.backend.SendTransaction(context.Background(), *tx.FlowTransaction())
	if err != nil {
		return nil, UnwrapStatusError(err)
	}
	return tx.FlowTransaction(), nil
}

func (g *EmulatorGateway) GetTransactionResult(tx *flow.Transaction, waitSeal bool) (*flow.TransactionResult, error) {
	result, err := g.backend.GetTransactionResult(g.ctx, tx.ID())
	if err != nil {
		return nil, UnwrapStatusError(err)
	}
	return result, nil
}

func (g *EmulatorGateway) GetTransaction(id flow.Identifier) (*flow.Transaction, error) {
	transaction, err := g.backend.GetTransaction(g.ctx, id)
	if err != nil {
		return nil, UnwrapStatusError(err)
	}
	return transaction, nil
}

func (g *EmulatorGateway) Ping() error {
	err := g.backend.Ping(g.ctx)
	if err != nil {
		return UnwrapStatusError(err)
	}
	return nil
}

func (g *EmulatorGateway) ExecuteScript(script []byte, arguments []cadence.Value) (cadence.Value, error) {

	args, err := cadenceValuesToMessages(arguments)
	if err != nil {
		return nil, UnwrapStatusError(err)
	}

	result, err := g.backend.ExecuteScriptAtLatestBlock(g.ctx, script, args)
	if err != nil {
		return nil, UnwrapStatusError(err)
	}

	value, err := messageToCadenceValue(result)
	if err != nil {
		return nil, UnwrapStatusError(err)
	}

	return value, nil
}

func (g *EmulatorGateway) GetLatestBlock() (*flow.Block, error) {
	block, err := g.backend.GetLatestBlock(g.ctx, true)
	if err != nil {
		return nil, UnwrapStatusError(err)
	}

	return convertBlock(block), nil
}

func cadenceValuesToMessages(values []cadence.Value) ([][]byte, error) {
	msgs := make([][]byte, len(values))
	for i, val := range values {
		msg, err := jsoncdc.Encode(val)
		if err != nil {
			return nil, fmt.Errorf("convert: %w", err)
		}
		msgs[i] = msg
	}
	return msgs, nil
}

func messageToCadenceValue(m []byte) (cadence.Value, error) {
	v, err := jsoncdc.Decode(nil, m)
	if err != nil {
		return nil, fmt.Errorf("convert: %w", err)
	}

	return v, nil
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
) ([]flow.BlockEvents, error) {
	events := make([]flow.BlockEvents, 0)

	for height := startHeight; height <= endHeight; height++ {
		events = append(events, g.getBlockEvent(height, eventType))
	}

	return events, nil
}

func (g *EmulatorGateway) getBlockEvent(height uint64, eventType string) flow.BlockEvents {
	block, _ := g.backend.GetBlockByHeight(g.ctx, height)
	events, _ := g.backend.GetEventsForBlockIDs(g.ctx, eventType, []flow.Identifier{flow.Identifier(block.ID())})

	result := flow.BlockEvents{
		BlockID:        flow.Identifier(block.ID()),
		Height:         block.Header.Height,
		BlockTimestamp: block.Header.Timestamp,
		Events:         []flow.Event{},
	}

	for _, e := range events {
		if e.BlockID == block.ID() {
			result.Events, _ = sdk.FlowEventsToSDK(e.Events)
			return result
		}
	}

	return result
}

func (g *EmulatorGateway) GetCollection(id flow.Identifier) (*flow.Collection, error) {
	collection, err := g.backend.GetCollectionByID(g.ctx, id)
	if err != nil {
		return nil, UnwrapStatusError(err)
	}
	return collection, nil
}

func (g *EmulatorGateway) GetBlockByID(id flow.Identifier) (*flow.Block, error) {
	block, err := g.backend.GetBlockByID(g.ctx, id)
	if err != nil {
		return nil, UnwrapStatusError(err)
	}
	return convertBlock(block), nil
}

func (g *EmulatorGateway) GetBlockByHeight(height uint64) (*flow.Block, error) {
	block, err := g.backend.GetBlockByHeight(g.ctx, height)
	if err != nil {
		return nil, UnwrapStatusError(err)
	}
	return convertBlock(block), nil
}

func (g *EmulatorGateway) GetLatestProtocolStateSnapshot() ([]byte, error) {
	snapshot, err := g.backend.GetLatestProtocolStateSnapshot(g.ctx)
	if err != nil {
		return nil, UnwrapStatusError(err)
	}
	return snapshot, nil
}

// SecureConnection placeholder func to complete gateway interface implementation
func (g *EmulatorGateway) SecureConnection() bool {
	return false
}
