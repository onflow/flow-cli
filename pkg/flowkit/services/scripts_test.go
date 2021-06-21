/*
 * Flow CLI
 *
 * Copyright 2019-2021 Dapper Labs, Inc.
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

package services

import (
	"testing"

	"github.com/onflow/cadence"

	"github.com/stretchr/testify/mock"

	"github.com/onflow/flow-cli/pkg/flowkit"
	"github.com/onflow/flow-cli/tests"
	"github.com/stretchr/testify/assert"
)

func TestScripts(t *testing.T) {

	t.Run("Execute Script", func(t *testing.T) {
		_, s, gw := setup()

		gw.ExecuteScript.Run(func(args mock.Arguments) {
			assert.Equal(t, len(string(args.Get(0).([]byte))), 78)
			assert.Equal(t, args.Get(1).([]cadence.Value)[0].String(), "\"Foo\"")
			gw.ExecuteScript.Return(cadence.MustConvertValue(""), nil)
		})

		args, _ := flowkit.ParseArgumentsCommaSplit([]string{"String:Foo"})

		_, err := s.Scripts.Execute(tests.ScriptArgString.Source, args, "", "")

		assert.NoError(t, err)
	})

}
