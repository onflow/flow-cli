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

package command

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/onflow/flow-go-sdk/access/grpc"
	"github.com/spf13/afero"
	"golang.org/x/exp/maps"

	"github.com/onflow/flowkit/v2/config"
	"github.com/onflow/flowkit/v2/output"
)

// Result interface describes all the formats for the result output.
type Result interface {
	// String will output the result in human readable output.
	String() string
	// Oneliner will output the result in "grep-able" format.
	Oneliner() string
	// JSON will output the result in JSON format
	JSON() any
}

type ResultWithExitCode interface {
	Result
	ExitCode() int
}

// ContainsFlag checks if output flag is present for the provided field.
func ContainsFlag(flags []string, field string) bool {
	for _, n := range flags {
		if strings.ToLower(n) == field {
			return true
		}
	}

	return false
}

// formatResult formats a result for printing.
func formatResult(result Result, filterFlag string, formatFlag string) (string, error) {
	if result == nil {
		return "", fmt.Errorf("missing result")
	}

	if filterFlag != "" {
		value, err := filterResultValue(result, filterFlag)
		if err != nil {
			return "", err
		}

		return fmt.Sprintf("%v", value), nil
	}

	switch strings.ToLower(formatFlag) {
	case FormatJSON:
		jsonRes, _ := json.Marshal(result.JSON())
		return string(jsonRes), nil
	case FormatInline:
		return result.Oneliner(), nil
	default:
		return result.String(), nil
	}
}

// outputResult to selected media.
func outputResult(result string, saveFlag string, formatFlag string, filterFlag string) error {
	if saveFlag != "" {
		af := afero.Afero{
			Fs: afero.NewOsFs(),
		}

		// create directory if doesn't exist
		dir := filepath.Dir(saveFlag)
		err := af.MkdirAll(dir, 0644)
		if err != nil {
			return err
		}

		fmt.Printf("%s result saved to: %s \n", output.SaveEmoji(), saveFlag)
		return af.WriteFile(saveFlag, []byte(result), 0644)
	}

	if formatFlag == FormatInline || filterFlag != "" {
		_, _ = fmt.Fprintf(os.Stdout, "%s", result)
	} else { // default normal output
		_, _ = fmt.Fprintf(os.Stdout, "\n%s\n\n", result)
	}
	return nil
}

// filterResultValue returns a value by its name filtered from other result values.
func filterResultValue(result Result, filter string) (any, error) {
	res, ok := result.JSON().(map[string]any)
	if !ok {
		return "", fmt.Errorf("not possible to filter by the value")
	}

	value := res[filter]
	if value == nil {
		value = res[strings.ToLower(filter)]
	}
	if value == nil {
		return nil, fmt.Errorf("value for filter: '%s' doesn't exists, possible values to filter by: %s", filter, maps.Keys(res))
	}

	return value, nil
}

// handleError handle errors returned from command execution, try to understand why error happens and offer help to the user.
func handleError(description string, err error) {
	if err == nil {
		return
	}

	// TODO(sideninja): refactor this to better handle errors not by string matching
	// handle rpc error
	switch t := err.(type) {
	case *grpc.RPCError:
		_, _ = fmt.Fprintf(os.Stderr, "%s Grpc Error: %s \n", output.ErrorEmoji(), t.GRPCStatus().Err().Error())
	default:
		if errors.Is(err, config.ErrOutdatedFormat) {
			_, _ = fmt.Fprintf(os.Stderr, "%s Config Error: %s \n", output.ErrorEmoji(), err.Error())
			_, _ = fmt.Fprintf(os.Stderr, "%s Please reset configuration using: 'flow init --reset'. Read more about new configuration here: https://github.com/onflow/flow-cli/releases/tag/v0.17.0", output.TryEmoji())
		} else if errors.Is(err, config.ErrDoesNotExist) {
			_, _ = fmt.Fprintf(os.Stderr, "%s Config Error: %s \n", output.ErrorEmoji(), err.Error())
			_, _ = fmt.Fprintf(os.Stderr, "%s Please create configuration using: flow init", output.TryEmoji())
		} else if strings.Contains(err.Error(), "transport:") {
			_, _ = fmt.Fprintf(os.Stderr, "%s %s \n", output.ErrorEmoji(), strings.Split(err.Error(), "transport:")[1])
			_, _ = fmt.Fprintf(os.Stderr, "%s Make sure your emulator is running or connection address is correct.", output.TryEmoji())
		} else if strings.Contains(err.Error(), "NotFound desc =") {
			_, _ = fmt.Fprintf(os.Stderr, "%s Not Found:%s \n", output.ErrorEmoji(), strings.Split(err.Error(), "NotFound desc =")[1])
		} else if strings.Contains(err.Error(), "code = InvalidArgument desc = ") {
			desc := strings.Split(err.Error(), "code = InvalidArgument desc = ")
			_, _ = fmt.Fprintf(os.Stderr, "%s Invalid argument: %s \n", output.ErrorEmoji(), desc[len(desc)-1])
			if strings.Contains(err.Error(), "is invalid for chain") {
				_, _ = fmt.Fprintf(os.Stderr, "%s Check you are connecting to the correct network or account address you use is correct.", output.TryEmoji())
			} else {
				_, _ = fmt.Fprintf(os.Stderr, "%s Check your argument and flags value, you can use --help.", output.TryEmoji())
			}
		} else if strings.Contains(err.Error(), "invalid signature:") {
			_, _ = fmt.Fprintf(os.Stderr, "%s Invalid signature: %s \n", output.ErrorEmoji(), strings.Split(err.Error(), "invalid signature:")[1])
			_, _ = fmt.Fprintf(os.Stderr, "%s Check the signer private key is provided or is in the correct format. If running emulator, make sure it's using the same configuration as this command.", output.TryEmoji())
		} else if strings.Contains(err.Error(), "signature could not be verified using public key with") {
			_, _ = fmt.Fprintf(os.Stderr, "%s %s: %s \n", output.ErrorEmoji(), description, err)
			_, _ = fmt.Fprintf(os.Stderr, "%s If you are running emulator locally make sure that the emulator was started with the same config as used in this command. \nTry restarting the emulator.", output.TryEmoji())
		} else {
			_, _ = fmt.Fprintf(os.Stderr, "%s %s: %s", output.ErrorEmoji(), description, err)
		}
	}

	fmt.Println()
	os.Exit(1)
}
