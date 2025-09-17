/*
 * Flow CLI
 *
 * Copyright Flow Foundation
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

package transactions

import (
	"context"
	"encoding/hex"
	"fmt"
	"math"
	"sort"
	"strconv"
	"strings"

	flowsdk "github.com/onflow/flow-go-sdk"
	"github.com/onflow/flow-go-sdk/crypto"
	"github.com/spf13/cobra"

	"github.com/onflow/flowkit/v2"
	"github.com/onflow/flowkit/v2/accounts"
	"github.com/onflow/flowkit/v2/config"
	"github.com/onflow/flowkit/v2/output"
	"github.com/onflow/flowkit/v2/transactions"

	"github.com/onflow/flow-cli/internal/command"
	"github.com/onflow/flow-cli/internal/prompt"
)

type flagsAddSig struct {
	Include  []string `default:"" flag:"include" info:"Fields to include in the output. Valid values: signatures, code, payload."`
	KeyIndex string   `default:"0" flag:"key-index" info:"Account key index"`
}

var addSigFlags = flagsAddSig{}

var addSigCommand = &command.Command{
	Cmd: &cobra.Command{
		Use:     "add-sig <built transaction filename> <address> <signature>",
		Short:   "Add signature to a prebuilt transaction",
		Example: "flow transactions add-sig ./built.rlp 99fa...25b",
		Args:    cobra.ExactArgs(3),
	},
	Flags: &addSigFlags,
	RunS:  addSig,
}

func addSig(
	args []string,
	globalFlags command.GlobalFlags,
	_ output.Logger,
	flow flowkit.Services,
	state *flowkit.State,
) (command.Result, error) {
	var payload []byte
	var err error
	filename := args[0]
	payload, err = state.ReadFile(filename)

	if err != nil {
		return nil, fmt.Errorf("failed to read partial transaction from %s: %v", filename, err)
	}

	address := flowsdk.HexToAddress(args[1])

	sig, err := hex.DecodeString(strings.ReplaceAll(args[2], "0x", ""))
	if err != nil {
		return nil, fmt.Errorf("invalid message signature: %w", err)
	}

	index, err := parseKeyIndex(addSigFlags.KeyIndex)
	if err != nil {
		return nil, err
	}

	var signed *transactions.Transaction
	var signers []*accounts.Account
	tx, err := transactions.NewFromPayload(payload)
	if err != nil {
		return nil, err
	}

	// validate all signers
	signerAccount := &accounts.Account{
		Address: address,
		Key:     &dummyKey{index: index, signer: crypto.NewAddSignatureSigner(sig, nil)},
	}

	//payer signs last
	sort.SliceStable(signers, func(i, j int) bool {
		return signers[i].Address.String() != tx.FlowTransaction().Payer.Hex()
	})

	if !globalFlags.Yes && !prompt.ApproveTransactionForSigningPrompt(tx.FlowTransaction()) {
		return nil, fmt.Errorf("transaction was not approved for signing")
	}

	signed, err = flow.SignTransactionPayload(context.Background(), signerAccount, payload)
	if err != nil {
		return nil, err
	}

	return &transactionResult{
		tx:      signed.FlowTransaction(),
		include: addSigFlags.Include,
		network: flow.Network().Name,
	}, nil
}

type dummyKey struct {
	keyType  config.KeyType
	index    uint32
	sigAlgo  crypto.SignatureAlgorithm
	hashAlgo crypto.HashAlgorithm
	signer   crypto.Signer
}

var _ accounts.Key = &dummyKey{}

func (a *dummyKey) Type() config.KeyType {
	return a.keyType
}

func (a *dummyKey) SigAlgo() crypto.SignatureAlgorithm {
	if a.sigAlgo == crypto.UnknownSignatureAlgorithm {
		return crypto.ECDSA_P256 // default value
	}
	return a.sigAlgo
}

func (a *dummyKey) HashAlgo() crypto.HashAlgorithm {
	if a.hashAlgo == crypto.UnknownHashAlgorithm {
		return crypto.SHA3_256 // default value
	}
	return a.hashAlgo
}

func (a *dummyKey) Index() uint32 {
	return a.index // default to 0
}

func (a *dummyKey) Validate() error {
	return nil
}
func (a *dummyKey) Signer(ctx context.Context) (crypto.Signer, error) {
	return a.signer, nil
}
func (a *dummyKey) ToConfig() config.AccountKey {
	return config.AccountKey{
		Type:           a.keyType,
		Index:          a.index,
		SigAlgo:        a.sigAlgo,
		HashAlgo:       a.hashAlgo,
		Mnemonic:       "",
		DerivationPath: "",
	}
}
func (a *dummyKey) PrivateKey() (*crypto.PrivateKey, error) {
	return nil, fmt.Errorf("This key type does not support private key retrieval")
}

func parseKeyIndex(value string) (uint32, error) {
	v, err := strconv.Atoi(value)
	if err != nil {
		return 0, fmt.Errorf("invalid index, must be a number")
	}
	if v < 0 {
		return 0, fmt.Errorf("invalid index, must be positive")
	}
	if v > math.MaxUint32 {
		return 0, fmt.Errorf("invalid index, must be less than %d", math.MaxUint32)
	}

	return uint32(v), nil
}
