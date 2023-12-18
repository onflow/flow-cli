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

package evm

import (
	"github.com/spf13/cobra"
)

var Cmd = &cobra.Command{
	Use:              "evm",
	Short:            "Interact with Flow EVM",
	TraverseChildren: true,
}

func init() {
	deployCommand.AddToParent(Cmd)
	createCommand.AddToParent(Cmd)
	getCommand.AddToParent(Cmd)
	runCommand.AddToParent(Cmd)
	rpcCommand.AddToParent(Cmd)
	fundCommand.AddToParent(Cmd)
}

type evmResult struct {
}

func (k *evmResult) JSON() any {
	return nil
}

func (k *evmResult) String() string {
	return ""
}

func (k *evmResult) Oneliner() string {
	return ""
}
