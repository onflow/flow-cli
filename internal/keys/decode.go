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
		Use:       "decode <rlp|pem> <encoded public key>",
		Short:     "Decode an encoded public key",
		Args:      cobra.ExactArgs(2),
		ValidArgs: []string{"rlp", "pem"},
		Example:   "flow keys decode rlp f847b8408...2402038203e8",
	},
	Flags: &decodeFlags,
	Run: func(
		cmd *cobra.Command,
		args []string,
		globalFlags command.GlobalFlags,
		services *services.Services,
	) (command.Result, error) {
		encoded := args[1]
		encoding := args[0]

		accountKey, err := services.Keys.Decode(encoded, encoding)
		if err != nil {
			return nil, err
		}

		pubKey := accountKey.PublicKey
		return &KeyResult{publicKey: pubKey, accountKey: accountKey}, err
	},
}
