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

package config

import (
	"github.com/spf13/cobra"
)

var Cmd = &cobra.Command{
	Use:              "config",
	Short:            "Utilities to manage configuration",
	TraverseChildren: true,
}

func init() {
	Cmd.AddCommand(addCmd)
	Cmd.AddCommand(removeCmd)
}

type result struct {
	result string
}

func (r *result) JSON() any {
	return nil
}

func (r *result) String() string {
	if r.result != "" {
		return r.result
	}

	return ""
}

func (r *result) Oneliner() string {
	return ""
}
