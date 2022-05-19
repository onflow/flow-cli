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
	"fmt"

	"github.com/onflow/flow-cli/pkg/flowkit/config"
)

type jsonEmulators map[string]jsonEmulator

// transformToConfig transforms json structures to config structure.
func (j jsonEmulators) transformToConfig() (config.Emulators, error) {
	emulators := make(config.Emulators, 0)

	for name, e := range j {
		if e.Port < 0 || e.Port > 65535 {
			return nil, fmt.Errorf("invalid port value")
		}

		emulator := config.Emulator{
			Name:           name,
			Port:           e.Port,
			ServiceAccount: e.ServiceAccount,
		}

		emulators = append(emulators, emulator)
	}

	return emulators, nil
}

// transformToJSON transforms config structure to json structures for saving.
func transformEmulatorsToJSON(emulators config.Emulators) jsonEmulators {
	jsonEmulators := jsonEmulators{}

	for _, e := range emulators {
		jsonEmulators[e.Name] = jsonEmulator{
			Port:           e.Port,
			ServiceAccount: e.ServiceAccount,
		}
	}

	return jsonEmulators
}

type jsonEmulator struct {
	Port           int    `json:"port"`
	ServiceAccount string `json:"serviceAccount"`
}
