/*
 * Flow CLI
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

package status

import (
	"log"

	"github.com/psiemens/sconfig"
	"github.com/spf13/cobra"

	"github.com/onflow/flow-cli/flow/cli"
)

type Flags struct {
	Host   string `flag:"host" info:"Flow Access API host address"`
	Sealed bool   `default:"true" flag:"sealed" info:"Wait for a sealed result"`
	Code   bool   `default:"false" flag:"code" info:"Display transaction code"`
}

var flags Flags

var Cmd = &cobra.Command{
	Use:   "status <tx_id>",
	Short: "Get the transaction status",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		project := cli.LoadProject()

		cli.GetTransactionResult(
			project.HostWithOverride(flags.Host), args[0], flags.Sealed, flags.Code,
		)
	},
}

func init() {
	initConfig()
}

func initConfig() {
	err := sconfig.New(&flags).
		FromEnvironment(cli.EnvPrefix).
		BindFlags(Cmd.PersistentFlags()).
		Parse()
	if err != nil {
		log.Fatal(err)
	}
}
