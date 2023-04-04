package util

import (
	"testing"

	"github.com/onflow/flow-go-sdk"
	"github.com/onflow/flow-go-sdk/crypto"
	"github.com/stretchr/testify/require"

	"github.com/onflow/flow-cli/pkg/flowkit"
	"github.com/onflow/flow-cli/pkg/flowkit/mocks"
	"github.com/onflow/flow-cli/pkg/flowkit/output"
	"github.com/onflow/flow-cli/pkg/flowkit/tests"
)

var NoLogger = output.NewStdoutLogger(output.NoneLog)
var TestID = flow.HexToID("24993fc99f81641c45c0afa307e683b4f08d407d90041aa9439f487acb33d633")

// TestMocks creates mock flowkit services, an empty state and a mock reader writer
func TestMocks(t *testing.T) (*mocks.MockServices, *flowkit.State, flowkit.ReaderWriter) {
	services := mocks.DefaultMockServices()
	rw, _ := tests.ReaderWriter()
	state, err := flowkit.Init(rw, crypto.ECDSA_P256, crypto.SHA3_256)
	require.NoError(t, err)

	return services, state, rw
}
