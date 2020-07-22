package cli

import (
	"context"

	"github.com/onflow/flow-go-sdk"
	"github.com/onflow/flow-go-sdk/client"
	"google.golang.org/grpc"
)

func GetBlockByID(host string, blockID flow.Identifier) *flow.Block {
	ctx := context.Background()

	flowClient, err := client.New(host, grpc.WithInsecure())
	if err != nil {
		Exitf(1, "Failed to connect to host: %s", err)
	}

	block, err := flowClient.GetBlockByID(ctx, blockID)
	if err != nil {
		Exitf(1, "Failed to retrieve block by ID %s: %s", blockID, err)
	}
	return block
}

func GetBlockByHeight(host string, height uint64) *flow.Block {
	ctx := context.Background()

	flowClient, err := client.New(host, grpc.WithInsecure())
	if err != nil {
		Exitf(1, "Failed to connect to host: %s", err)
	}

	block, err := flowClient.GetBlockByHeight(ctx, height)
	if err != nil {
		Exitf(1, "Failed to retrieve block by height %d: %s", height, err)
	}
	return block
}

func GetLatestBlock(host string) *flow.Block {
	ctx := context.Background()

	flowClient, err := client.New(host, grpc.WithInsecure())
	if err != nil {
		Exitf(1, "Failed to connect to host: %s", err)
	}

	block, err := flowClient.GetLatestBlock(ctx, true)
	if err != nil {
		Exitf(1, "Failed to retrieve latest block: %s", err)
	}
	return block
}
