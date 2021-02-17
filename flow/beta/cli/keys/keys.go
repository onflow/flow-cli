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

package keys

import (
	"fmt"

	"github.com/onflow/flow-go-sdk/crypto"

	"github.com/onflow/flow-cli/flow/beta/cli/config"
)

type AccountKey interface {
	Type() config.KeyType
	Index() int
	SigAlgo() crypto.SignatureAlgorithm
	HashAlgo() crypto.HashAlgorithm
	Signer() crypto.Signer
}

func NewAccountKey(accountKeyConf config.AccountKey) (AccountKey, error) {
	switch accountKeyConf.Type {
	case config.KeyTypeHex:
		return newHexAccountKey(accountKeyConf)
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

type HexAccountKey struct {
	*baseAccountKey
	privateKey crypto.PrivateKey
}

const privateKeyField = "privateKey"

func newHexAccountKey(accountKeyConf config.AccountKey) (*HexAccountKey, error) {
	privateKeyHex, ok := accountKeyConf.Context[privateKeyField]
	if !ok {
		return nil, fmt.Errorf("\"%s\" field is required", privateKeyField)
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

func (a *HexAccountKey) Signer() crypto.Signer {
	return crypto.NewInMemorySigner(a.privateKey, a.HashAlgo())
}
