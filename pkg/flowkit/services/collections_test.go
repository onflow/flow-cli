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

package services

import (
	"testing"

	"github.com/onflow/flow-go-sdk"
	"github.com/stretchr/testify/assert"
)

func TestCollections(t *testing.T) {
	t.Parallel()

	t.Run("Get Collection", func(t *testing.T) {
		_, s, gw := setup()
		ID := flow.HexToID("a310685082f0b09f2a148b2e8905f08ea458ed873596b53b200699e8e1f6536f")

		_, err := s.Collections.Get(ID)

		assert.NoError(t, err)
		gw.Mock.AssertCalled(t, "GetCollection", ID)
	})
}
