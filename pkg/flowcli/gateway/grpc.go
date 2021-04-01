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
	"fmt"
	"time"

	"github.com/onflow/cadence"
	"github.com/onflow/flow-cli/pkg/flowcli/project"
	"github.com/onflow/flow-go-sdk"
	"github.com/onflow/flow-go-sdk/client"
	"google.golang.org/grpc"
)

const (
	defaultGasLimit = 1000
)

// GrpcGateway contains all functions that need flow client to execute
type GrpcGateway struct {
	client *client.Client
	ctx    context.Context
}

// NewGrpcGateway creates new grpc gateway
func NewGrpcGateway(host string) (*GrpcGateway, error) {
	gClient, err := client.New(host, grpc.WithInsecure())
	ctx := context.Background()

	if err != nil || gClient == nil {
		return nil, fmt.Errorf("failed to connect to host %s", host)
	}

	return &GrpcGateway{
		client: gClient,
		ctx:    ctx,
	}, nil
}

// GetAccount gets account by the address from flow
func (g *GrpcGateway) GetAccount(address flow.Address) (*flow.Account, error) {
	account, err := g.client.GetAccountAtLatestBlock(g.ctx, address)
	if err != nil {
		return nil, fmt.Errorf("Failed to get account with address %s: %s", address, err)
	}

	return account, nil
}

// PrepareTransactionPayload prepares the payload for the transaction from the network
func (g *GrpcGateway) PrepareTransactionPayload(tx *project.Transaction) (*project.Transaction, error) {
	signerAddress := tx.Signer().Address()
	account, err := g.GetAccount(signerAddress)
	if err != nil {
		return nil, err
	}

	accountKey := account.Keys[tx.Signer().DefaultKey().Index()]

	sealed, err := g.client.GetLatestBlockHeader(g.ctx, true)
	if err != nil {
		return nil, fmt.Errorf("Failed to get latest sealed block: %s", err)
	}

	tx.FlowTransaction().
		SetReferenceBlockID(sealed.ID).
		SetGasLimit(defaultGasLimit).
		SetProposalKey(signerAddress, accountKey.Index, accountKey.SequenceNumber)

	return tx, nil
}

// SendTransaction prepares, signs and sends the transaction to the network
func (g *GrpcGateway) SendTransaction(transaction *project.Transaction) (*flow.Transaction, error) {
	tx, err := g.PrepareTransactionPayload(transaction)
	if err != nil {
		return nil, err
	}

	tx, err = tx.Sign()
	if err != nil {
		return nil, err
	}

	return g.SendSignedTransaction(tx)
}

// SendSignedTransaction sends a transaction to flow that is already prepared and signed
func (g *GrpcGateway) SendSignedTransaction(transaction *project.Transaction) (*flow.Transaction, error) {
	tx := transaction.FlowTransaction()

	err := g.client.SendTransaction(g.ctx, *tx)
	if err != nil {
		return nil, fmt.Errorf("Failed to submit transaction: %s", err)
	}

	return tx, nil
}

// GetTransaction gets transaction by id
func (g *GrpcGateway) GetTransaction(id flow.Identifier) (*flow.Transaction, error) {
	return g.client.GetTransaction(g.ctx, id)
}

// GetTransactionResult gets result of a transaction on flow
func (g *GrpcGateway) GetTransactionResult(tx *flow.Transaction, waitSeal bool) (*flow.TransactionResult, error) {
	result, err := g.client.GetTransactionResult(g.ctx, tx.ID())
	if err != nil {
		return nil, err
	}

	if result.Status != flow.TransactionStatusSealed && waitSeal {
		time.Sleep(time.Second)
		return g.GetTransactionResult(tx, waitSeal)
	}

	return result, nil
}

// ExecuteScript execute scripts on flow
func (g *GrpcGateway) ExecuteScript(script []byte, arguments []cadence.Value) (cadence.Value, error) {

	value, err := g.client.ExecuteScriptAtLatestBlock(g.ctx, script, arguments)
	if err != nil {
		return nil, fmt.Errorf("Failed to submit executable script %s", err)
	}

	return value, nil
}

// GetLatestBlock gets latest block from flow
func (g *GrpcGateway) GetLatestBlock() (*flow.Block, error) {
	return g.client.GetLatestBlock(g.ctx, true)
}

// GetBlockByID get block by id from flow
func (g *GrpcGateway) GetBlockByID(id flow.Identifier) (*flow.Block, error) {
	return g.client.GetBlockByID(g.ctx, id)
}

// GetBlockByHeight get block by id from flow
func (g *GrpcGateway) GetBlockByHeight(height uint64) (*flow.Block, error) {
	return g.client.GetBlockByHeight(g.ctx, height)
}

// GetEvents gets event from start and end height
func (g *GrpcGateway) GetEvents(
	eventType string,
	startHeight uint64,
	endHeight uint64,
) ([]client.BlockEvents, error) {

	events, err := g.client.GetEventsForHeightRange(
		g.ctx,
		client.EventRangeQuery{
			Type:        eventType,
			StartHeight: startHeight,
			EndHeight:   endHeight,
		},
	)

	return events, err
}

// GetCollection get collection by id from flow
func (g *GrpcGateway) GetCollection(id flow.Identifier) (*flow.Collection, error) {
	return g.client.GetCollection(g.ctx, id)
}
