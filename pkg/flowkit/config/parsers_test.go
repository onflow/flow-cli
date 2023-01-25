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

package config

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_ParseAddress(t *testing.T) {
	tests := []string{
		"0x0000000000000002",
		"0x4c41cf317eec148e", // mainnet
		"0x78b84cd3c394708c", // testnet
		"0xf8d6e0586b0a20c7", // emulator
	}

	for _, test := range tests {
		addr, err := StringToAddress(test)
		require.NoError(t, err)
		assert.Equal(t, strings.TrimPrefix(test, "0x"), addr.Hex())
	}

}
