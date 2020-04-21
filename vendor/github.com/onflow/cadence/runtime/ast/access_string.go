/*
 * Cadence - The resource-oriented smart contract programming language
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

// Code generated by "stringer -type=Access"; DO NOT EDIT.

package ast

import "strconv"

func _() {
	// An "invalid array index" compiler error signifies that the constant values have changed.
	// Re-run the stringer command to generate them again.
	var x [1]struct{}
	_ = x[AccessNotSpecified-0]
	_ = x[AccessPrivate-1]
	_ = x[AccessContract-2]
	_ = x[AccessAccount-3]
	_ = x[AccessPublic-4]
	_ = x[AccessPublicSettable-5]
}

const _Access_name = "AccessNotSpecifiedAccessPrivateAccessContractAccessAccountAccessPublicAccessPublicSettable"

var _Access_index = [...]uint8{0, 18, 31, 45, 58, 70, 90}

func (i Access) String() string {
	if i < 0 || i >= Access(len(_Access_index)-1) {
		return "Access(" + strconv.FormatInt(int64(i), 10) + ")"
	}
	return _Access_name[_Access_index[i]:_Access_index[i+1]]
}
