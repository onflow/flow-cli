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

	t.Run("Test Vector SLIP-0010", func(t *testing.T) {
		// test against SLIP-0010 test vector. All data are taken from:
		//  https://github.com/satoshilabs/slips/blob/master/slip-0010.md#test-vectors
		t.Parallel()
		_, s, _ := setup()

		type testEntry struct {
			sigAlgo    crypto.SignatureAlgorithm
			seed       string
			path       string
			privateKey string
		}

		testVector := []testEntry{
			testEntry{
				sigAlgo:    crypto.ECDSA_secp256k1,
				seed:       "000102030405060708090a0b0c0d0e0f",
				path:       "m/0'/1/2'/2/1000000000",
				privateKey: "0x471b76e389e528d6de6d816857e012c5455051cad6660850e58372a6c3e6e7c8",
			},
			testEntry{
				sigAlgo:    crypto.ECDSA_P256,
				seed:       "000102030405060708090a0b0c0d0e0f",
				path:       "m/0'/1/2'/2/1000000000",
				privateKey: "0x21c4f269ef0a5fd1badf47eeacebeeaa3de22eb8e5b0adcd0f27dd99d34d0119",
			},
			testEntry{
				sigAlgo:    crypto.ECDSA_secp256k1,
				seed:       "fffcf9f6f3f0edeae7e4e1dedbd8d5d2cfccc9c6c3c0bdbab7b4b1aeaba8a5a29f9c999693908d8a8784817e7b7875726f6c696663605d5a5754514e4b484542",
				path:       "m/0/2147483647'/1/2147483646'/2",
				privateKey: "0xbb7d39bdb83ecf58f2fd82b6d918341cbef428661ef01ab97c28a4842125ac23",
			},
			testEntry{
				sigAlgo:    crypto.ECDSA_P256,
				seed:       "fffcf9f6f3f0edeae7e4e1dedbd8d5d2cfccc9c6c3c0bdbab7b4b1aeaba8a5a29f9c999693908d8a8784817e7b7875726f6c696663605d5a5754514e4b484542",
				path:       "m/0/2147483647'/1/2147483646'/2",
				privateKey: "0xbb0a77ba01cc31d77205d51d08bd313b979a71ef4de9b062f8958297e746bd67",
			},
		}

		for _, test := range testVector {
			seed, err := hex.DecodeString(test.seed)
			assert.NoError(t, err)
			// use derivePrivateKeyFromSeed to test instead of DerivePrivateKeyFromMnemonic
			// because the test vector provides seeds, while it's not possible to derive mnemonics
			// corresponding to seeds.
			privateKey, err := s.Keys.derivePrivateKeyFromSeed(seed, test.sigAlgo, test.path)
			assert.NoError(t, err)
			assert.Equal(t, test.privateKey, privateKey.String())
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
