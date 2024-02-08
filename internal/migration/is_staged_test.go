package migration

import (
	"testing"

	"github.com/onflow/cadence"
	"github.com/onflow/flow-cli/internal/command"
	"github.com/onflow/flow-cli/internal/util"
	flowsdk "github.com/onflow/flow-go-sdk"
	"github.com/onflow/flowkit"
	"github.com/onflow/flowkit/tests"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func Test_IsStaged(t *testing.T) {
	srv, state, _ := util.TestMocks(t)

	testContract := tests.ContractSimple

	t.Run("Success", func(t *testing.T) {

		srv.ExecuteScript.Run(func(args mock.Arguments) {
		script := args.Get(1).(flowkit.Script)

		actualContractAddressArg, actualContractNameArg := script.Args[0], script.Args[1]

		contractName, _ := cadence.NewString(testContract.Name)
		contractAddr := cadence.NewAddress(flowsdk.HexToAddress("0xSomeAddress"))
		assert.Equal(t, contractName, actualContractNameArg)
		assert.Equal(t, contractAddr, actualContractAddressArg)
		}).Return(cadence.NewMeteredOptional(nil, nil), nil)

	result, err := isStaged(
		[]string{testContract.Name, "0xSomeAddress"}, 
		command.GlobalFlags{
			Network: "testnet",
		},
		 util.NoLogger, 
		 srv.Mock, 
		 state,
		)
	assert.NoError(t, err)
	// TODO: fix this
	assert.Nil(t, result)
})
}
