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

import (
	"fmt"

	"github.com/onflow/cadence"
)

func NewStakingInfoFromValue(value cadence.Value) ([]map[string]interface{}, error) {
	stakingInfo := make([]map[string]interface{}, 0)
	arrayValue, ok := value.(cadence.Array)
	if !ok {
		return stakingInfo, fmt.Errorf("staking info must be a cadence array")
	}

	if len(arrayValue.Values) == 0 {
		return stakingInfo, nil
	}

	for _, v := range arrayValue.Values {
		vs, ok := v.(cadence.Struct)
		if !ok {
			return stakingInfo, fmt.Errorf("staking info must be a cadence array of structs")
		}

		keys := make([]string, 0)
		values := make(map[string]interface{})
		for _, field := range vs.StructType.Fields {
			keys = append(keys, field.Identifier)
		}
		for j, value := range vs.Fields {
			values[keys[j]] = value
		}
		stakingInfo = append(stakingInfo, values)
	}

	return stakingInfo, nil
}
