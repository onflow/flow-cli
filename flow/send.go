package cli

import (
	"context"
	"fmt"

	"github.com/onflow/flow-go-sdk"
	"github.com/onflow/flow-go-sdk/client"
	"github.com/onflow/flow-go-sdk/crypto"
	"google.golang.org/grpc"
)

func SendTransaction(host string, signerAccount *Account, tx *flow.Transaction, withResults bool) {
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

	tx.SetReferenceBlockID(sealed.ID).
		SetProposalKey(signerAddress, accountKey.Index, accountKey.SequenceNumber).
		SetPayer(signerAddress)

	err = tx.SignEnvelope(signerAddress, accountKey.Index, signer)
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
	if withResults {
		res, err := waitForSeal(ctx, flowClient, tx.ID())
		if err != nil {
			Exitf(1, "Failed to seal transaction: %s", err)
		}
		printTxResult(res)
	}
}
