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

package scripts

import (
	"context"
	"fmt"
	"strings"

	"github.com/onflow/flixkit-go"

	"github.com/onflow/cadence"
	flowsdk "github.com/onflow/flow-go-sdk"
	"github.com/spf13/cobra"

	"github.com/onflow/flow-cli/flowkit"
	"github.com/onflow/flow-cli/flowkit/arguments"
	"github.com/onflow/flow-cli/flowkit/output"
	"github.com/onflow/flow-cli/internal/command"
)

type flagsScripts struct {
	ArgsJSON    string `default:"" flag:"args-json" info:"arguments in JSON-Cadence format"`
	BlockID     string `default:"" flag:"block-id" info:"block ID to execute the script at"`
	BlockHeight uint64 `default:"" flag:"block-height" info:"block height to execute the script at"`
}

var scriptFlags = flagsScripts{}

var executeCommand = &command.Command{
	Cmd: &cobra.Command{
		Use:     "execute <filename> [<argument> <argument> ...]",
		Short:   "Execute a script",
		Example: `flow scripts execute script.cdc "Meow" "Woof"`,
		Args:    cobra.MinimumNArgs(1),
	},
	Flags: &scriptFlags,
	Run:   execute,
}

func execute(
	args []string,
	_ command.GlobalFlags,
	_ output.Logger,
	readerWriter flowkit.ReaderWriter,
	flow flowkit.Services,
) (command.Result, error) {
	filenameOrAction := args[0]

	if strings.HasPrefix(filenameOrAction, "flix") {
		return executeFlixScript(args, filenameOrAction, readerWriter, flow)
	}

	return executeLocalScript(args, filenameOrAction, readerWriter, flow)
}

func sendScript(code []byte, argsArr []string, location string, flow flowkit.Services) (command.Result, error) {
	var cadenceArgs []cadence.Value
	var err error
	if scriptFlags.ArgsJSON != "" {
		cadenceArgs, err = arguments.ParseJSON(scriptFlags.ArgsJSON)
	} else {
		cadenceArgs, err = arguments.ParseWithoutType(argsArr, code, location)
	}

	if err != nil {
		return nil, fmt.Errorf("error parsing script arguments: %w", err)
	}

	query := flowkit.ScriptQuery{}
	if scriptFlags.BlockHeight != 0 {
		query.Height = scriptFlags.BlockHeight
	} else if scriptFlags.BlockID != "" {
		query.ID = flowsdk.HexToID(scriptFlags.BlockID)
	} else {
		query.Latest = true
	}

	value, err := flow.ExecuteScript(
		context.Background(),
		flowkit.Script{
			Code:     code,
			Args:     cadenceArgs,
			Location: location,
		},
		query,
	)
	if err != nil {
		return nil, err
	}

	return &scriptResult{value}, nil
}

func executeLocalScript(args []string, filename string, readerWriter flowkit.ReaderWriter, flow flowkit.Services) (command.Result, error) {
	code, err := readerWriter.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("error loading script file: %w", err)
	}

	return sendScript(code, args[1:], filename, flow)
}

func getFlixCadence(args []string, action string, readerWriter flowkit.ReaderWriter) (*flixkit.FlowInteractionTemplate, []string, error) {
	commandParts := strings.Split(action, ":")

	if len(commandParts) != 3 {
		return nil, nil, fmt.Errorf("invalid flix command")
	}

	flixFindMethod := commandParts[1]
	flixIdentifier := commandParts[2]

	var flixService = flixkit.NewFlixService(&flixkit.Config{})
	var parsedFlixTemplate *flixkit.FlowInteractionTemplate
	var argsArr []string

	switch flixFindMethod {
	case "name":
		argsArr = args[1:]
		ctx := context.Background()
		flixTemplate, err := flixService.GetFlix(ctx, flixIdentifier)
		if err != nil {
			return nil, nil, fmt.Errorf("could not find flix template")
		}
		parsedFlixTemplate = flixTemplate

	case "id":
		argsArr = args[1:]
		ctx := context.Background()
		flixTemplate, err := flixService.GetFlixByID(ctx, flixIdentifier)
		if err != nil {
			return nil, nil, fmt.Errorf("could not find flix template")
		}
		parsedFlixTemplate = flixTemplate

	case "local":
		if flixIdentifier == "path" {
			filePath := args[1]
			argsArr = args[2:]

			flixTemplate, err := readerWriter.ReadFile(filePath)
			if err != nil {
				return nil, nil, fmt.Errorf("error loading script file: %w", err)
			}

			parsedTemplate, err := flixkit.ParseFlix(string(flixTemplate))
			if err != nil {
				return nil, nil, fmt.Errorf("error parsing script file: %w", err)
			}

			parsedFlixTemplate = parsedTemplate
		} else {
			return nil, nil, fmt.Errorf("invalid flix command")
		}

	default:
		return nil, nil, fmt.Errorf("invalid flix command")
	}

	return parsedFlixTemplate, argsArr, nil
}

func executeFlixScript(args []string, action string, readerWriter flowkit.ReaderWriter, flow flowkit.Services) (command.Result, error) {
	flix, updatedArgs, err := getFlixCadence(args, action, readerWriter)

	if flix.IsTransaction() {
		return nil, fmt.Errorf("invalid command for a transaction")
	}

	cadenceWithImportsReplaced, err := flix.GetAndReplaceCadenceImports("testnet")
	if err != nil {
		return nil, fmt.Errorf("could not replace imports")
	}

	return sendScript([]byte(cadenceWithImportsReplaced), updatedArgs, "", flow)
}
