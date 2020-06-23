package cli

import (
	"context"
	"fmt"
	"time"

	"github.com/onflow/flow-go-sdk"
	"github.com/onflow/flow-go-sdk/client"
	"google.golang.org/grpc"
)

func GetTransactionResult(host string, id string) {
	ctx := context.Background()

	flowClient, err := client.New(host, grpc.WithInsecure())
	if err != nil {
		Exitf(1, "Failed to connect to host: %s", err)
	}

	txID := flow.HexToID(id)

	res, err := waitForSeal(ctx, flowClient, txID)
	if err != nil {
		Exitf(1, "Failed to get transaction result: %s", err)
	}

	// Print out results of the TX to std out
	printResult(res)
}

func waitForSeal(ctx context.Context, c *client.Client, id flow.Identifier) (*flow.TransactionResult, error) {
	result, err := c.GetTransactionResult(ctx, id)
	if err != nil {
		return nil, err
	}

	fmt.Printf("Waiting for transaction %s to be sealed...\n", id)
	for result.Status != flow.TransactionStatusSealed {
		time.Sleep(time.Second)
		fmt.Print(".")
		result, err = c.GetTransactionResult(ctx, id)
		if err != nil {
			return nil, err
		}
	}

	fmt.Println()
	fmt.Printf("Transaction %s sealed\n", id)

	return result, nil
}

func printResult(res *flow.TransactionResult) {
	fmt.Println()
	fmt.Println("Status: " + res.Status.String())
	if res.Error != nil {
		fmt.Println("Execution Error: " + res.Error.Error())
		return
	}

	printEvents(res.Events)
	fmt.Println()
}

func printEvents(events []flow.Event) {
	// Basic event info printing
	for _, event := range events {
		fmt.Printf("Event %d: %s\n", event.EventIndex, event.String())
		fmt.Println("  Fields:")
		for i, field := range event.Value.EventType.Fields {
			fmt.Printf("    %s: ", field.Identifier)
			v := event.Value.Fields[i].ToGoValue()
			// Try the two most obvious cases
			if address, ok := v.([8]byte); ok {
				fmt.Printf("%x", address)
			} else if isByteSlice(v) || field.Identifier == "publicKey" {
				// make exception for public key, since it get's interpreted as
				// []*big.Int
				for _, b := range v.([]interface{}) {
					fmt.Printf("%x", b)
				}
			} else {
				fmt.Printf("%v", v)
			}
			fmt.Println()
		}
	}
}

func isByteSlice(v interface{}) bool {
	slice, isSlice := v.([]interface{})
	if !isSlice {
		return false
	}
	_, isBytes := slice[0].(byte)
	return isBytes
}
