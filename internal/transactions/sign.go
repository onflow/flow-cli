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

package transactions

import (
	"fmt"

	"github.com/onflow/flow-cli/pkg/flowkit"

	"github.com/spf13/cobra"

	"github.com/onflow/flow-cli/internal/command"
	"github.com/onflow/flow-cli/pkg/flowkit/output"
	"github.com/onflow/flow-cli/pkg/flowkit/services"
)

type flagsSign struct {
	Signer        string   `default:"emulator-account" flag:"signer" info:"name of the account used to sign"`
	Include       []string `default:"" flag:"include" info:"Fields to include in the output. Valid values: signatures, code, payload."`
	FromRemoteUrl string   `default:"" flag:"from-remote-url" info:"server URL where RLP can be fetched, signed RLP will be posted back to remote URL."`
}

var signFlags = flagsSign{}

var SignCommand = &command.Command{
	Cmd: &cobra.Command{
		Use:     "sign [<built transaction filename> | --from-remote-url <url>]",
		Short:   "Sign built transaction",
		Example: "flow transactions sign ./built.rlp --signer alice",
		Args:    cobra.MaximumNArgs(1),
	},
	Flags: &signFlags,
	RunS:  sign,
}

func sign(
	args []string,
	readerWriter flowkit.ReaderWriter,
	globalFlags command.GlobalFlags,
	services *services.Services,
	state *flowkit.State,
) (command.Result, error) {
	var payload []byte
	var err error
	var filenameOrUrl string

	if signFlags.FromRemoteUrl != "" {
		if globalFlags.Yes {
			return nil, fmt.Errorf("--yes is not supported with this flag")
		}
		filenameOrUrl = signFlags.FromRemoteUrl
		payload, err = services.Transactions.GetRLP(filenameOrUrl)
	} else {
		if len(args) == 0 {
			return nil, fmt.Errorf("filename argument is required")
		}
		filenameOrUrl = args[0]
		payload, err = readerWriter.ReadFile(filenameOrUrl)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to read partial transaction from %s: %v", filenameOrUrl, err)
	}

	signer, err := state.Accounts().ByName(signFlags.Signer)
	if err != nil {
		return nil, fmt.Errorf("signer account: [%s] doesn't exists in configuration", signFlags.Signer)
	}

	signed, err := services.Transactions.Sign(signer, payload, globalFlags.Yes)
	if err != nil {
		return nil, err
	}

	if signFlags.FromRemoteUrl != "" {
		tx := signed.FlowTransaction()
		err = services.Transactions.PostRLP(filenameOrUrl, tx)

		if err != nil {
			return nil, err
		}
		fmt.Printf("%s Signed RLP Posted successfully\n", output.SuccessEmoji())
	}

	return &TransactionResult{
		tx:      signed.FlowTransaction(),
		include: signFlags.Include,
	}, nil
}
