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
