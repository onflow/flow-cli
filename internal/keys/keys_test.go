package keys

import (
	"github.com/onflow/flow-go-sdk/crypto"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestDecodeKeys(t *testing.T) {
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
}
