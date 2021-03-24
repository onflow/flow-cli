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

package keys

import (
	"context"
	"encoding/hex"
	"fmt"

	"github.com/onflow/flow-go-sdk"

	"github.com/onflow/flow-go-sdk/crypto/cloudkms"

	"github.com/onflow/flow-cli/pkg/flow/config"
	"github.com/onflow/flow-go-sdk/crypto"
)

type AccountKey interface {
	Type() config.KeyType
	Index() int
	SigAlgo() crypto.SignatureAlgorithm
	HashAlgo() crypto.HashAlgorithm
	Signer(ctx context.Context) (crypto.Signer, error)
	ToConfig() config.AccountKey
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

type KmsAccountKey struct {
	*baseAccountKey
	kmsKey cloudkms.Key
}

func (a *KmsAccountKey) ToConfig() config.AccountKey {
	return config.AccountKey{
		Type:     a.keyType,
		Index:    a.index,
		SigAlgo:  a.sigAlgo,
		HashAlgo: a.hashAlgo,
		Context: map[string]string{
			config.KMSContextField: a.kmsKey.ResourceID(),
		},
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

func KeyContextFromKMSResourceID(resourceID string) (map[string]string, error) {
	ctx := make(map[string]string)
	ctx[config.KMSContextField] = resourceID

	_, err := cloudkms.KeyFromResourceID(resourceID)
	if err != nil {
		return nil, err
	}

	return ctx, nil
}

func newKmsAccountKey(key config.AccountKey) (AccountKey, error) {
	accountKMSKey, err := cloudkms.KeyFromResourceID(key.Context[config.KMSContextField])
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

func newHexAccountKey(accountKeyConf config.AccountKey) (*HexAccountKey, error) {
	privateKeyHex, ok := accountKeyConf.Context[config.PrivateKeyField]
	if !ok {
		return nil, fmt.Errorf("\"%s\" field is required", config.PrivateKeyField)
	}

	privateKey, err := crypto.DecodePrivateKeyHex(accountKeyConf.SigAlgo, privateKeyHex)
	if err != nil {
		return nil, err
	}

	return &HexAccountKey{
		baseAccountKey: newBaseAccountKey(accountKeyConf),
		privateKey:     privateKey,
	}, nil
}

type HexAccountKey struct {
	*baseAccountKey
	privateKey crypto.PrivateKey
}

func (a *HexAccountKey) Signer(ctx context.Context) (crypto.Signer, error) {
	return crypto.NewInMemorySigner(a.privateKey, a.HashAlgo()), nil
}

func (a *HexAccountKey) PrivateKey() crypto.PrivateKey {
	return a.privateKey
}

func (a *HexAccountKey) ToConfig() config.AccountKey {
	return config.AccountKey{
		Type:     a.keyType,
		Index:    a.index,
		SigAlgo:  a.sigAlgo,
		HashAlgo: a.hashAlgo,
		Context: map[string]string{
			config.PrivateKeyField: a.PrivateKeyHex(),
		},
	}
}

func (a *HexAccountKey) PrivateKeyHex() string {
	return hex.EncodeToString(a.privateKey.Encode())
}
