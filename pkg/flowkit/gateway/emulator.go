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
	"fmt"
	"time"

	"github.com/onflow/flow-cli/pkg/flowkit"
	"github.com/onflow/flow-cli/pkg/flowkit/config"

	"github.com/onflow/cadence"
	emulator "github.com/onflow/flow-emulator"
	"github.com/onflow/flow-go-sdk"
	"github.com/onflow/flow-go-sdk/client"
	"github.com/onflow/flow-go-sdk/client/convert"
	flowGo "github.com/onflow/flow-go/model/flow"
)

type EmulatorGateway struct {
	emulator *emulator.Blockchain
}

func NewEmulatorGateway(serviceAccount *flowkit.Account) *EmulatorGateway {
	return &EmulatorGateway{
		emulator: newEmulator(serviceAccount),
	}
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
	return g.emulator.GetAccount(address)
}

func (g *EmulatorGateway) SendSignedTransaction(tx *flowkit.Transaction) (*flow.Transaction, error) {
	t := tx.FlowTransaction()
	err := g.emulator.AddTransaction(*t)
	if err != nil {
		return nil, fmt.Errorf("failed to submit transaction: %w", err)
	}

	_, err = g.emulator.ExecuteNextTransaction()
	if err != nil {
		return nil, fmt.Errorf("failed to submit transaction: %w", err)
	}

	_, err = g.emulator.CommitBlock()
	if err != nil {
		return nil, fmt.Errorf("failed to submit transaction: %w", err)
	}

	return t, nil
}

func (g *EmulatorGateway) GetTransactionResult(tx *flow.Transaction, waitSeal bool) (*flow.TransactionResult, error) {
	result, err := g.emulator.GetTransactionResult(tx.ID())
	if err != nil {
		return nil, err
	}

	if result.Status != flow.TransactionStatusSealed && waitSeal {
		time.Sleep(time.Second)
		return g.GetTransactionResult(tx, waitSeal)
	}

	return result, nil
}

func (g *EmulatorGateway) GetTransaction(id flow.Identifier) (*flow.Transaction, error) {
	return g.emulator.GetTransaction(id)
}

func (g *EmulatorGateway) Ping() error {
	return nil
}

func (g *EmulatorGateway) ExecuteScript(script []byte, arguments []cadence.Value) (cadence.Value, error) {
	args, err := convert.CadenceValuesToMessages(arguments)
	if err != nil {
		return nil, err
	}

	result, err := g.emulator.ExecuteScript(script, args)
	if err != nil {
		return nil, err
	}

	if result.Error != nil {
		return nil, result.Error
	}

	return result.Value, nil
}

func (g *EmulatorGateway) GetLatestBlock() (*flow.Block, error) {
	block, err := g.emulator.GetLatestBlock()
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
	events, _ := g.emulator.GetEventsByHeight(height, eventType)
	block, _ := g.emulator.GetBlockByHeight(height)

	flowEvents := make([]flow.Event, 0)

	for _, e := range events {
		flowEvents = append(flowEvents, flow.Event{
			Type:             e.Type,
			TransactionID:    e.TransactionID,
			TransactionIndex: e.TransactionIndex,
			EventIndex:       e.EventIndex,
			Value:            e.Value,
		})
	}

	return client.BlockEvents{
		BlockID:        flow.Identifier(block.Header.ID()),
		Height:         block.Header.Height,
		BlockTimestamp: block.Header.Timestamp,
		Events:         flowEvents,
	}
}

func (g *EmulatorGateway) GetCollection(id flow.Identifier) (*flow.Collection, error) {
	return g.emulator.GetCollection(id)
}

func (g *EmulatorGateway) GetBlockByID(id flow.Identifier) (*flow.Block, error) {
	block, err := g.emulator.GetBlockByID(id)
	return convertBlock(block), err
}

func (g *EmulatorGateway) GetBlockByHeight(height uint64) (*flow.Block, error) {
	block, err := g.emulator.GetBlockByHeight(height)
	return convertBlock(block), err
}
