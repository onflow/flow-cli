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

package state

import (
	"github.com/spf13/cobra"
)

var Cmd = &cobra.Command{
	Use:              "state",
	Short:            "Utilities to manage configuration",
	TraverseChildren: true,
}

func init() {
}

// StateResult result from state
type StateResult struct {
	result string
}

// JSON convert result to JSON
func (r *StateResult) JSON() interface{} {
	return nil
}

func (r *StateResult) String() string {
	if r.result != "" {
		return r.result
	}

	return ""
}

// Oneliner show result as one liner grep friendly
func (r *StateResult) Oneliner() string {
	return ""
}
