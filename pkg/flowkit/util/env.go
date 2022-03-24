/*
 * Flow CLI
 *
 * Copyright 2019 Dapper Labs, Inc.
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

package util

import (
	"github.com/onflow/flow-core-contracts/lib/go/templates"
	"github.com/onflow/flow-go-sdk"
)

func EnvFromNetwork(network flow.ChainID) templates.Environment {
	if network == flow.Mainnet {
		return templates.Environment{
			IDTableAddress:       "8624b52f9ddcd04a",
			FungibleTokenAddress: "f233dcee88fe0abe",
			FlowTokenAddress:     "1654653399040a61",
			LockedTokensAddress:  "8d0e87b65159ae63",
			StakingProxyAddress:  "62430cf28c26d095",
		}
	}

	if network == flow.Testnet {
		return templates.Environment{
			IDTableAddress:       "9eca2b38b18b5dfe",
			FungibleTokenAddress: "9a0766d93b6608b7",
			FlowTokenAddress:     "7e60df042a9c0868",
			LockedTokensAddress:  "95e019a17d0e23d7",
			StakingProxyAddress:  "7aad92e5a0715d21",
		}
	}

	return templates.Environment{}
}
