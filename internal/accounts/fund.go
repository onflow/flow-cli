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

package accounts

import (
	"fmt"
	"time"

	flowsdk "github.com/onflow/flow-go-sdk"

	"github.com/pkg/browser"
	"github.com/spf13/cobra"

	"github.com/onflow/flow-cli/internal/command"
	"github.com/onflow/flowkit"
	"github.com/onflow/flowkit/output"
)

type flagsFund struct {
	Include []string `default:"" flag:"include" info:"Fields to include in the output. Valid values: contracts."`
}

var fundFlags = flagsFund{}

var fundCommand = &command.Command{
	Cmd: &cobra.Command{
		Use:     "fund <address>",
		Short:   "Funds an account by address through the Testnet Faucet",
		Example: "flow accounts fund 8e94eaa81771313a",
		Args:    cobra.ExactArgs(1),
	},
	Flags: &fundFlags,
	Run:   fund,
}

func fund(
	args []string,
	_ command.GlobalFlags,
	logger output.Logger,
	_ flowkit.ReaderWriter,
	flow flowkit.Services,
) (command.Result, error) {
	address := flowsdk.HexToAddress(args[0])
	if !address.IsValid(flowsdk.Testnet) {
		return nil, fmt.Errorf("unsupported address %s, faucet can only work for valid Testnet addresses", address.String())
	}

	logger.Info(
		fmt.Sprintf(
			"Opening the Testnet faucet to fund 0x%s on your native browser."+
				"\n\nIf there is an issue, please use this link instead: %s",
			address.String(),
			testnetFaucetURL(address),
		))
	// wait for the user to read the message
	time.Sleep(5 * time.Second)

	if err := browser.OpenURL(testnetFaucetURL(address)); err != nil {
		return nil, err
	}

	return nil, nil
}
