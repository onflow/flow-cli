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

package super

import (
	"context"
	"encoding/hex"
	"fmt"
	"os"

	"github.com/onflow/flixkit-go"
	"github.com/onflow/flixkit-go/bindings"

	"github.com/onflow/flow-cli/flowkit"
	"github.com/onflow/flow-cli/flowkit/output"
	"github.com/onflow/flow-cli/internal/command"
	"github.com/onflow/flow-cli/internal/scripts"
	"github.com/onflow/flow-cli/internal/transactions"

	"github.com/spf13/cobra"
)

type flixFlags struct {
	ArgsJSON    string   `default:"" flag:"args-json" info:"arguments in JSON-Cadence format"`
	BlockID     string   `default:"" flag:"block-id" info:"block ID to execute the script at"`
	BlockHeight uint64   `default:"" flag:"block-height" info:"block height to execute the script at"`
	Signer      string   `default:"" flag:"signer" info:"Account name from configuration used to sign the transaction as proposer, payer and suthorizer"`
	Proposer    string   `default:"" flag:"proposer" info:"Account name from configuration used as proposer"`
	Payer       string   `default:"" flag:"payer" info:"Account name from configuration used as payer"`
	Authorizers []string `default:"" flag:"authorizer" info:"Name of a single or multiple comma-separated accounts used as authorizers from configuration"`
	Include     []string `default:"" flag:"include" info:"Fields to include in the output"`
	Exclude     []string `default:"" flag:"exclude" info:"Fields to exclude from the output (events)"`
	GasLimit    uint64   `default:"1000" flag:"gas-limit" info:"transaction gas limit"`
}

type flixResult struct {
	flixQuery string
	result    string
}

var flags = flixFlags{}
var FlixCmd = &cobra.Command{
	Use:              "flix",
	Short:            "execute, bindings",
	TraverseChildren: true,
	GroupID:          "tools",
}

var executeCommand = &command.Command{
	Cmd: &cobra.Command{
		Use:     "execute <id | name | path>",
		Short:   "execute FLIX template with a given id, name, or local filename",
		Example: "flow flix execute transfer-flow 1 0x123456789",
		Args:    cobra.MinimumNArgs(1),
	},
	Flags: &flags,
	RunS:  executeCmd,
}

var bindingCommand = &command.Command{
	Cmd: &cobra.Command{
		Use:     "bindings <id | name | path>",
		Short:   "generate binding file for FLIX template fcl-js is default",
		Example: "flow flix bindings multiply.template.json",
		Args:    cobra.MinimumNArgs(1),
	},
	Flags: &flags,
	RunS:  bindingsCmd,
}

var generateCommand = &command.Command{
	Cmd: &cobra.Command{
		Use:     "generate cadence.cdc",
		Short:   "generate FLIX json template given local filename",
		Example: "flow flix generate multiply.cdc",
		Args:    cobra.MinimumNArgs(1),
	},
	Flags: &flags,
	RunS:  generateCmd,
}

func init() {
	executeCommand.AddToParent(FlixCmd)
	bindingCommand.AddToParent(FlixCmd)
	generateCommand.AddToParent(FlixCmd)
}

type flixQueryTypes string

const (
	flixName flixQueryTypes = "name"
	flixPath flixQueryTypes = "path"
	flixId   flixQueryTypes = "id"
)

func isHex(str string) bool {
	if len(str) != 64 {
		return false
	}
	_, err := hex.DecodeString(str)
	return err == nil
}

func isPath(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func getType(s string) flixQueryTypes {
	switch {
	case isPath(s):
		return flixPath
	case isHex(s):
		return flixId
	default:
		return flixName
	}
}

func executeCmd(
	args []string,
	_ command.GlobalFlags,
	logger output.Logger,
	flow flowkit.Services,
	state *flowkit.State,
) (result command.Result, err error) {
	flixService := flixkit.NewFlixService(&flixkit.Config{})
	flixQuery := args[0]
	template, err := getTemplate(state, flixService, flixQuery)
	if err != nil {
		return nil, err
	}

	cadenceWithImportsReplaced, err := template.GetAndReplaceCadenceImports(flow.Network().Name)
	if err != nil {
		logger.Error("could not replace imports")
		return nil, err
	}

	if template.IsScript() {
		scriptsFlags := scripts.Flags{
			ArgsJSON:    flags.ArgsJSON,
			BlockID:     flags.BlockID,
			BlockHeight: flags.BlockHeight,
		}
		return scripts.SendScript([]byte(cadenceWithImportsReplaced), args[1:], "", flow, scriptsFlags)
	}

	transactionFlags := transactions.Flags{
		ArgsJSON:    flags.ArgsJSON,
		Signer:      flags.Signer,
		Proposer:    flags.Proposer,
		Payer:       flags.Payer,
		Authorizers: flags.Authorizers,
		Include:     flags.Include,
		Exclude:     flags.Exclude,
		GasLimit:    flags.GasLimit,
	}
	return transactions.SendTransaction([]byte(cadenceWithImportsReplaced), args[1:], "", flow, state, transactionFlags)
}

func bindingsCmd(
	args []string,
	_ command.GlobalFlags,
	logger output.Logger,
	flow flowkit.Services,
	state *flowkit.State,
) (result command.Result, err error) {
	flixService := flixkit.NewFlixService(&flixkit.Config{})
	flixQuery := args[0]

	template, err := getTemplate(state, flixService, flixQuery)
	if err != nil {
		return nil, err
	}

	isLocal := false
	if getType(flixQuery) == flixPath {
		isLocal = true
	}

	fclJsGen := bindings.NewFclJSGenerator()

	out, err := fclJsGen.Generate(template, flixQuery, isLocal)

	return &flixResult{
		flixQuery: flixQuery,
		result:    out,
	}, err
}

func generateCmd(
	args []string,
	_ command.GlobalFlags,
	logger output.Logger,
	flow flowkit.Services,
	state *flowkit.State,
) (result command.Result, err error) {
	cadenceFile := args[0]

	if cadenceFile == "" {
		return nil, fmt.Errorf("no cadence code found")
	}

	code, err := state.ReadFile(cadenceFile)
	if err != nil {
		return nil, fmt.Errorf("could not read cadence file %s: %w", cadenceFile, err)
	}

	flixGen := generator.Generator1_0_0()
	flix, err := flixGen.Generate(code)
	if err != nil {
		return nil, fmt.Errorf("could not generate flix %w", err)
	}

	return &flixResult{
		flixQuery: cadenceFile,
		result:    flix,
	}, err
}

func (fr *flixResult) JSON() any {
	result := make(map[string]any)
	result["flixQuery"] = fr.flixQuery
	result["result"] = fr.result
	return result
}

func (fr *flixResult) String() string {
	return fr.result
}

func (fr *flixResult) Oneliner() string {
	return fr.result
}

func getTemplate(state *flowkit.State, flixService flixkit.FlixService, flixQuery string) (*flixkit.FlowInteractionTemplate, error) {
	var template *flixkit.FlowInteractionTemplate
	var err error
	ctx := context.Background()
	switch getType(flixQuery) {
	case flixId:
		template, err = flixService.GetFlixByID(ctx, flixQuery)
		if err != nil {
			return nil, fmt.Errorf("could not find flix with id %s: %w", flixQuery, err)
		}

	case flixName:
		template, err = flixService.GetFlix(ctx, flixQuery)
		if err != nil {
			return nil, fmt.Errorf("could not find flix with name %s: %w", flixQuery, err)
		}

	case flixPath:
		file, err := state.ReadFile(flixQuery)
		if err != nil {
			return nil, fmt.Errorf("could not read flix file %s: %w", flixQuery, err)
		}
		template, err = flixkit.ParseFlix(string(file))
		if err != nil {
			return nil, fmt.Errorf("could not parse flix from file %s: %w", flixQuery, err)
		}

	default:
		return nil, fmt.Errorf("invalid flix query type: %s", flixQuery)
	}
	return template, nil
}
