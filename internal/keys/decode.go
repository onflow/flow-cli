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

package keys

import (
	"github.com/spf13/cobra"

	"github.com/onflow/flow-cli/internal/command"
	"github.com/onflow/flow-cli/pkg/flowcli/services"
)

type flagsDecode struct{}

var decodeFlags = flagsDecode{}

var DecodeCommand = &command.Command{
	Cmd: &cobra.Command{
		Use:     "decode <public key>",
		Short:   "Decode a public account key hex string",
		Args:    cobra.ExactArgs(1),
		Example: "flow keys decode 4a22246...31bce1e71a7b6d11",
	},
	Flags: &decodeFlags,
	Run: func(
		cmd *cobra.Command,
		args []string,
		globalFlags command.GlobalFlags,
		services *services.Services,
	) (command.Result, error) {
		accountKey, err := services.Keys.Decode(
			args[0], // public key
		)
		if err != nil {
			return nil, err
		}

		pubKey := accountKey.PublicKey
		return &KeyResult{publicKey: pubKey, accountKey: accountKey}, err
	},
}
