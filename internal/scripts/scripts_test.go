/*
 * Flow CLI
 *
 * Copyright 2019 Dapper Labs, Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *   http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package scripts

import (
	"fmt"
	"github.com/onflow/cadence"
	"github.com/onflow/flow-cli/internal/command"
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
			assert.Equal(t, tests.ScriptArgString.Filename, script.Location)
		}).Return(cadence.NewInt(1), nil)

		result, err := execute(inArgs, command.GlobalFlags{}, util.NoLogger, rw, srv.Mock)
		assert.NotNil(t, result)
		assert.NoError(t, err)
	})

	t.Run("Fail non-existing file", func(t *testing.T) {
		inArgs := []string{"non-existing"}
		result, err := execute(inArgs, command.GlobalFlags{}, util.NoLogger, rw, srv.Mock)
		assert.Nil(t, result)
		assert.EqualError(t, err, "error loading script file: open non-existing: file does not exist")
	})

	t.Run("Fail parsing invalid JSON args", func(t *testing.T) {
		inArgs := []string{tests.TestScriptSimple.Filename}
		scriptFlags.ArgsJSON = "invalid"

		result, err := execute(inArgs, command.GlobalFlags{}, util.NoLogger, rw, srv.Mock)
		assert.Nil(t, result)
		assert.EqualError(t, err, "error parsing script arguments: invalid character 'i' looking for beginning of value")
	})

}
