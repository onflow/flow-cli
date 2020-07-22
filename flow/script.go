package cli

import (
	"context"
	"fmt"

	jsoncdc "github.com/onflow/cadence/encoding/json"
	"github.com/onflow/flow-go-sdk/client"
	"google.golang.org/grpc"
)

func ExecuteScript(host string, script []byte) {
	ctx := context.Background()

	flowClient, err := client.New(host, grpc.WithInsecure())
	if err != nil {
		Exitf(1, "Failed to connect to host: %s", err)
	}
	value, err := flowClient.ExecuteScriptAtLatestBlock(ctx, script, nil)
	if err != nil {
		Exitf(1, "Failed to submit executable script: %s", err)
	}
	b, err := jsoncdc.Encode(value)
	if err != nil {
		Exitf(1, "Failed to decode cadence value: %s", err)
	}
	fmt.Println(string(b))
}
