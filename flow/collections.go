package cli

import (
	"context"

	"github.com/onflow/flow-go-sdk"
	"github.com/onflow/flow-go-sdk/client"
	"google.golang.org/grpc"
)

func GetCollectionByID(host string, collectionID flow.Identifier) *flow.Collection {
	ctx := context.Background()

	flowClient, err := client.New(host, grpc.WithInsecure())
	if err != nil {
		Exitf(1, "Failed to connect to host: %s", err)
	}

	collection, err := flowClient.GetCollection(ctx, collectionID)
	if err != nil {
		Exitf(1, "Failed to retrieve collection by ID %s: %s", collectionID, err)
	}
	return collection
}
