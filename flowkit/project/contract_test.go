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

package project

import (
	"testing"

	"github.com/onflow/cadence"
	"github.com/onflow/flow-go-sdk"
	"github.com/stretchr/testify/assert"
)

func TestContractFunctions(t *testing.T) {
	name := "contract1"
	location := "/path/to/contract1"
	code := []byte("code1")
	accountAddress := flow.HexToAddress("0x123456789")
	accountName := "account1"
	str, _ := cadence.NewString("arg1")
	args := []cadence.Value{str}

	contract := NewContract(name, location, code, accountAddress, accountName, args)

	assert.Equal(t, name, contract.Name, "contract.Name should be equal to expected value")
	assert.Equal(t, location, contract.location, "contract.location should be equal to expected value")
	assert.Equal(t, code, contract.code, "contract.code should be equal to expected value")
	assert.Equal(t, accountAddress, contract.AccountAddress, "contract.AccountAddress should be equal to expected value")
	assert.Equal(t, accountName, contract.AccountName, "contract.AccountName should be equal to expected value")
	assert.Equal(t, args, contract.Args, "contract.Args should be equal to expected value")

	assert.Equal(t, code, contract.Code(), "contract.Code() should return expected value")

	newCode := []byte("newcode")
	contract.SetCode(newCode)
	assert.Equal(t, newCode, contract.code, "contract.code should be equal to expected value")

	assert.Equal(t, location, contract.Location(), "contract.Location() should return expected value")
}
