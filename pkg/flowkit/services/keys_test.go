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

package services

import (
	"encoding/hex"
	"testing"

	goeth "github.com/ethereum/go-ethereum/accounts"
	slip10 "github.com/lmars/go-slip10"

	"github.com/onflow/flow-go-sdk/crypto"
	"github.com/stretchr/testify/assert"
)

func TestKeys(t *testing.T) {
	t.Parallel()

	t.Run("Generate Keys", func(t *testing.T) {
		t.Parallel()

		_, s, _ := setup()
		key, err := s.Keys.Generate("", crypto.ECDSA_P256)
		assert.NoError(t, err)

		assert.Equal(t, len(key.String()), 66)
	})

	t.Run("Generate Keys with seed", func(t *testing.T) {
		t.Parallel()

		_, s, _ := setup()
		key, err := s.Keys.Generate("aaaaaaaaaaaaaaaaaaaaaaannndddddd_its_gone", crypto.ECDSA_P256)

		assert.NoError(t, err)
		assert.Equal(t, key.String(), "0x134f702d0872dba9c7aea15498aab9b2ffedd5aeebfd8ac3cf47c591f0d7ce52")
	})

	t.Run("Test Slip10 - P256", func(t *testing.T) {
		t.Parallel()

		curve := slip10.CurveP256
		sigAlgo := crypto.ECDSA_P256
		seed := []byte{0x0, 0x1, 0x2, 0x3, 0x4, 0x5, 0x6, 0x7, 0x8, 0x9, 0xa, 0xb, 0xc, 0xd, 0x0e, 0x0f}
		path, _ := goeth.ParseDerivationPath("m/0'/1/2'/2/1000000000")

		accountKey, _ := slip10.NewMasterKeyWithCurve(seed, curve)
		// https://github.com/satoshilabs/slips/blob/master/slip-0010.md#test-vector-1-for-nist256p1
		privateKey, _ := crypto.DecodePrivateKey(sigAlgo, accountKey.Key)

		assert.Equal(t, privateKey.String(), "0x612091aaa12e22dd2abef664f8a01a82cae99ad7441b7ef8110424915c268bc2")

		expectedPrivateKeys := []string{
			"0x6939694369114c67917a182c59ddb8cafc3004e63ca5d3b84403ba8613debc0c",
			"0x284e9d38d07d21e4e281b645089a94f4cf5a5a81369acf151a1c3a57f18b2129",
			"0x694596e8a54f252c960eb771a3c41e7e32496d03b954aeb90f61635b8e092aa7",
			"0x5996c37fd3dd2679039b23ed6f70b506c6b56b3cb5e424681fb0fa64caf82aaa",
			"0x21c4f269ef0a5fd1badf47eeacebeeaa3de22eb8e5b0adcd0f27dd99d34d0119",
		}
		for i, n := range path {
			accountKey, _ = accountKey.NewChildKey(n)
			privateKey, _ := crypto.DecodePrivateKey(sigAlgo, accountKey.Key)

			assert.Equal(t, privateKey.String(), expectedPrivateKeys[i])

		}
	})

	t.Run("Test Slip10 - secp256k1", func(t *testing.T) {
		t.Parallel()

		curve := slip10.CurveBitcoin
		sigAlgo := crypto.ECDSA_secp256k1
		seed := []byte{0x0, 0x1, 0x2, 0x3, 0x4, 0x5, 0x6, 0x7, 0x8, 0x9, 0xa, 0xb, 0xc, 0xd, 0x0e, 0x0f}
		path, _ := goeth.ParseDerivationPath("m/0'/1/2'/2/1000000000")

		accountKey, _ := slip10.NewMasterKeyWithCurve(seed, curve)
		// https://github.com/satoshilabs/slips/blob/master/slip-0010.md#test-vector-1-for-nist256p1
		privateKey, _ := crypto.DecodePrivateKey(sigAlgo, accountKey.Key)

		assert.Equal(t, privateKey.String(), "0xe8f32e723decf4051aefac8e2c93c9c5b214313817cdb01a1494b917c8436b35")

		expectedPrivateKeys := []string{
			"0xedb2e14f9ee77d26dd93b4ecede8d16ed408ce149b6cd80b0715a2d911a0afea",
			"0x3c6cb8d0f6a264c91ea8b5030fadaa8e538b020f0a387421a12de9319dc93368",
			"0xcbce0d719ecf7431d88e6a89fa1483e02e35092af60c042b1df2ff59fa424dca",
			"0x0f479245fb19a38a1954c5c7c0ebab2f9bdfd96a17563ef28a6a4b1a2a764ef4",
			"0x471b76e389e528d6de6d816857e012c5455051cad6660850e58372a6c3e6e7c8",
		}
		for i, n := range path {
			accountKey, _ = accountKey.NewChildKey(n)
			privateKey, _ := crypto.DecodePrivateKey(sigAlgo, accountKey.Key)

			assert.Equal(t, privateKey.String(), expectedPrivateKeys[i])

		}
	})

	t.Run("Generate Keys with mnemonic (default path)", func(t *testing.T) {
		t.Parallel()

		_, s, _ := setup()
		key, err := s.Keys.DerivePrivateKeyFromMnemonic("normal dune pole key case cradle unfold require tornado mercy hospital buyer", crypto.ECDSA_P256, "")

		assert.NoError(t, err)
		assert.Equal(t, key.String(), "0x638dc9ad0eee91d09249f0fd7c5323a11600e20d5b9105b66b782a96236e74cf")
	})

	//https://github.com/onflow/ledger-app-flow/blob/dc61213a9c3d73152b78b7391d04165d07f1ad89/tests_speculos/test-basic-show-address-expert.js#L28
	t.Run("Generate Keys with mnemonic (custom path - Ledger)", func(t *testing.T) {
		t.Parallel()

		_, s, _ := setup()
		//ledger test mnemonic: https://github.com/onflow/ledger-app-flow#using-a-real-device-for-integration-tests-nano-s-and-nano-s-plus
		key, err := s.Keys.DerivePrivateKeyFromMnemonic("equip will roof matter pink blind book anxiety banner elbow sun young", crypto.ECDSA_secp256k1, "m/44'/539'/513'/0/0")

		assert.NoError(t, err)
		assert.Equal(t, key.String(), "0xd18d051afca7150781fef111f3387d132d31c4a6250268db0f61f836a205e0b8")

		assert.Equal(t, hex.EncodeToString(key.PublicKey().Encode()), "d7482bbaff7827035d5b238df318b10604673dc613808723efbd23fbc4b9fad34a415828d924ec7b83ac0eddf22ef115b7c203ee39fb080572d7e51775ee54be")
	})

	t.Run("Generate Keys with private key", func(t *testing.T) {
		t.Parallel()

		_, s, _ := setup()
		key, err := s.Keys.ParsePrivateKey("af232020ea7a7256eebdcebd609457d0dea51436a4377d2b577a3cf1f6d45c44", crypto.ECDSA_P256)

		assert.NoError(t, err)
		assert.Equal(t, key.String(), "0xaf232020ea7a7256eebdcebd609457d0dea51436a4377d2b577a3cf1f6d45c44")
		assert.Equal(t, key.PublicKey().String(), "0x3da1d2eb3d9f1a0f57b434dca6bac2068216ccc5c69221a70f5c060152a39296ad28ad260536977f88eea45da9064b81a18c17f5cdc30e638752767359f0b496")
	})

	t.Run("Generate Keys Invalid", func(t *testing.T) {
		t.Parallel()

		_, s, _ := setup()
		inputs := []map[string]crypto.SignatureAlgorithm{
			{"im_short": crypto.ECDSA_P256},
			{"": crypto.StringToSignatureAlgorithm("JUSTNO")},
		}

		errs := []string{
			"failed to generate private key: crypto: insufficient seed length 8, must be at least 32 bytes for ECDSA_P256",
			"failed to generate private key: crypto: Go SDK does not support key generation for UNKNOWN algorithm",
		}

		for x, in := range inputs {
			for k, v := range in {
				_, err := s.Keys.Generate(k, v)
				assert.Equal(t, err.Error(), errs[x])
				x++
			}
		}

	})

	t.Run("Decode RLP Key", func(t *testing.T) {
		t.Parallel()

		_, s, _ := setup()
		dkey, err := s.Keys.DecodeRLP("f847b84084d716c14b051ad6b001624f738f5d302636e6b07cc75e4530af7776a4368a2b586dbefc0564ee28384c2696f178cbed52e62811bcc9ecb59568c996d342db2402038203e8")

		assert.NoError(t, err)
		assert.Equal(t, dkey.PublicKey.String(), "0x84d716c14b051ad6b001624f738f5d302636e6b07cc75e4530af7776a4368a2b586dbefc0564ee28384c2696f178cbed52e62811bcc9ecb59568c996d342db24")
		assert.Equal(t, dkey.SigAlgo.String(), "ECDSA_P256")
	})

	t.Run("Decode RLP Key Invalid", func(t *testing.T) {
		t.Parallel()

		_, s, _ := setup()
		_, err := s.Keys.DecodeRLP("aaa")
		assert.Equal(t, err.Error(), "failed to decode public key: encoding/hex: odd length hex string")
	})

	t.Run("Decode PEM Key", func(t *testing.T) {
		t.Parallel()

		_, s, _ := setup()
		dkey, err := s.Keys.DecodePEM("-----BEGIN PUBLIC KEY-----\nMFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAE1HmzzcntvdsZXLErNRYa3oJrAypk\nvdQGLMh/s7p+ccnPZG/yOZC7RTLKRcRFx+kIzvJ4ssRhU2ADmmZgo2apXw==\n-----END PUBLIC KEY-----", crypto.ECDSA_P256)

		assert.NoError(t, err)
		assert.Equal(t, dkey.PublicKey.String(), "0xd479b3cdc9edbddb195cb12b35161ade826b032a64bdd4062cc87fb3ba7e71c9cf646ff23990bb4532ca45c445c7e908cef278b2c4615360039a6660a366a95f")
		assert.Equal(t, dkey.SigAlgo.String(), "ECDSA_P256")
	})

	t.Run("Decode PEM Key Invalid", func(t *testing.T) {
		t.Parallel()

		_, s, _ := setup()
		_, err := s.Keys.DecodePEM("nope", crypto.ECDSA_P256)
		assert.Equal(t, err.Error(), "crypto: failed to parse PEM string, not all bytes in PEM key were decoded: 6e6f7065")
	})
}
