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
	"strings"
	"time"

	"github.com/onflow/cadence"
	"github.com/onflow/flow-go-sdk"
	grpcAccess "github.com/onflow/flow-go-sdk/access/grpc"
	"github.com/onflow/flow-go/utils/grpcutils"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/onflow/flow-cli/flowkit/v2/config"
)

// maxGRPCMessageSize 60mb
const maxGRPCMessageSize = 1024 * 1024 * 60

// GrpcGateway is a gateway implementation that uses the Flow Access gRPC API.
type GrpcGateway struct {
	client       *grpcAccess.Client
	ctx          context.Context
	secureClient bool
}

// NewGrpcGateway returns a new gRPC gateway.
func NewGrpcGateway(network config.Network) (*GrpcGateway, error) {

	gClient, err := grpcAccess.NewClient(
		network.Host,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(maxGRPCMessageSize)),
	)
	ctx := context.Background()

	if err != nil || gClient == nil {
		return nil, fmt.Errorf("failed to connect to host %s", network.Host)
	}

	return &GrpcGateway{
		client:       gClient,
		ctx:          ctx,
		secureClient: false,
	}, nil
}

// NewSecureGrpcGateway returns a new gRPC gateway with a secure client connection.
func NewSecureGrpcGateway(network config.Network) (*GrpcGateway, error) {
	secureDialOpts, err := grpcutils.SecureGRPCDialOpt(strings.TrimPrefix(network.Key, "0x"))
	if err != nil {
		return nil, fmt.Errorf("failed to create secure GRPC dial options with network key \"%s\": %w", network.Key, err)
	}

	gClient, err := grpcAccess.NewClient(
		network.Host,
		secureDialOpts,
		grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(maxGRPCMessageSize)),
	)
	ctx := context.Background()

	if err != nil || gClient == nil {
		return nil, fmt.Errorf("failed to connect to host %s", network.Host)
	}

	return &GrpcGateway{
		client:       gClient,
		ctx:          ctx,
		secureClient: true,
	}, nil
}

// GetAccount gets an account by address from the Flow Access API.
func (g *GrpcGateway) GetAccount(address flow.Address) (*flow.Account, error) {
	account, err := g.client.GetAccountAtLatestBlock(g.ctx, address)
	if err != nil {
		return nil, fmt.Errorf("failed to get account with address %s: %w", address, err)
	}

	return account, nil
}

// SendSignedTransaction sends a transaction to flow that is already prepared and signed.
func (g *GrpcGateway) SendSignedTransaction(tx *flow.Transaction) (*flow.Transaction, error) {
	err := g.client.SendTransaction(g.ctx, *tx)
	if err != nil {
		return nil, fmt.Errorf("failed to submit transaction: %w", err)
	}

	return tx, nil
}

// GetTransaction gets a transaction by ID from the Flow Access API.
func (g *GrpcGateway) GetTransaction(ID flow.Identifier) (*flow.Transaction, error) {
	return g.client.GetTransaction(g.ctx, ID)
}

func (g *GrpcGateway) GetTransactionResultsByBlockID(blockID flow.Identifier) ([]*flow.TransactionResult, error) {
	return g.client.GetTransactionResultsByBlockID(g.ctx, blockID)
}

func (g *GrpcGateway) GetTransactionsByBlockID(blockID flow.Identifier) ([]*flow.Transaction, error) {
	return g.client.GetTransactionsByBlockID(g.ctx, blockID)
}

// GetTransactionResult gets a transaction result by ID from the Flow Access API.
func (g *GrpcGateway) GetTransactionResult(ID flow.Identifier, waitSeal bool) (*flow.TransactionResult, error) {
	result, err := g.client.GetTransactionResult(g.ctx, ID)
	if err != nil {
		return nil, err
	}

	if result.Status != flow.TransactionStatusSealed && waitSeal {
		time.Sleep(time.Second)
		return g.GetTransactionResult(ID, waitSeal)
	}

	return result, nil
}

// ExecuteScript executes a script on Flow through the Access API.
func (g *GrpcGateway) ExecuteScript(script []byte, arguments []cadence.Value) (cadence.Value, error) {
	return g.client.ExecuteScriptAtLatestBlock(g.ctx, script, arguments)
}

// ExecuteScriptAtHeight executes a script at block height.
func (g *GrpcGateway) ExecuteScriptAtHeight(script []byte, arguments []cadence.Value, height uint64) (cadence.Value, error) {
	return g.client.ExecuteScriptAtBlockHeight(g.ctx, height, script, arguments)
}

// ExecuteScriptAtID executes a script at block ID.
func (g *GrpcGateway) ExecuteScriptAtID(script []byte, arguments []cadence.Value, ID flow.Identifier) (cadence.Value, error) {
	return g.client.ExecuteScriptAtBlockID(g.ctx, ID, script, arguments)
}

// GetLatestBlock gets the latest block on Flow through the Access API.
func (g *GrpcGateway) GetLatestBlock() (*flow.Block, error) {
	return g.client.GetLatestBlock(g.ctx, true)
}

// GetBlockByID get block by ID from the Flow Access API.
func (g *GrpcGateway) GetBlockByID(id flow.Identifier) (*flow.Block, error) {
	return g.client.GetBlockByID(g.ctx, id)
}

// GetBlockByHeight get block by height from the Flow Access API.
func (g *GrpcGateway) GetBlockByHeight(height uint64) (*flow.Block, error) {
	return g.client.GetBlockByHeight(g.ctx, height)
}

// GetEvents gets events by name and block range from the Flow Access API.
func (g *GrpcGateway) GetEvents(
	eventType string,
	startHeight uint64,
	endHeight uint64,
) ([]flow.BlockEvents, error) {

	events, err := g.client.GetEventsForHeightRange(
		g.ctx,
		eventType,
		startHeight,
		endHeight,
	)

	return events, err
}

// GetCollection gets a collection by ID from the Flow Access API.
func (g *GrpcGateway) GetCollection(id flow.Identifier) (*flow.Collection, error) {
	return g.client.GetCollection(g.ctx, id)
}

// GetLatestProtocolStateSnapshot gets the latest finalized protocol state snapshot
func (g *GrpcGateway) GetLatestProtocolStateSnapshot() ([]byte, error) {
	return g.client.GetLatestProtocolStateSnapshot(g.ctx)
}

// Ping is used to check if the access node is alive and healthy.
func (g *GrpcGateway) Ping() error {
	return g.client.Ping(g.ctx)
}

// SecureConnection is used to log warning if a service should be using a secure client but is not
func (g *GrpcGateway) SecureConnection() bool {
	return g.secureClient
}
