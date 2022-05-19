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

package languageserver

import (
	"log"

	"github.com/onflow/flow-cli/pkg/flowkit/util"

	"github.com/onflow/cadence/languageserver"
	"github.com/psiemens/sconfig"
	"github.com/spf13/cobra"
)

type Config struct {
	EnableFlowClient bool `default:"true" flag:"enable-flow-client" info:"Enable Flow client functionality"`
}

var conf Config

var Cmd = &cobra.Command{
	Use:   "language-server",
	Short: "Start the Cadence language server",
	Run: func(cmd *cobra.Command, args []string) {
		languageserver.RunWithStdio(conf.EnableFlowClient)
	},
}

func init() {
	initConfig()
}

func initConfig() {
	err := sconfig.New(&conf).
		FromEnvironment(util.EnvPrefix).
		BindFlags(Cmd.PersistentFlags()).
		Parse()
	if err != nil {
		log.Fatal(err)
	}
}
