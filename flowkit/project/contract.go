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
	"github.com/onflow/cadence"
	"github.com/onflow/flow-go-sdk"
)

// Contract is a Cadence contract definition for a project.
type Contract struct {
	Name           string
	location       string
	code           []byte
	AccountAddress flow.Address
	AccountName    string
	Args           []cadence.Value
}

func NewContract(
	name string,
	location string,
	code []byte,
	accountAddress flow.Address,
	accountName string,
	args []cadence.Value,
) *Contract {
	return &Contract{
		Name:           name,
		location:       location,
		code:           code,
		AccountAddress: accountAddress,
		AccountName:    accountName,
		Args:           args,
	}
}

func (c *Contract) Code() []byte {
	return c.code
}

func (c *Contract) SetCode(code []byte) {
	c.code = code
}

func (c *Contract) Location() string {
	return c.location
}

// LocationAliases map contract locations to fixed addresses on Flow network
type LocationAliases map[string]string
