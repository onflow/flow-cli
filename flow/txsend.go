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
