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

	"github.com/onflow/flow-go-sdk"
	"github.com/onflow/flow-go-sdk/crypto"
	"github.com/stretchr/testify/assert"

	"github.com/onflow/flow-cli/pkg/flowcli/output"
	"github.com/onflow/flow-cli/pkg/flowcli/project"
	"github.com/onflow/flow-cli/tests"
)

func TestCollections(t *testing.T) {
	mock := &tests.MockGateway{}

	proj, err := project.Init(crypto.ECDSA_P256, crypto.SHA3_256)
	assert.NoError(t, err)

	collections := NewCollections(mock, proj, output.NewStdoutLogger(output.InfoLog))

	t.Run("Get Collection", func(t *testing.T) {
		called := false
		mock.GetCollectionMock = func(id flow.Identifier) (*flow.Collection, error) {
			called = true
			return tests.NewCollection(), nil
		}

		_, err := collections.Get("a310685082f0b09f2a148b2e8905f08ea458ed873596b53b200699e8e1f6536f")

		assert.NoError(t, err)
		assert.True(t, called)
	})
}
