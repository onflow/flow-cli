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

package signatures

import (
	"github.com/onflow/flow-cli/internal/util"
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_Verify(t *testing.T) {
	srv, state, _ := util.TestMocks(t)

	t.Run("Success", func(t *testing.T) {
		inArgs := []string{
			"test signature",
			"f80f6007dbe6795bcf343e5586d40d0ba26a6c1d7edda5653cbdb377c9c20034cdbf899bb20fa2388d4993f6c88b5c97cbe05963d6d9799e6868902c2c14bc22",
			"0xab70e9e341a38861fd7f9fb1cda4c560465cfeb3ce4abcd2be552550c85ebbef9a2a9cac731c6bfa73b10c701c93038f0c18253487d4962d3bc6d5291f9c5eae",
		}

		result, err := verify(inArgs, util.NoFlags, util.NoLogger, srv.Mock, state)
		assert.NoError(t, err)
		assert.NotNil(t, result)
	})

	t.Run("Invalid signature", func(t *testing.T) {
		inArgsTests := []struct {
			args []string
			err  string
		}{{
			args: []string{"invalid", "invalid"},
			err:  "invalid message signature: encoding/hex: invalid byte: U+0069 'i'",
		}, {
			args: []string{"invalid", "0xaaaa", "invalid"},
			err:  "invalid public key: encoding/hex: invalid byte: U+0069 'i'",
		}, {
			args: []string{"invalid", "0xaaaa", "0x1234"},
			err:  "invalid public key: input has incorrect ECDSA_P256 key size, got 2, expects 64",
		}}

		for _, test := range inArgsTests {
			result, err := verify(test.args, util.NoFlags, util.NoLogger, srv.Mock, state)
			assert.EqualError(t, err, test.err)
			assert.Nil(t, result)
		}
	})

}

func Test_Sign(t *testing.T) {
	srv, state, _ := util.TestMocks(t)

	t.Run("Success", func(t *testing.T) {
		inArgs := []string{"test message"}

		result, err := sign(inArgs, util.NoFlags, util.NoLogger, srv.Mock, state)
		assert.NoError(t, err)
		assert.NotNil(t, result)
	})

	t.Run("Fail unknown signer", func(t *testing.T) {
		inArgs := []string{"test message"}
		generateFlags.Signer = "invalid"
		result, err := sign(inArgs, util.NoFlags, util.NoLogger, srv.Mock, state)
		assert.EqualError(t, err, "could not find account with name invalid in the configuration")
		assert.Nil(t, result)
	})
}
