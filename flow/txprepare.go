/*
 * Flow CLI
 *
 * Copyright 2019-2021 Dapper Labs, Inc.
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

func PrepareTransaction(host string, proposerAccount *Account, tx *flow.Transaction, payer flow.Address) *flow.Transaction {
	ctx := context.Background()

	flowClient, err := client.New(host, grpc.WithInsecure())
	if err != nil {
		Exitf(1, "Failed to connect to host: %s", err)
	}

	proposerAddress := proposerAccount.Address

	fmt.Printf("Getting information for account with address 0x%s ...\n", proposerAddress.Hex())

	account, err := flowClient.GetAccount(ctx, proposerAddress)
	if err != nil {
		Exitf(1, "Failed to get account with address %s: 0x%s", proposerAddress.Hex(), err)
	}

	// Default 0, i.e. first key
	accountKey := account.Keys[proposerAccount.KeyIndex]

	sealed, err := flowClient.GetLatestBlockHeader(ctx, true)
	if err != nil {
		Exitf(1, "Failed to get latest sealed block: %s", err)
	}

	tx.SetReferenceBlockID(sealed.ID).
		SetProposalKey(proposerAddress, accountKey.Index, accountKey.SequenceNumber).
		SetPayer(payer)

	return tx
}
