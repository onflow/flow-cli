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

package accounts

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/onflow/flow-cli/flowkit/config"
)

func Test_KMS_Keys(t *testing.T) {
	confKey := config.AccountKey{
		Type:       config.KeyTypeGoogleKMS,
		Index:      0,
		SigAlgo:    config.DefaultSigAlgo,
		HashAlgo:   config.DefaultHashAlgo,
		ResourceID: "projects/my-project/locations/global/keyRings/flow/cryptoKeys/my-account/cryptoKeyVersions/1",
	}

	kmsKey, err := kmsKeyFromConfig(confKey)
	assert.NoError(t, err)

	_, err = kmsKey.PrivateKey()
	assert.EqualError(t, err, "private key not accessible")
	assert.Equal(t, confKey, kmsKey.ToConfig())
}

func Test_File_key(t *testing.T) {
	confKey := config.AccountKey{
		Type:     config.KeyTypeFile,
		Index:    0,
		SigAlgo:  config.DefaultSigAlgo,
		HashAlgo: config.DefaultHashAlgo,
		Location: "./test.pkey",
	}

	fileKey, err := fileKeyFromConfig(confKey)
	assert.NoError(t, err)

	cKey := fileKey.ToConfig()
	assert.Equal(t, cKey, confKey)

	key := NewFileKey(confKey.Location, confKey.Index, confKey.SigAlgo, confKey.HashAlgo)
	assert.Equal(t, confKey, key.ToConfig())
}

func Test_BIP44(t *testing.T) {
	confKey := config.AccountKey{
		Type:           config.KeyTypeBip44,
		Index:          0,
		SigAlgo:        config.DefaultSigAlgo,
		HashAlgo:       config.DefaultHashAlgo,
		Mnemonic:       "version field tornado move level pretty inject stereo ten catalog salon swallow",
		DerivationPath: "m/44'/539'/0'/0/0",
	}

	key, err := bip44KeyFromConfig(confKey)
	assert.NoError(t, err)

	const pubKey = "0x2d6daea8b0ba5b1d5935f7846ccdd7e6f9f981e34d3c0a02a927cc79c837eba56c0f9a979195e41143495b72314ffcab60da6b7031060c80dc12f01f7f2096be"
	assert.Equal(t, confKey, key.ToConfig())
	pkey, err := key.PrivateKey()
	assert.NoError(t, err)
	assert.Equal(t, pubKey, (*pkey).PublicKey().String())

	sig, err := key.Signer(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, pubKey, sig.PublicKey().String())
}
