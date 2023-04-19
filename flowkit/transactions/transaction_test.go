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

package transactions_test

import (
	"testing"

	"github.com/onflow/cadence"
	jsoncdc "github.com/onflow/cadence/encoding/json"
	"github.com/onflow/flow-go-sdk"
	"github.com/onflow/flow-go-sdk/crypto"
	"github.com/stretchr/testify/assert"

	"github.com/onflow/flow-cli/flowkit/accounts"
	"github.com/onflow/flow-cli/flowkit/tests"
	"github.com/onflow/flow-cli/flowkit/transactions"
)

func TestNew(t *testing.T) {
	tx := transactions.New()

	assert.NotNil(t, tx)
	assert.NotNil(t, tx.FlowTransaction())

	computeLimit := uint64(1000)
	tx.SetComputeLimit(computeLimit)
	assert.Equal(t, computeLimit, tx.FlowTransaction().GasLimit)

	arg := cadence.NewInt(1)
	_ = tx.AddArgument(arg)
	enc, _ := jsoncdc.Encode(arg)
	assert.Equal(t, string(enc), string(tx.FlowTransaction().Arguments[0]))

	payer := flow.HexToAddress("0x01")
	tx.SetPayer(payer)
	assert.Equal(t, payer, tx.FlowTransaction().Payer)

	script := []byte(`
		transaction (arg: Int) {
			prepare(auth: AuthAccount) {}
			execute {}
		}
	`)
	err := tx.SetScriptWithArgs(script, []cadence.Value{arg})
	assert.NoError(t, err)
	assert.Equal(t, script, tx.FlowTransaction().Script)
	assert.Equal(t, string(enc), string(tx.FlowTransaction().Arguments[0]))

	auths := []flow.Address{flow.HexToAddress("0x02")}
	tx, err = tx.AddAuthorizers(auths)
	assert.NoError(t, err)
	assert.Equal(t, auths, tx.FlowTransaction().Authorizers)

	addr := flow.HexToAddress("0x03")
	index := 1
	proposer := tests.NewAccountWithAddress(addr.String())
	err = tx.SetProposer(proposer, index)
	assert.NoError(t, err)
	assert.Equal(t, addr.String(), tx.FlowTransaction().ProposalKey.Address.String())
	assert.Equal(t, proposer.Keys[index].Index, tx.FlowTransaction().ProposalKey.KeyIndex)

	sig, _ := accounts.NewEmulatorAccount(crypto.ECDSA_P256, crypto.SHA3_256)
	sig.Address = flow.HexToAddress("0x01")
	err = tx.SetSigner(sig)
	assert.NoError(t, err)
	assert.Equal(t, sig, tx.Signer())

	signed, err := tx.Sign()
	assert.NoError(t, err)
	assert.Len(t, signed.FlowTransaction().EnvelopeSignatures, 1)
}
