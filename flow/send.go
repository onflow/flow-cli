package cli

import (
	"context"
	"fmt"
	"time"

	"github.com/onflow/flow-go-sdk"
	"github.com/onflow/flow-go-sdk/client"
	"github.com/onflow/flow-go-sdk/crypto"
	"google.golang.org/grpc"
)

func SendTransaction(host string, signerAccount *Account, script []byte) {
	ctx := context.Background()

	flowClient, err := client.New(host, grpc.WithInsecure())
	if err != nil {
		Exitf(1, "Failed to connect to host: %s", err)
	}

	signerAddress := signerAccount.Address

	fmt.Printf("Getting information for account with address 0x%s ...\n", signerAddress.Hex())

	account, err := flowClient.GetAccount(ctx, signerAddress)
	if err != nil {
		Exitf(1, "Failed to get account with address %s: 0x%s", signerAddress.Hex(), err)
	}

	signer := crypto.NewNaiveSigner(
		signerAccount.PrivateKey,
		signerAccount.HashAlgo,
	)

	// TODO: always use first?
	accountKey := account.Keys[0]

	sealed, err := flowClient.GetLatestBlockHeader(ctx, true)
	if err != nil {
		Exitf(1, "Failed to get latest sealed block: %s", err)
	}

	tx := flow.NewTransaction().
		SetReferenceBlockID(sealed.ID).
		SetScript(script).
		SetProposalKey(signerAddress, accountKey.ID, accountKey.SequenceNumber).
		SetPayer(signerAddress).
		AddAuthorizer(signerAddress)

	err = tx.SignEnvelope(signerAddress, accountKey.ID, signer)
	if err != nil {
		Exitf(1, "Failed to sign transaction: %s", err)
	}

	fmt.Printf("Submitting transaction with ID %s ...\n", tx.ID())

	err = flowClient.SendTransaction(context.Background(), *tx)
	if err == nil {
		fmt.Printf("Successfully submitted transaction with ID %s\n", tx.ID())
	} else {
		Exitf(1, "Failed to submit transaction: %s", err)
	}
	_, err = waitForSeal(ctx, flowClient, tx.ID())
	if err != nil {
		Exitf(1, "Failed to seal transaction: %s", err)
	}
}

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
	fmt.Println()
	fmt.Println("Status: " + res.Status.String())
	if res.Error != nil {
		fmt.Println("Execution Error: " + res.Error.Error())
		return
	}

	printEvents(res)
	fmt.Println()
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

func printEvents(res *flow.TransactionResult) {
	// Basic event info printing
	for _, event := range res.Events {
		fmt.Printf("Event %d: %s\n", event.EventIndex, event.String())
		fmt.Println("  Fields:")
		for i, field := range event.Value.EventType.Fields {
			fmt.Printf("    %s: ", field.Identifier)
			v := event.Value.Fields[i].ToGoValue()
			// Try the two most obvious cases
			if address, ok := v.([8]byte); ok {
				fmt.Printf("%x", address)
			} else if isByteSlice(v) {
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
