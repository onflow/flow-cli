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
	"os"
	"path"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/onflow/flow-emulator/convert/sdk"
	"github.com/onflow/flow-emulator/server/backend"

	"github.com/onflow/cadence"
	emulator "github.com/onflow/flow-emulator"
	"github.com/onflow/flow-emulator/storage/badger"
	"github.com/onflow/flow-go-sdk"
	"github.com/onflow/flow-go-sdk/client"
	"github.com/onflow/flow-go-sdk/crypto"
	flowGo "github.com/onflow/flow-go/model/flow"
	"github.com/spf13/afero"

	"github.com/onflow/flow-cli/pkg/flowcli/config"
	"github.com/onflow/flow-cli/pkg/flowcli/project"
	"github.com/onflow/flow-go-sdk/client/convert"
)

type EmulatorGateway struct {
	backend *backend.Backend
	ctx     context.Context
}

func NewEmulatorGateway(serviceAccount *project.Account) (*EmulatorGateway, error) {
	b, err := newBackend(serviceAccount)
	if err != nil {
		return nil, err
	}

	return &EmulatorGateway{
		backend: b,
		ctx:     context.Background(), // todo refactor
	}, nil
}

func newBackend(serviceAccount *project.Account) (*backend.Backend, error) {
	var opts []emulator.Option
	if serviceAccount != nil && serviceAccount.DefaultKey().Type() == config.KeyTypeHex {
		rawKey := serviceAccount.DefaultKey().ToConfig().Context[config.PrivateKeyField]
		privKey, err := crypto.DecodePrivateKeyHex(serviceAccount.DefaultKey().SigAlgo(), rawKey)
		if err != nil {
			return nil, err
		}

		opts = append(opts, emulator.WithServicePublicKey(
			privKey.PublicKey(),
			serviceAccount.DefaultKey().SigAlgo(),
			serviceAccount.DefaultKey().HashAlgo(),
		))

		// todo refactor to pass instance
		af := afero.Afero{
			Fs: afero.NewOsFs(),
		}

		exists, err := afero.DirExists(af, config.StateDir) // todo refactor
		if !exists {
			err := os.Mkdir(config.StateDir, os.FileMode(0755))
			if err != nil {
				return nil, err
			}
		}

		store, err := badger.New(badger.WithPath(
			path.Join(config.StateDir, config.MainState),
		))
		if err != nil {
			return nil, err
		}

		opts = append(opts, emulator.WithStore(store))
	}

	em, err := emulator.NewBlockchain(opts...)
	if err != nil {
		return nil, err
	}

	// todo does lorgus handle writer close
	f, err := os.OpenFile(
		path.Join(config.StateDir, config.MainState, config.StateLog),
		os.O_APPEND|os.O_CREATE|os.O_WRONLY,
		0644,
	)
	if err != nil {
		return nil, err
	}

	log := logrus.New()
	log.SetLevel(logrus.DebugLevel)
	log.SetOutput(f)

	b := backend.New(log, em)
	b.EnableAutoMine()

	return b, nil
}

func (g *EmulatorGateway) GetAccount(address flow.Address) (*flow.Account, error) {
	return g.backend.GetAccount(g.ctx, address)
}

func (g *EmulatorGateway) SendSignedTransaction(transaction *project.Transaction) (*flow.Transaction, error) {
	tx := transaction.FlowTransaction()

	err := g.backend.SendTransaction(g.ctx, *tx)
	if err != nil {
		return nil, err
	}

	return tx, err
}

func (g *EmulatorGateway) GetTransactionResult(tx *flow.Transaction, waitSeal bool) (*flow.TransactionResult, error) {
	result, err := g.backend.GetTransactionResult(g.ctx, tx.ID())
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
	return g.backend.GetTransaction(g.ctx, id)
}

func (g *EmulatorGateway) Ping() error {
	return nil
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
	events, err := g.backend.GetEventsForHeightRange(
		g.ctx,
		eventType,
		startHeight,
		endHeight,
	)

	var flowEvents []client.BlockEvents
	for _, h := range events {
		e, _ := sdk.FlowEventsToSDK(h.Events)

		flowEvents = append(flowEvents, client.BlockEvents{
			BlockID:        sdk.FlowIdentifierToSDK(h.BlockID),
			Height:         h.BlockHeight,
			BlockTimestamp: h.BlockTimestamp,
			Events:         e,
		})
	}

	return flowEvents, err
}

func (g *EmulatorGateway) GetCollection(id flow.Identifier) (*flow.Collection, error) {
	return g.backend.GetCollectionByID(g.ctx, id)
}

func (g *EmulatorGateway) GetBlockByID(id flow.Identifier) (*flow.Block, error) {
	block, err := g.backend.GetBlockByID(g.ctx, id)
	if err != nil {
		return nil, err
	}

	return convertBlock(block), err
}

func (g *EmulatorGateway) GetBlockByHeight(height uint64) (*flow.Block, error) {
	block, err := g.backend.GetBlockByHeight(g.ctx, height)
	return convertBlock(block), err
}
