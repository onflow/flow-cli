/*
 * Flow CLI
 *
 * Copyright 2019-2020 Dapper Labs, Inc.
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

package cli

import (
	"context"
	"fmt"
	"time"

	"github.com/onflow/cadence"
	"github.com/onflow/flow-go-sdk"
	"github.com/onflow/flow-go-sdk/client"
	"google.golang.org/grpc"
)

func GetTransactionResult(host string, id string, sealed bool, showTransactionCode bool) {
	ctx := context.Background()

	flowClient, err := client.New(host, grpc.WithInsecure(), grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(MaxGRPCMessageSize)))
	if err != nil {
		Exitf(1, "Failed to connect to host: %s", err)
	}

	txID := flow.HexToID(id)

	var res *flow.TransactionResult
	if sealed {
		res, err = waitForSeal(ctx, flowClient, txID)
	} else {
		res, err = flowClient.GetTransactionResult(ctx, txID)
	}
	if err != nil {
		Exitf(1, "Failed to get transaction result: %s", err)
	}

	tx, err := flowClient.GetTransaction(ctx, txID)
	if err != nil {
		Exitf(1, "Failed to get transaction: %s", err)
	}

	// Print out results of the TX to std out
	printTxResult(tx, res, showTransactionCode)
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

func printTxResult(tx *flow.Transaction, res *flow.TransactionResult, showCode bool) {
	fmt.Println()
	fmt.Println("Status: " + res.Status.String())
	if res.Error != nil {
		fmt.Println("Execution Error: " + res.Error.Error())
	}

	if showCode {
		fmt.Println("Code: ")
		fmt.Println(string(tx.Script))
	}

	fmt.Println("Events:")
	printEvents(res.Events, false)
	fmt.Println()
}

func GetBlockEvents(host string, height uint64, eventType string) {
	ctx := context.Background()

	flowClient, err := client.New(host, grpc.WithInsecure(), grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(MaxGRPCMessageSize)))
	if err != nil {
		Exitf(1, "Failed to connect to host: %s", err)
	}

	events, err := flowClient.GetEventsForHeightRange(ctx, client.EventRangeQuery{
		Type:        eventType,
		StartHeight: height,
		EndHeight:   height,
	})

	if err != nil {
		Exitf(1, "Failed to query block event by height: %s", err)
	}

	for _, blockEvent := range events {
		fmt.Printf("Events for Block %s:", blockEvent.BlockID)
		printEvents(blockEvent.Events, true)
	}
}

func printEvents(events []flow.Event, txID bool) {
	if len(events) == 0 {
		fmt.Println("  None")
	}
	// Basic event info printing
	for _, event := range events {
		fmt.Printf("  Event %d: %s\n", event.EventIndex, event.String())
		if txID {
			fmt.Printf("  Tx ID: %s\n", event.TransactionID)
		}
		fmt.Println("    Fields:")
		for i, field := range event.Value.EventType.Fields {
			value := event.Value.Fields[i]
			printField(field, value)
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

func printField(field cadence.Field, value cadence.Value) {
	v := value.ToGoValue()
	typeInfo := "Unknown"
	if field.Type != nil {
		typeInfo = field.Type.ID()
	} else if _, isAddress := v.([8]byte); isAddress {
		typeInfo = "Address"
	}
	fmt.Printf("      %s (%s): ", field.Identifier, typeInfo)
	// Try the two most obvious cases
	if address, ok := v.([8]byte); ok {
		fmt.Printf("%x", address)
	} else if isByteSlice(v) || field.Identifier == "publicKey" {
		// make exception for public key, since it get's interpreted as
		// []*big.Int
		for _, b := range v.([]interface{}) {
			fmt.Printf("%x", b)
		}
	} else if uintVal, ok := v.(uint64); typeInfo == "UFix64" && ok {
		fmt.Print(FormatUFix64(uintVal))
	} else {
		fmt.Printf("%v", v)
	}
	fmt.Println()
}
