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

package gateway

import (
	"github.com/onflow/cadence"
	"github.com/onflow/flow-go-sdk"
)

//go:generate  mockery --name=Gateway

// Gateway describes blockchain access interface
type Gateway interface {
	GetAccount(flow.Address) (*flow.Account, error)
	SendSignedTransaction(*flow.Transaction) (*flow.Transaction, error)
	GetTransaction(flow.Identifier) (*flow.Transaction, error)
	GetTransactionResultsByBlockID(blockID flow.Identifier) ([]*flow.TransactionResult, error)
	GetTransactionResult(flow.Identifier, bool) (*flow.TransactionResult, error)
	GetTransactionsByBlockID(blockID flow.Identifier) ([]*flow.Transaction, error)
	ExecuteScript([]byte, []cadence.Value, *util.ScriptQuery) (cadence.Value, error)
	GetLatestBlock() (*flow.Block, error)
	GetBlockByHeight(uint64) (*flow.Block, error)
	GetBlockByID(flow.Identifier) (*flow.Block, error)
	GetEvents(string, uint64, uint64) ([]flow.BlockEvents, error)
	GetCollection(flow.Identifier) (*flow.Collection, error)
	GetLatestProtocolStateSnapshot() ([]byte, error)
	Ping() error
	SecureConnection() bool
}
