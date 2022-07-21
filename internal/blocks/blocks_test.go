package blocks

import (
	"testing"

	"github.com/onflow/flow-cli/internal/command"
	"github.com/onflow/flow-cli/pkg/flowkit/tests"
	"github.com/stretchr/testify/assert"

	"github.com/onflow/flow-cli/pkg/flowkit"
	"github.com/onflow/flow-cli/pkg/flowkit/output"
	"github.com/onflow/flow-cli/pkg/flowkit/services"

	"github.com/onflow/flow-go-sdk/crypto"
)

func setup() (*flowkit.State, *services.Services, *tests.TestGateway) {
	readerWriter := tests.ReaderWriter()
	state, err := flowkit.Init(readerWriter, crypto.ECDSA_P256, crypto.SHA3_256)
	if err != nil {
		panic(err)
	}

	gw := tests.DefaultMockGateway()
	s := services.NewServices(gw.Mock, state, output.NewStdoutLogger(output.NoneLog))

	return state, s, gw
}

func Test_blocks(t *testing.T) {
	t.Run("Get Latest Block command", func(t *testing.T) {
		_, s, _ := setup()

		readerWriter := tests.ReaderWriter()
		res, err := get([]string{"latest"}, readerWriter, command.Flags, s)
		if err != nil {
			panic(err)
		}

		expected := "some expected abstracted string value"
		assert.Equal(t, res.String(), expected, "Get command output does not match the expected output")
	})
}
