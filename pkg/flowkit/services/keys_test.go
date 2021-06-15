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

package services

import (
	"testing"

	"github.com/onflow/flow-go-sdk/crypto"
	"github.com/stretchr/testify/assert"
)

func TestKeys(t *testing.T) {

	t.Run("Generate Keys", func(t *testing.T) {
		_, s, _ := setup()
		key, err := s.Keys.Generate("", crypto.ECDSA_P256)
		assert.NoError(t, err)

		assert.Equal(t, len(key.String()), 66)
	})

	t.Run("Generate Keys with seed", func(t *testing.T) {
		_, s, _ := setup()
		key, err := s.Keys.Generate("aaaaaaaaaaaaaaaaaaaaaaannndddddd_its_gone", crypto.ECDSA_P256)

		assert.NoError(t, err)
		assert.Equal(t, key.String(), "0x134f702d0872dba9c7aea15498aab9b2ffedd5aeebfd8ac3cf47c591f0d7ce52")
	})

	t.Run("Fail generate keys, too short seed", func(t *testing.T) {
		_, s, _ := setup()
		_, err := s.Keys.Generate("im_short", crypto.ECDSA_P256)

		assert.Equal(t, err.Error(), "failed to generate private key: crypto: insufficient seed length 8, must be at least 32 bytes for ECDSA_P256")
	})

	t.Run("Fail generate keys, invalid sig algo", func(t *testing.T) {
		_, s, _ := setup()
		_, err := s.Keys.Generate("", crypto.StringToSignatureAlgorithm("JUSTNO"))

		assert.Equal(t, err.Error(), "failed to generate private key: crypto: Go SDK does not support key generation for UNKNOWN algorithm")
	})

	t.Run("RLP decode keys", func(t *testing.T) {
		_, s, _ := setup()
		dkey, err := s.Keys.DecodeRLP("f847b84084d716c14b051ad6b001624f738f5d302636e6b07cc75e4530af7776a4368a2b586dbefc0564ee28384c2696f178cbed52e62811bcc9ecb59568c996d342db2402038203e8")

		assert.NoError(t, err)
		assert.Equal(t, dkey.PublicKey.String(), "0x84d716c14b051ad6b001624f738f5d302636e6b07cc75e4530af7776a4368a2b586dbefc0564ee28384c2696f178cbed52e62811bcc9ecb59568c996d342db24")
		assert.Equal(t, dkey.SigAlgo.String(), "ECDSA_P256")
	})
}
