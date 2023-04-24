/*
* Flow CLI
*
* Copyright 2019-2020 Dapper Labs, Inc.
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
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEmulatorsDefault(t *testing.T) {
	emulators := Emulators{{
		Name: "emulator-1",
		Port: 1234,
	}, {
		Name:           "default",
		Port:           3569,
		ServiceAccount: "emulator-account",
	}}

	expected := &Emulator{
		Name:           "default",
		ServiceAccount: "emulator-account",
		Port:           3569,
	}

	assert.Equal(t, expected, emulators.Default())
}

func TestEmulatorsAddOrUpdate(t *testing.T) {
	emulator1 := Emulator{
		Name: "emulator-1",
		Port: 1234,
	}
	emulator2 := Emulator{
		Name:           "emulator-account",
		Port:           3569,
		ServiceAccount: "emulator-account",
	}

	emulators := Emulators{emulator1}

	emulators.AddOrUpdate(emulator1.Name, emulator2)

	assert.Len(t, emulators, 1)
	assert.Equal(t, emulator2, emulators[0])

	emulators.AddOrUpdate("emulator-2", Emulator{
		Name: "emulator-2",
		Port: 2345,
	})

	assert.Len(t, emulators, 2)
	assert.Contains(t, emulators, emulator2)
	assert.Contains(t, emulators, Emulator{
		Name: "emulator-2",
		Port: 2345,
	})
}
