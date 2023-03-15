package keys

import (
	"github.com/onflow/flow-cli/internal/util"
	"github.com/onflow/flow-go-sdk/crypto"
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_DecodeKeys(t *testing.T) {
	t.Run("Decode RLP Key", func(t *testing.T) {
		t.Parallel()
		dkey, err := decodeRLP("f847b84084d716c14b051ad6b001624f738f5d302636e6b07cc75e4530af7776a4368a2b586dbefc0564ee28384c2696f178cbed52e62811bcc9ecb59568c996d342db2402038203e8")

		assert.NoError(t, err)
		assert.Equal(t, dkey.PublicKey.String(), "0x84d716c14b051ad6b001624f738f5d302636e6b07cc75e4530af7776a4368a2b586dbefc0564ee28384c2696f178cbed52e62811bcc9ecb59568c996d342db24")
		assert.Equal(t, dkey.SigAlgo.String(), "ECDSA_P256")
	})

	t.Run("Decode RLP Key Invalid", func(t *testing.T) {
		t.Parallel()

		_, err := decodeRLP("aaa")
		assert.EqualError(t, err, "failed to decode public key: encoding/hex: odd length hex string")
	})

	t.Run("Decode PEM Key", func(t *testing.T) {
		t.Parallel()

		dkey, err := decodePEM("-----BEGIN PUBLIC KEY-----\nMFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAE1HmzzcntvdsZXLErNRYa3oJrAypk\nvdQGLMh/s7p+ccnPZG/yOZC7RTLKRcRFx+kIzvJ4ssRhU2ADmmZgo2apXw==\n-----END PUBLIC KEY-----", crypto.ECDSA_P256)

		assert.NoError(t, err)
		assert.Equal(t, dkey.PublicKey.String(), "0xd479b3cdc9edbddb195cb12b35161ade826b032a64bdd4062cc87fb3ba7e71c9cf646ff23990bb4532ca45c445c7e908cef278b2c4615360039a6660a366a95f")
		assert.Equal(t, dkey.SigAlgo.String(), "ECDSA_P256")
	})

	t.Run("Decode PEM Key Invalid", func(t *testing.T) {
		t.Parallel()

		_, err := decodePEM("nope", crypto.ECDSA_P256)
		assert.EqualError(t, err, "crypto: failed to parse PEM string, not all bytes in PEM key were decoded: 6e6f7065")
	})

	srv, _, rw := util.TestMocks(t)
	t.Run("Success", func(t *testing.T) {
		inArgsTests := [][]string{
			{"rlp", "f847b84084d716c14b051ad6b001624f738f5d302636e6b07cc75e4530af7776a4368a2b586dbefc0564ee28384c2696f178cbed52e62811bcc9ecb59568c996d342db2402038203e8"},
			{"pem", "-----BEGIN PUBLIC KEY-----\nMFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAE1HmzzcntvdsZXLErNRYa3oJrAypk\nvdQGLMh/s7p+ccnPZG/yOZC7RTLKRcRFx+kIzvJ4ssRhU2ADmmZgo2apXw==\n-----END PUBLIC KEY-----"},
		}

		for _, inArgs := range inArgsTests {
			result, err := decode(inArgs, util.NoFlags, util.NoLogger, rw, srv.Mock)
			assert.NoError(t, err)
			assert.NotNil(t, result)
		}
	})

	t.Run("Success from file", func(t *testing.T) {
		inArgs := []string{"rlp"}
		_ = rw.WriteFile("test", []byte("f847b84084d716c14b051ad6b001624f738f5d302636e6b07cc75e4530af7776a4368a2b586dbefc0564ee28384c2696f178cbed52e62811bcc9ecb59568c996d342db2402038203e8"), 0677)
		decodeFlags.FromFile = "test"

		result, err := decode(inArgs, util.NoFlags, util.NoLogger, rw, srv.Mock)
		assert.NoError(t, err)
		assert.NotNil(t, result)
		decodeFlags.FromFile = "" // reset to default
	})

	t.Run("Fail invalid args", func(t *testing.T) {
		inArgs := []string{"invalid", "invalid"}
		result, err := decode(inArgs, util.NoFlags, util.NoLogger, rw, srv.Mock)
		assert.EqualError(t, err, "encoding type not supported. Valid encoding: RLP and PEM")
		assert.Nil(t, result)
	})

	t.Run("Fail invalid args", func(t *testing.T) {
		inArgs := []string{"", ""}
		result, err := decode(inArgs, util.NoFlags, util.NoLogger, rw, srv.Mock)
		assert.EqualError(t, err, "provide argument for encoded key or use from file flag")
		assert.Nil(t, result)
	})

	t.Run("Fail invalid flags", func(t *testing.T) {
		inArgs := []string{"rlp", "some public key"}
		decodeFlags.FromFile = "from file"

		result, err := decode(inArgs, util.NoFlags, util.NoLogger, rw, srv.Mock)
		assert.EqualError(t, err, "can not pass both command argument and from file flag")
		assert.Nil(t, result)
	})
}

func Test_DeriveKeys(t *testing.T) {
	srv, _, rw := util.TestMocks(t)

	t.Run("Success", func(t *testing.T) {
		inArgs := []string{"cf3178b20a73846dc8bf6255c79be47178b0744dd8244bcff099e449a9700d7f"}
		result, err := derive(inArgs, util.NoFlags, util.NoLogger, rw, srv.Mock)
		assert.NoError(t, err)
		assert.NotNil(t, result)
	})

	t.Run("Fail invalid key", func(t *testing.T) {
		inArgs := []string{"invalid"}

		result, err := derive(inArgs, util.NoFlags, util.NoLogger, rw, srv.Mock)
		assert.EqualError(t, err, "failed to decode private key: encoding/hex: invalid byte: U+0069 'i'")
		assert.Nil(t, result)
	})

	t.Run("Fail invalid signature algorithm", func(t *testing.T) {
		inArgs := []string{"cf3178b20a73846dc8bf6255c79be47178b0744dd8244bcff099e449a9700d7f"}
		deriveFlags.KeySigAlgo = "invalid"

		result, err := derive(inArgs, util.NoFlags, util.NoLogger, rw, srv.Mock)
		assert.EqualError(t, err, "invalid signature algorithm: invalid")
		assert.Nil(t, result)
	})
}

func Test_Generate(t *testing.T) {
	srv, _, rw := util.TestMocks(t)

	t.Run("Fail invalid signature algorithm", func(t *testing.T) {
		generateFlags.KeySigAlgo = "invalid"
		_, err := generate([]string{}, util.NoFlags, util.NoLogger, rw, srv.Mock)
		assert.EqualError(t, err, "invalid signature algorithm: invalid")
	})
}
