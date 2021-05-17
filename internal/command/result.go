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

package command

import (
	"github.com/onflow/flow-cli/pkg/flowcli/util"
)

type Result interface {
	String() string
	Oneliner() string
	JSON() interface{}
}

// FieldIncluded checks if field should be included in the result by checking if included flags contains field
func FieldIncluded(included []string, field string) bool {
	return util.ContainsStringInsensitive(included, field)
}

// FieldExcluded checks if field should be excluded from the result by checking the flags
func FieldExcluded(excluded []string, field string) bool {
	return util.ContainsStringInsensitive(excluded, field)
}
