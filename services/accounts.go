package services

import (
	"context"
	"fmt"
	"github.com/onflow/flow-go-sdk"
	"github.com/onflow/flow-go-sdk/client"
	"google.golang.org/grpc"
	"strings"
)

type Services struct {
}

func GetAccount(host string, address string) (*flow.Account, error) {
	ctx := context.Background()

	flowAddress := flow.HexToAddress(
		strings.ReplaceAll(address, "0x", ""),
	)

	flowClient, err := client.New(host, grpc.WithInsecure())
	if err != nil {
		return nil, fmt.Errorf("failed to connect to host %s", host)
	}

	account, err := flowClient.GetAccount(ctx, flowAddress)
	if err != nil {
		return nil, fmt.Errorf("Failed to get account with address %s: %s", address, err)
	}

	return account, nil
}
