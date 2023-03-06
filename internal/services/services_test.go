package services

import (
	"github.com/onflow/cadence"
	"github.com/onflow/flow-cli/pkg/flowkit"
	"github.com/onflow/flow-cli/pkg/flowkit/config"
	"github.com/onflow/flow-cli/pkg/flowkit/output"
	"github.com/onflow/flow-cli/pkg/flowkit/tests"
	"github.com/onflow/flow-go-sdk"
	"github.com/onflow/flow-go-sdk/crypto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"strings"
	"testing"
)

func setup() (*flowkit.State, Services, *tests.TestGateway) {
	readerWriter, _ := tests.ReaderWriter()
	state, err := flowkit.Init(readerWriter, crypto.ECDSA_P256, crypto.SHA3_256)
	if err != nil {
		panic(err)
	}

	gw := tests.DefaultMockGateway()
	srv := Services{
		state:   state,
		network: config.DefaultEmulatorNetwork(),
		gateway: gw,
		logger:  output.NewStdoutLogger(output.NoneLog),
	}

	return state, srv, gw
}

func TestAccounts(t *testing.T) {

	t.Run("Staking Info for Account", func(t *testing.T) {
		_, srv, gw := setup()

		count := 0
		gw.ExecuteScript.Run(func(args mock.Arguments) {
			count++
			assert.True(t, strings.Contains(string(args.Get(0).([]byte)), "import FlowIDTableStaking from 0x9eca2b38b18b5dfe"))
			gw.ExecuteScript.Return(cadence.NewArray([]cadence.Value{}), nil)
		})

		val1, val2, err := srv.StakingInfo(flow.HexToAddress("df9c30eb2252f1fa"))
		assert.NoError(t, err)
		assert.NotNil(t, val1)
		assert.NotNil(t, val2)
		assert.Equal(t, 2, count)
	})
	t.Run("Staking Info for Account fetches node total", func(t *testing.T) {
		_, srv, gw := setup()

		count := 0
		gw.ExecuteScript.Run(func(args mock.Arguments) {
			assert.True(t, strings.Contains(string(args.Get(0).([]byte)), "import FlowIDTableStaking from 0x9eca2b38b18b5dfe"))
			if count < 2 {
				gw.ExecuteScript.Return(cadence.NewArray(
					[]cadence.Value{
						cadence.Struct{
							StructType: &cadence.StructType{
								Fields: []cadence.Field{
									{
										Identifier: "id",
									},
								},
							},
							Fields: []cadence.Value{
								cadence.String("8f4d09dae7918afbf62c48fa968a9e8b0891cee8442065fa47cc05f4bc9a8a91"),
							},
						},
					}), nil)
			} else {
				assert.True(t, strings.Contains(args.Get(1).([]cadence.Value)[0].String(), "8f4d09dae7918afbf62c48fa968a9e8b0891cee8442065fa47cc05f4bc9a8a91"))
				gw.ExecuteScript.Return(cadence.NewUFix64("1.0"))
			}
			count++
		})

		val1, val2, err := srv.StakingInfo(flow.HexToAddress("df9c30eb2252f1fa"))
		assert.NoError(t, err)
		assert.NotNil(t, val1)
		assert.NotNil(t, val2)
		assert.Equal(t, 3, count)
	})
}

func TestKeys(t *testing.T) {

	t.Run("Decode RLP Key", func(t *testing.T) {
		t.Parallel()

		_, srv, _ := setup()
		dkey, err := srv.DecodeRLPKey("f847b84084d716c14b051ad6b001624f738f5d302636e6b07cc75e4530af7776a4368a2b586dbefc0564ee28384c2696f178cbed52e62811bcc9ecb59568c996d342db2402038203e8")

		assert.NoError(t, err)
		assert.Equal(t, dkey.PublicKey.String(), "0x84d716c14b051ad6b001624f738f5d302636e6b07cc75e4530af7776a4368a2b586dbefc0564ee28384c2696f178cbed52e62811bcc9ecb59568c996d342db24")
		assert.Equal(t, dkey.SigAlgo.String(), "ECDSA_P256")
	})

	t.Run("Decode RLP Key Invalid", func(t *testing.T) {
		t.Parallel()

		_, srv, _ := setup()
		_, err := srv.DecodeRLPKey("aaa")
		assert.Equal(t, err.Error(), "failed to decode public key: encoding/hex: odd length hex string")
	})

	t.Run("Decode PEM Key", func(t *testing.T) {
		t.Parallel()

		_, srv, _ := setup()
		dkey, err := srv.DecodePEMKey("-----BEGIN PUBLIC KEY-----\nMFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAE1HmzzcntvdsZXLErNRYa3oJrAypk\nvdQGLMh/s7p+ccnPZG/yOZC7RTLKRcRFx+kIzvJ4ssRhU2ADmmZgo2apXw==\n-----END PUBLIC KEY-----", crypto.ECDSA_P256)

		assert.NoError(t, err)
		assert.Equal(t, dkey.PublicKey.String(), "0xd479b3cdc9edbddb195cb12b35161ade826b032a64bdd4062cc87fb3ba7e71c9cf646ff23990bb4532ca45c445c7e908cef278b2c4615360039a6660a366a95f")
		assert.Equal(t, dkey.SigAlgo.String(), "ECDSA_P256")
	})

	t.Run("Decode PEM Key Invalid", func(t *testing.T) {
		t.Parallel()

		_, srv, _ := setup()
		_, err := srv.DecodePEMKey("nope", crypto.ECDSA_P256)
		assert.Equal(t, err.Error(), "crypto: failed to parse PEM string, not all bytes in PEM key were decoded: 6e6f7065")
	})
}
