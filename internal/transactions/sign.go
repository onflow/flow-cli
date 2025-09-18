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
	"bytes"
	"context"
	"encoding/hex"
	"fmt"
	"io"
	"math"
	"net/http"
	"sort"
	"strconv"
	"strings"

	"github.com/onflow/flow-cli/internal/prompt"
	"github.com/onflow/flow-go-sdk/crypto"

	"github.com/onflow/flowkit/v2/config"
	"github.com/onflow/flowkit/v2/transactions"

	"github.com/onflow/flowkit/v2/accounts"

	flowsdk "github.com/onflow/flow-go-sdk"
	"github.com/spf13/cobra"

	"github.com/onflow/flowkit/v2"
	"github.com/onflow/flowkit/v2/output"

	"github.com/onflow/flow-cli/internal/command"
)

type flagsSign struct {
	Signer        []string `default:"emulator-account" flag:"signer" info:"name of a single or multiple comma-separated accounts used to sign"`
	Include       []string `default:"" flag:"include" info:"Fields to include in the output. Valid values: signatures, code, payload."`
	FromRemoteUrl string   `default:"" flag:"from-remote-url" info:"server URL where RLP can be fetched, signed RLP will be posted back to remote URL."`
	RawSig        []string `default:"" flag:"raw-sig" info:"Raw hex-encoded signature to add to the transaction, instead of signing with a private key from an account in the config."`
}

var signFlags = flagsSign{}

var signCommand = &command.Command{
	Cmd: &cobra.Command{
		Use:     "sign [<built transaction filename> | --from-remote-url <url>]",
		Short:   "Sign built transaction",
		Example: "flow transactions sign ./built.rlp --signer alice",
		Args:    cobra.MaximumNArgs(1),
	},
	Flags: &signFlags,
	RunS:  sign,
}

func sign(
	args []string,
	globalFlags command.GlobalFlags,
	_ output.Logger,
	flow flowkit.Services,
	state *flowkit.State,
) (command.Result, error) {
	var payload []byte
	var err error
	var filenameOrUrl string

	if signFlags.FromRemoteUrl != "" && len(args) > 0 {
		return nil, fmt.Errorf("only use one, filename argument or --from-remote-url <url>")
	}

	if signFlags.FromRemoteUrl != "" {
		if globalFlags.Yes {
			return nil, fmt.Errorf("--yes is not supported with this flag")
		}
		filenameOrUrl = signFlags.FromRemoteUrl
		payload, err = getRLPTransaction(filenameOrUrl)
	} else {
		if len(args) == 0 {
			return nil, fmt.Errorf("filename argument is required")
		}
		filenameOrUrl = args[0]
		payload, err = state.ReadFile(filenameOrUrl)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to read partial transaction from %s: %v", filenameOrUrl, err)
	}

	var signed *transactions.Transaction
	var signers []*accounts.Account
	tx, err := transactions.NewFromPayload(payload)
	if err != nil {
		return nil, err
	}

	// Use raw signatures if provided, and if the number of signatures matches the number of signers
	useRawSig := len(signFlags.RawSig) == len(signFlags.Signer)

	// validate all signers
	for i, signerName := range signFlags.Signer {
		signer, err := state.Accounts().ByName(signerName)
		if err != nil {
			return nil, fmt.Errorf("signer account: [%s] doesn't exists in configuration", signerName)
		}
		if useRawSig {
			sig, err := hex.DecodeString(strings.ReplaceAll(signFlags.RawSig[i], "0x", ""))
			if err != nil {
				return nil, fmt.Errorf("invalid message signature: %w", err)
			}
			// If a raw signature is provided, use a dummy key that returns the signature instead of signing
			signer.Key = &sigAddKey{
				index:    signer.Key.Index(),
				signer:   crypto.NewAddSignatureSigner(sig, nil),
				sigAlgo:  signer.Key.SigAlgo(),
				hashAlgo: signer.Key.HashAlgo(),
				keyType:  signer.Key.Type(),
			}
		}
		signers = append(signers, signer)
	}

	//payer signs last
	sort.SliceStable(signers, func(i, j int) bool {
		return signers[i].Address.String() != tx.FlowTransaction().Payer.Hex()
	})

	for _, signer := range signers {
		if !globalFlags.Yes && !prompt.ApproveTransactionForSigningPrompt(tx.FlowTransaction()) {
			return nil, fmt.Errorf("transaction was not approved for signing")
		}

		signed, err = flow.SignTransactionPayload(context.Background(), signer, payload)
		if err != nil {
			return nil, err
		}

		payload = []byte(hex.EncodeToString(signed.FlowTransaction().Encode()))
	}

	if signFlags.FromRemoteUrl != "" {
		tx := signed.FlowTransaction()
		err = postRLPTransaction(filenameOrUrl, tx)

		if err != nil {
			return nil, err
		}
		fmt.Printf("%s Signed RLP Posted successfully\n", output.SuccessEmoji())
	}

	return &transactionResult{
		tx:      signed.FlowTransaction(),
		include: signFlags.Include,
		network: flow.Network().Name,
	}, nil
}

// getRLPTransaction payload from a remote server.
func getRLPTransaction(rlpUrl string) ([]byte, error) {
	client := http.Client{
		CheckRedirect: func(r *http.Request, via []*http.Request) error {
			r.URL.Opaque = r.URL.Path
			return nil
		},
	}
	resp, err := client.Get(rlpUrl)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("error downloading RLP identifier")
	}

	return io.ReadAll(resp.Body)
}

// postRLPTransaction signed payload to a remote server.
func postRLPTransaction(rlpUrl string, tx *flowsdk.Transaction) error {
	signedRlp := hex.EncodeToString(tx.Encode())
	resp, err := http.Post(rlpUrl, "application/text", bytes.NewBufferString(signedRlp))

	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("error posting signed RLP")
	}

	return nil
}

type sigAddKey struct {
	keyType  config.KeyType
	index    uint32
	sigAlgo  crypto.SignatureAlgorithm
	hashAlgo crypto.HashAlgorithm
	signer   crypto.Signer
}

var _ accounts.Key = &sigAddKey{}

func (a *sigAddKey) Type() config.KeyType {
	return a.keyType
}

func (a *sigAddKey) SigAlgo() crypto.SignatureAlgorithm {
	if a.sigAlgo == crypto.UnknownSignatureAlgorithm {
		return crypto.ECDSA_P256 // default value
	}
	return a.sigAlgo
}

func (a *sigAddKey) HashAlgo() crypto.HashAlgorithm {
	if a.hashAlgo == crypto.UnknownHashAlgorithm {
		return crypto.SHA3_256 // default value
	}
	return a.hashAlgo
}

func (a *sigAddKey) Index() uint32 {
	return a.index // default to 0
}

func (a *sigAddKey) Validate() error {
	return nil
}
func (a *sigAddKey) Signer(ctx context.Context) (crypto.Signer, error) {
	return a.signer, nil
}
func (a *sigAddKey) ToConfig() config.AccountKey {
	return config.AccountKey{
		Type:           a.keyType,
		Index:          a.index,
		SigAlgo:        a.sigAlgo,
		HashAlgo:       a.hashAlgo,
		Mnemonic:       "",
		DerivationPath: "",
	}
}
func (a *sigAddKey) PrivateKey() (*crypto.PrivateKey, error) {
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
