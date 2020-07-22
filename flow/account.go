package cli

import (
	"context"

	"github.com/onflow/flow-go-sdk"
	"github.com/onflow/flow-go-sdk/client"
	"google.golang.org/grpc"
)

func GetAccount(host string, address flow.Address) *flow.Account {
	ctx := context.Background()

	flowClient, err := client.New(host, grpc.WithInsecure())
	if err != nil {
		Exitf(1, "Failed to connect to host: %s", err)
	}

	account, err := flowClient.GetAccount(ctx, address)
	if err != nil {
		Exitf(1, "Failed to get account with address %s: %s", address, err)
	}
	return account
}
