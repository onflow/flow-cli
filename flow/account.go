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

func GetAddressNetwork(address flow.Address) (flow.ChainID, error) {
	networks := []flow.ChainID{
		flow.Mainnet,
		flow.Testnet,
		flow.Emulator,
	}
	for _, net := range networks {
		if address.IsValid(net) {
			return net, nil
		}
	}
	return flow.ChainID(""), fmt.Errorf("Unrecognized address not valid for any known chain: %s", address)
}
