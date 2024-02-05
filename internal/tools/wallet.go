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

package tools

import (
	"fmt"
	"strings"

	devWallet "github.com/onflow/fcl-dev-wallet/go/wallet"
	"github.com/spf13/cobra"

	"github.com/onflow/flowkit"
	"github.com/onflow/flowkit/output"

	"github.com/onflow/flow-cli/internal/command"
)

type flagsWallet struct {
	Port uint   `default:"8701" flag:"port" info:"Dev wallet port to listen on"`
	Host string `default:"http://localhost:8888" flag:"emulator-host" info:"Host for access node connection"`
}

var walletFlags = flagsWallet{}

var DevWallet = &command.Command{
	Cmd: &cobra.Command{
		Use:     "dev-wallet",
		Short:   "Run a development wallet",
		Example: "flow dev-wallet",
		Args:    cobra.ExactArgs(0),
		GroupID: "tools",
	},
	Flags: &walletFlags,
	RunS:  wallet,
}

func wallet(
	_ []string,
	_ command.GlobalFlags,
	_ output.Logger,
	_ flowkit.Services,
	state *flowkit.State,
) (command.Result, error) {
	service, err := state.EmulatorServiceAccount()
	if err != nil {
		return nil, err
	}

	privateKey, err := service.Key.PrivateKey()
	if err != nil {
		return nil, err
	}

	conf := devWallet.FlowConfig{
		Address:    fmt.Sprintf("0x%s", service.Address.String()),
		PrivateKey: strings.TrimPrefix((*privateKey).String(), "0x"),
		PublicKey:  strings.TrimPrefix((*privateKey).PublicKey().String(), "0x"),
		AccessNode: walletFlags.Host,
	}

	srv, err := devWallet.NewHTTPServer(walletFlags.Port, &conf)
	if err != nil {
		return nil, err
	}

	fmt.Printf("%s Starting dev wallet server on port %d\n", output.SuccessEmoji(), walletFlags.Port)
	fmt.Printf("%s  Make sure the emulator is running\n", output.WarningEmoji())

	srv.Start()
	return nil, nil
}
