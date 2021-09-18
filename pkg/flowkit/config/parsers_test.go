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

package config_test

import (
	"testing"

	"github.com/onflow/flow-cli/pkg/flowkit/config"
	"github.com/stretchr/testify/assert"
)

func TestStringToDeployments(t *testing.T) {
	testCases := []struct {
		name, network, account string
		contracts              []string
	}{
		{"TestEmulator", "emulator", "emulator-account", []string{"HelloWorld"}},
	}
	for _, c := range testCases {
		t.Run(c.name, func(t *testing.T) {
			d := config.StringToDeployment(c.network, c.account, c.contracts)
			assert.Equal(t, c.account, d.Account)
			assert.Equal(t, c.network, d.Network)
			deploymentContractNames := make([]string, len(d.Contracts))
			for i, contract := range d.Contracts {
				deploymentContractNames[i] = contract.Name
			}
			assert.ElementsMatch(t, c.contracts, deploymentContractNames)
		})
	}

}
