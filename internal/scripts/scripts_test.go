package scripts

import (
	"fmt"
	"github.com/onflow/cadence"
	"github.com/onflow/flow-cli/internal/util"
	"github.com/onflow/flow-cli/pkg/flowkit"
	"github.com/onflow/flow-cli/pkg/flowkit/tests"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"testing"
)

func Test_Execute(t *testing.T) {
	srv, _, rw := util.TestMocks(t)

	t.Run("Success", func(t *testing.T) {
		inArgs := []string{tests.ScriptArgString.Filename, "foo"}

		srv.ExecuteScript.Run(func(args mock.Arguments) {
			script := args.Get(1).(*flowkit.Script)
			assert.Equal(t, fmt.Sprintf("\"%s\"", inArgs[1]), script.Args[0].String())
			assert.Equal(t, tests.ScriptArgString.Filename, script.Location())
		}).Return(cadence.NewInt(1), nil)

		result, err := execute(inArgs, util.NoFlags, util.NoLogger, rw, srv.Mock)
		assert.NotNil(t, result)
		assert.NoError(t, err)
	})

	t.Run("Fail non-existing file", func(t *testing.T) {
		inArgs := []string{"non-existing"}
		result, err := execute(inArgs, util.NoFlags, util.NoLogger, rw, srv.Mock)
		assert.Nil(t, result)
		assert.EqualError(t, err, "error loading script file: open non-existing: file does not exist")
	})

	t.Run("Fail parsing invalid JSON args", func(t *testing.T) {
		inArgs := []string{tests.TestScriptSimple.Filename}
		scriptFlags.ArgsJSON = "invalid"

		result, err := execute(inArgs, util.NoFlags, util.NoLogger, rw, srv.Mock)
		assert.Nil(t, result)
		assert.EqualError(t, err, "error parsing script arguments: invalid character 'i' looking for beginning of value")
	})

}
