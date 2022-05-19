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
	InitCommand.AddToParent(Cmd)
	Cmd.AddCommand(AddCmd)
	Cmd.AddCommand(RemoveCmd)
}

type Result struct {
	result string
}

func (r *Result) JSON() interface{} {
	return nil
}

func (r *Result) String() string {
	if r.result != "" {
		return r.result
	}

	return ""
}

func (r *Result) Oneliner() string {
	return ""
}
