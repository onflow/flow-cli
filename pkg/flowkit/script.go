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

package flowkit

import "github.com/onflow/cadence"

// Script includes Cadence code and optional arguments and filename.
//
// Filename is only required to be passed if you want to resolve imports.
type Script struct {
	code     []byte
	Args     []cadence.Value
	location string
}

func NewScript(code []byte, args []cadence.Value, location string) *Script {
	return &Script{
		code:     code,
		Args:     args,
		location: location,
	}
}

func (s *Script) Code() []byte {
	return s.code
}

func (s *Script) SetCode(code []byte) {
	s.code = code
}

func (s *Script) Location() string {
	return s.location
}
