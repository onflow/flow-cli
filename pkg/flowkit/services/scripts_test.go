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

	"github.com/onflow/flow-cli/pkg/flowkit"

	"github.com/onflow/cadence"
	"github.com/onflow/flow-go-sdk/crypto"
	"github.com/stretchr/testify/assert"

	"github.com/onflow/flow-cli/pkg/flowkit/output"
	"github.com/onflow/flow-cli/tests"
)

func TestScripts(t *testing.T) {
	mock := &tests.MockGateway{}

	proj, err := flowkit.Init(crypto.ECDSA_P256, crypto.SHA3_256)
	assert.NoError(t, err)

	scripts := NewScripts(mock, proj, output.NewStdoutLogger(output.InfoLog))

	t.Run("Execute Script", func(t *testing.T) {
		mock.ExecuteScriptMock = func(script []byte, arguments []cadence.Value) (cadence.Value, error) {
			assert.Equal(t, len(string(script)), 69)
			assert.Equal(t, arguments[0].String(), "\"Foo\"")
			return arguments[0], nil
		}

		_, err := scripts.Execute("../../../tests/script.cdc", []string{"String:Foo"}, "", "")

		assert.NoError(t, err)
	})
}
