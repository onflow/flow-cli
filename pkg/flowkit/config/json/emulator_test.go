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
package json

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_ConfigEmulatorSimple(t *testing.T) {
	b := []byte(`{
		 "default": {
				"port": 9000,
				"serviceAccount": "emulator-account"
		 }
	 }`)

	var jsonEmulators jsonEmulators
	err := json.Unmarshal(b, &jsonEmulators)
	assert.NoError(t, err)

	emulators, err := jsonEmulators.transformToConfig()
	assert.NoError(t, err)

	assert.Equal(t, emulators[0].Name, "default")
	assert.Equal(t, emulators[0].Port, 9000)
}
