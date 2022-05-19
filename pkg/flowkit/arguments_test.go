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

package flowkit_test

import (
	"fmt"
	"testing"

	"github.com/onflow/cadence"
	"github.com/stretchr/testify/assert"

	"github.com/onflow/flow-cli/pkg/flowkit"
)

func TestArguments(t *testing.T) {
	var sampleValues []cadence.Value = []cadence.Value{
		cadence.NewAddress([8]byte{0, 0, 0, 0, 0, 0, 0, 1}),
		cadence.NewBool(true),
		cadence.NewInt(-42),
		cadence.NewInt128(-424242),
		cadence.NewInt32(-42),
		cadence.NewInt64(-42),
		cadence.NewInt8(-42),
		cadence.NewUInt128(424242),
		cadence.NewUInt32(42),
		cadence.NewUInt64(42),
		cadence.NewUInt8(42),
		cadence.String("42"),
	}

	for _, sample := range sampleValues {

		var sampleType string = sample.Type().ID()

		args, err := flowkit.ParseArgumentsWithoutType(
			"",
			[]byte(fmt.Sprintf(`
			pub fun main(test: %s): Void {
			}`, sampleType)),

			[]string{sample.String()},
		)
		assert.NoError(t, err)
		assert.Equal(t, 1, len(args))
		assert.Equal(t, []cadence.Value{sample}, args)
	}
}
