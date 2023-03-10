package accounts

import (
	"github.com/onflow/flow-cli/internal/command"
	"github.com/onflow/flow-cli/pkg/flowkit"
	"github.com/onflow/flow-cli/pkg/flowkit/output"
	"github.com/onflow/flow-cli/pkg/flowkit/tests"
	"github.com/onflow/flow-cli/pkg/flowkit/tests/mocks"
	"github.com/onflow/flow-go-sdk"
	"github.com/onflow/flow-go-sdk/crypto"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"testing"
)

var noFlags = command.GlobalFlags{}
var noLogger = output.NewStdoutLogger(output.NoneLog)
var services = &mocks.Services{}

func Test_AddContract(t *testing.T) {
	args := []string{tests.ContractA.Filename}

	rw, mockFs := tests.ReaderWriter()
	_ = afero.WriteFile(mockFs, tests.ContractA.Filename, tests.ContractA.Source, 0644)

	state, err := flowkit.Init(rw, crypto.ECDSA_P256, crypto.SHA3_256)
	require.NoError(t, err)

	services.
		On("GetAccount", mock.Anything, mock.AnythingOfType("flow.Address")).
		Return(tests.NewAccountWithAddress("0x01"), nil)

	services.On(
		"AddContract",
		mock.Anything,
		mock.AnythingOfType("*flowkit.Account"),
		mock.AnythingOfType("*flowkit.Script"),
		false,
	).Return(flow.EmptyID, false, nil)

	result, err := addContract(args, noFlags, noLogger, services, state)

	require.NoError(t, err)
	assert.Equal(t, "Address: 0x0000000000000001, Balance: 0.00000010, Public Keys: [0x8da60bd98a827c87e21622c5070ae3ee440abf0927d5db33f9652cb1303eb8a04dfe41dea2c9ea64ee83ee8d7c8d068db8386c7bab98694af956e0fdae37184e 0xc8a2a318b9099cc6c872a0ec3dcd9f59d17837e4ffd6cd8a1f913ddfa769559605e1ad6ad603ebb511f5a6c8125f863abc2e9c600216edaa07104a0fe320dba7]", result.Oneliner())

}
