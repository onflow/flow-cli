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

	"github.com/onflow/flow-go-sdk"
	"github.com/onflow/flow-go-sdk/client"
	"google.golang.org/grpc"
)

type SignerRole string

const (
	SignerRoleAuthorizer SignerRole = "authorizer"
	SignerRoleProposer   SignerRole = "proposer"
	SignerRolePayer      SignerRole = "payer"
)

func PrepareTransaction(host string, signerAccount *Account, tx *flow.Transaction, payer flow.Address) *flow.Transaction {
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

	// Default 0, i.e. first key
	accountKey := account.Keys[signerAccount.KeyIndex]

	sealed, err := flowClient.GetLatestBlockHeader(ctx, true)
	if err != nil {
		Exitf(1, "Failed to get latest sealed block: %s", err)
	}

	tx.SetReferenceBlockID(sealed.ID).
		SetProposalKey(signerAddress, accountKey.Index, accountKey.SequenceNumber).
		SetPayer(payer)

	return tx
}

func SendTransaction(host string, signerAccount *Account, tx *flow.Transaction, withResults bool) {
	ctx := context.Background()

	flowClient, err := client.New(host, grpc.WithInsecure())
	if err != nil {
		Exitf(1, "Failed to connect to host: %s", err)
	}

	tx = signTransaction(ctx, flowClient, signerAccount, SignerRolePayer, tx)

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
		printTxResult(tx, res, true)
	}
}

func PrepareAndSendTransaction(host string, signerAccount *Account, tx *flow.Transaction, payer flow.Address, withResults bool) {
	preparedTx := PrepareTransaction(host, signerAccount, tx, payer)
	SendTransaction(host, signerAccount, preparedTx, withResults)
}

func SignTransaction(host string, signerAccount *Account, signerRole SignerRole, tx *flow.Transaction) *flow.Transaction {
	ctx := context.Background()

	flowClient, err := client.New(host, grpc.WithInsecure())
	if err != nil {
		Exitf(1, "Failed to connect to host: %s", err)
	}

	tx = signTransaction(ctx, flowClient, signerAccount, signerRole, tx)
	return tx
}

func signTransaction(
	ctx context.Context,
	flowClient *client.Client,
	signerAccount *Account,
	signerRole SignerRole,
	tx *flow.Transaction,
) *flow.Transaction {
	signerAddress := signerAccount.Address
	account, err := flowClient.GetAccount(ctx, signerAddress)
	if err != nil {
		Exitf(1, "Failed to get account with address %s: %s", signerAddress.Hex(), err)
	}
	accountKey := account.Keys[signerAccount.KeyIndex]
	switch signerRole {
	case SignerRoleAuthorizer:
		err := tx.SignPayload(signerAddress, accountKey.Index, signerAccount.Signer)
		if err != nil {
			Exitf(1, "Failed to sign transaction: %s", err)
		}
	case SignerRolePayer:
		err := tx.SignEnvelope(signerAddress, accountKey.Index, signerAccount.Signer)
		if err != nil {
			Exitf(1, "Failed to sign transaction: %s", err)
		}
	default:
		Exitf(1, "Failed to sign transaction: unknown signer role %s", signerRole)
	}

	return tx
}
