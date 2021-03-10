package gateway

import (
	"context"
	"fmt"
	"time"

	"github.com/onflow/cadence"
	"github.com/onflow/flow-cli/flow/cli"
	"github.com/onflow/flow-go-sdk"
	"github.com/onflow/flow-go-sdk/client"
	"google.golang.org/grpc"
)

// GrpcGateway contains all functions that need flow client to execute
type GrpcGateway struct {
	client *client.Client
	ctx    context.Context
}

// NewGrpcGateway creates new grpc gateway
func NewGrpcGateway(host string) (*GrpcGateway, error) {
	client, err := client.New(host, grpc.WithInsecure())
	ctx := context.Background()

	if err != nil || client == nil {
		return nil, fmt.Errorf("failed to connect to host %s", host)
	}

	return &GrpcGateway{
		client: client,
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

// TODO: replace with txsender - much nicer implemented
// SendTransaction send a transaction to flow
func (g *GrpcGateway) SendTransaction(tx *flow.Transaction, signer *cli.Account) (*flow.Transaction, error) {
	//fmt.Printf("Getting information for account with address 0x%s ...\n", signer.Address()) TODO: change to log

	account, err := g.GetAccount(signer.Address())
	if err != nil {
		return nil, fmt.Errorf("Failed to get account with address %s: 0x%s", signer.Address(), err)
	}

	// Default 0, i.e. first key
	accountKey := account.Keys[0]

	sealed, err := g.client.GetLatestBlockHeader(g.ctx, true)
	if err != nil {
		return nil, fmt.Errorf("Failed to get latest sealed block: %s", err)
	}

	tx.SetReferenceBlockID(sealed.ID).
		SetProposalKey(signer.Address(), accountKey.Index, accountKey.SequenceNumber).
		SetPayer(signer.Address())

	err = tx.SignEnvelope(signer.Address(), accountKey.Index, signer.DefaultKey().Signer())
	if err != nil {
		return nil, fmt.Errorf("Failed to sign transaction: %s", err)
	}

	//fmt.Printf("Submitting transaction with ID %s ...\n", tx.ID())

	err = g.client.SendTransaction(g.ctx, *tx)
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
func (g *GrpcGateway) GetTransactionResult(tx *flow.Transaction) (*flow.TransactionResult, error) {
	result, err := g.client.GetTransactionResult(g.ctx, tx.ID())
	if err != nil {
		return nil, err
	}

	//fmt.Printf("Waiting for transaction %s to be sealed...\n", result.Status) // todo: change to log
	if result.Status != flow.TransactionStatusSealed {
		time.Sleep(time.Second)
		//fmt.Print(".") // todo: change to spinner loader
		return g.GetTransactionResult(tx)
	}

	//fmt.Printf("Transaction %s sealed \n", id) todo: change to log
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
