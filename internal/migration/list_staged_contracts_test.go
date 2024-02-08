package migration

import (
	"testing"

	"github.com/onflow/cadence"
	flowsdk "github.com/onflow/flow-go-sdk"
	"github.com/onflow/flowkit"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/onflow/flow-cli/internal/command"
	"github.com/onflow/flow-cli/internal/util"
)

func Test_ListStagedContracts(t *testing.T) {
	srv, state, _ := util.TestMocks(t)

	t.Run("Success", func(t *testing.T) {

		srv.ExecuteScript.Run(func(args mock.Arguments) {
			script := args.Get(1).(flowkit.Script)

			actualContractAddressArg := script.Args[0]

			contractAddr := cadence.NewAddress(flowsdk.HexToAddress("0xSomeAddress"))
			assert.Equal(t, contractAddr, actualContractAddressArg)
		}).Return(cadence.NewMeteredBool(nil, true), nil)

		result, err := listStagedContracts(
			[]string{"0xSomeAddress"},
			command.GlobalFlags{
				Network: "testnet",
			},
			util.NoLogger,
			srv.Mock,
			state,
		)
		assert.NoError(t, err)
		assert.NotNil(t, result)
	})
}
