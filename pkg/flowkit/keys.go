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

package flowkit

import (
	"context"
	"encoding/hex"
	"fmt"

	"github.com/onflow/flow-cli/pkg/flowkit/util"

	"github.com/onflow/flow-go-sdk"
	"github.com/onflow/flow-go-sdk/crypto"
	"github.com/onflow/flow-go-sdk/crypto/cloudkms"

	"github.com/onflow/flow-cli/pkg/flowkit/config"
)

type AccountKey interface {
	Type() config.KeyType
	Index() int
	SigAlgo() crypto.SignatureAlgorithm
	HashAlgo() crypto.HashAlgorithm
	Signer(ctx context.Context) (crypto.Signer, error)
	ToConfig() config.AccountKey
	Validate() error
	PrivateKey() (*crypto.PrivateKey, error)
}

func NewAccountKey(accountKeyConf config.AccountKey) (AccountKey, error) {
	switch accountKeyConf.Type {
	case config.KeyTypeHex:
		return newHexAccountKey(accountKeyConf)
	case config.KeyTypeGoogleKMS:
		return newKmsAccountKey(accountKeyConf)
	}

	return nil, fmt.Errorf(`invalid key type: "%s"`, accountKeyConf.Type)
}

type baseAccountKey struct {
	keyType  config.KeyType
	index    int
	sigAlgo  crypto.SignatureAlgorithm
	hashAlgo crypto.HashAlgorithm
}

func newBaseAccountKey(accountKeyConf config.AccountKey) *baseAccountKey {
	return &baseAccountKey{
		keyType:  accountKeyConf.Type,
		index:    accountKeyConf.Index,
		sigAlgo:  accountKeyConf.SigAlgo,
		hashAlgo: accountKeyConf.HashAlgo,
	}
}

func (a *baseAccountKey) Type() config.KeyType {
	return a.keyType
}

func (a *baseAccountKey) SigAlgo() crypto.SignatureAlgorithm {
	return a.sigAlgo
}

func (a *baseAccountKey) HashAlgo() crypto.HashAlgorithm {
	return a.hashAlgo
}

func (a *baseAccountKey) Index() int {
	return a.index
}

func (a *baseAccountKey) Validate() error {
	return nil
}

type KmsAccountKey struct {
	*baseAccountKey
	kmsKey cloudkms.Key
}

func (a *KmsAccountKey) ToConfig() config.AccountKey {
	return config.AccountKey{
		Type:       a.keyType,
		Index:      a.index,
		SigAlgo:    a.sigAlgo,
		HashAlgo:   a.hashAlgo,
		ResourceID: a.kmsKey.ResourceID(),
	}
}

func (a *KmsAccountKey) Signer(ctx context.Context) (crypto.Signer, error) {
	kmsClient, err := cloudkms.NewClient(ctx)
	if err != nil {
		return nil, err
	}

	accountKMSSigner, err := kmsClient.SignerForKey(
		ctx,
		flow.Address{}, // TODO: this is temporary workaround as SignerForKey accepts address but never uses it so should be removed
		a.kmsKey,
	)
	if err != nil {
		return nil, err
	}

	return accountKMSSigner, nil
}

func (a *KmsAccountKey) Validate() error {
	return util.GcloudApplicationSignin(a.ToConfig().ResourceID)
}

func (a *KmsAccountKey) PrivateKey() (*crypto.PrivateKey, error) {
	return nil, fmt.Errorf("private key not accessible")
}

func newKmsAccountKey(key config.AccountKey) (AccountKey, error) {
	accountKMSKey, err := cloudkms.KeyFromResourceID(key.ResourceID)
	if err != nil {
		return nil, err
	}

	return &KmsAccountKey{
		baseAccountKey: &baseAccountKey{
			keyType:  config.KeyTypeGoogleKMS,
			index:    key.Index,
			sigAlgo:  key.SigAlgo,
			hashAlgo: key.HashAlgo,
		},
		kmsKey: accountKMSKey,
	}, nil
}

func NewHexAccountKeyFromPrivateKey(
	index int,
	hashAlgo crypto.HashAlgorithm,
	privateKey crypto.PrivateKey,
) *HexAccountKey {
	return &HexAccountKey{
		baseAccountKey: &baseAccountKey{
			keyType:  config.KeyTypeHex,
			index:    index,
			sigAlgo:  privateKey.Algorithm(),
			hashAlgo: hashAlgo,
		},
		privateKey: privateKey,
	}
}

func newHexAccountKey(accountKey config.AccountKey) (*HexAccountKey, error) {
	return &HexAccountKey{
		baseAccountKey: newBaseAccountKey(accountKey),
		privateKey:     accountKey.PrivateKey,
	}, nil
}

type HexAccountKey struct {
	*baseAccountKey
	privateKey crypto.PrivateKey
}

func (a *HexAccountKey) Signer(ctx context.Context) (crypto.Signer, error) {
	return crypto.NewInMemorySigner(a.privateKey, a.HashAlgo()), nil
}

func (a *HexAccountKey) PrivateKey() (*crypto.PrivateKey, error) {
	return &a.privateKey, nil
}

func (a *HexAccountKey) ToConfig() config.AccountKey {
	return config.AccountKey{
		Type:       a.keyType,
		Index:      a.index,
		SigAlgo:    a.sigAlgo,
		HashAlgo:   a.hashAlgo,
		PrivateKey: a.privateKey,
	}
}

func (a *HexAccountKey) Validate() error {
	_, err := crypto.DecodePrivateKeyHex(a.sigAlgo, a.PrivateKeyHex())
	if err != nil {
		return fmt.Errorf("invalid private key")
	}

	return nil
}

func (a *HexAccountKey) PrivateKeyHex() string {
	return hex.EncodeToString(a.privateKey.Encode())
}
