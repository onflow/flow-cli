package gateway

import (
	"context"
	"fmt"
	"github.com/onflow/flow-cli/flow/cli"
	"github.com/onflow/flow-go-sdk"
	"github.com/onflow/flow-go-sdk/client"
	"google.golang.org/grpc"
	"time"
)

type GrpcGateway struct {
	client *client.Client
	ctx    context.Context
}

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

func (g *GrpcGateway) GetAccount(address flow.Address) (*flow.Account, error) {
	account, err := g.client.GetAccount(g.ctx, address)
	if err != nil {
		return nil, fmt.Errorf("Failed to get account with address %s: %s", address, err)
	}

	return account, nil
}

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

	//fmt.Printf("Transaction %s sealed\n", id) todo: change to log
	return result, nil
}

func (g *GrpcGateway) GetEvents() {}
