/*
 * Flow CLI
 *
 * Copyright Flow Foundation
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

	"github.com/onflow/flow-cli/common/branding"
)

// Error styling
var (
	errorStyle        = branding.ErrorStyle
	errorMessageStyle = branding.ErrorStyle
	suggestionStyle   = branding.GreenStyle.Copy().Italic(true)
	descriptionStyle  = branding.GrayStyle.Copy().Bold(true)
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
		errorMsg := errorStyle.Render(fmt.Sprintf("%s Grpc Error:", output.ErrorEmoji()))
		detailMsg := errorMessageStyle.Render(t.GRPCStatus().Err().Error())
		_, _ = fmt.Fprintf(os.Stderr, "%s %s\n", errorMsg, detailMsg)
	default:
		// targeted suggestions for common user errors
		if isMissingSignerKeyFile(err) {
			errorMsg := errorStyle.Render(fmt.Sprintf("%s %s:", output.ErrorEmoji(), description))
			detailMsg := errorMessageStyle.Render(err.Error())
			_, _ = fmt.Fprintf(os.Stderr, "%s %s\n", errorMsg, detailMsg)

			suggestion := suggestionStyle.Render(buildMissingSignerKeySuggestion(err))
			_, _ = fmt.Fprintf(os.Stderr, "%s\n", suggestion)
			return
		}
		if errors.Is(err, config.ErrOutdatedFormat) {
			errorMsg := errorStyle.Render(fmt.Sprintf("%s Config Error:", output.ErrorEmoji()))
			detailMsg := errorMessageStyle.Render(err.Error())
			_, _ = fmt.Fprintf(os.Stderr, "%s %s\n", errorMsg, detailMsg)

			suggestion := suggestionStyle.Render(fmt.Sprintf("%s Please reset configuration using: 'flow init --reset'. Read more about new configuration here: https://github.com/onflow/flow-cli/releases/tag/v0.17.0", output.TryEmoji()))
			_, _ = fmt.Fprintf(os.Stderr, "%s\n", suggestion)
		} else if errors.Is(err, config.ErrDoesNotExist) {
			errorMsg := errorStyle.Render(fmt.Sprintf("%s Config Error:", output.ErrorEmoji()))
			detailMsg := errorMessageStyle.Render(err.Error())
			_, _ = fmt.Fprintf(os.Stderr, "%s %s\n", errorMsg, detailMsg)

			suggestion := suggestionStyle.Render(fmt.Sprintf("%s Please create configuration using: flow init", output.TryEmoji()))
			_, _ = fmt.Fprintf(os.Stderr, "%s\n", suggestion)
		} else if strings.Contains(err.Error(), "transport:") {
			errorMsg := errorStyle.Render(fmt.Sprintf("%s Connection Error:", output.ErrorEmoji()))
			detailMsg := errorMessageStyle.Render(strings.Split(err.Error(), "transport:")[1])
			_, _ = fmt.Fprintf(os.Stderr, "%s %s\n", errorMsg, detailMsg)

			suggestion := suggestionStyle.Render(fmt.Sprintf("%s Make sure your emulator is running or connection address is correct.", output.TryEmoji()))
			_, _ = fmt.Fprintf(os.Stderr, "%s\n", suggestion)
		} else if strings.Contains(err.Error(), "NotFound desc =") {
			errorMsg := errorStyle.Render(fmt.Sprintf("%s Not Found:", output.ErrorEmoji()))
			detailMsg := errorMessageStyle.Render(strings.Split(err.Error(), "NotFound desc =")[1])
			_, _ = fmt.Fprintf(os.Stderr, "%s%s\n", errorMsg, detailMsg)
		} else if strings.Contains(err.Error(), "code = InvalidArgument desc = ") {
			desc := strings.Split(err.Error(), "code = InvalidArgument desc = ")
			errorMsg := errorStyle.Render(fmt.Sprintf("%s Invalid argument:", output.ErrorEmoji()))
			detailMsg := errorMessageStyle.Render(desc[len(desc)-1])
			_, _ = fmt.Fprintf(os.Stderr, "%s %s\n", errorMsg, detailMsg)

			if strings.Contains(err.Error(), "is invalid for chain") {
				suggestion := suggestionStyle.Render(fmt.Sprintf("%s Check you are connecting to the correct network or account address you use is correct.", output.TryEmoji()))
				_, _ = fmt.Fprintf(os.Stderr, "%s\n", suggestion)
			} else {
				suggestion := suggestionStyle.Render(fmt.Sprintf("%s Check your argument and flags value, you can use --help.", output.TryEmoji()))
				_, _ = fmt.Fprintf(os.Stderr, "%s\n", suggestion)
			}
		} else if strings.Contains(err.Error(), "invalid signature:") {
			errorMsg := errorStyle.Render(fmt.Sprintf("%s Invalid signature:", output.ErrorEmoji()))
			detailMsg := errorMessageStyle.Render(strings.Split(err.Error(), "invalid signature:")[1])
			_, _ = fmt.Fprintf(os.Stderr, "%s %s\n", errorMsg, detailMsg)

			suggestion := suggestionStyle.Render(fmt.Sprintf("%s Check the signer private key is provided or is in the correct format. If running emulator, make sure it's using the same configuration as this command.", output.TryEmoji()))
			_, _ = fmt.Fprintf(os.Stderr, "%s\n", suggestion)
		} else if strings.Contains(err.Error(), "signature could not be verified using public key with") {
			errorMsg := errorStyle.Render(fmt.Sprintf("%s %s:", output.ErrorEmoji(), description))
			detailMsg := errorMessageStyle.Render(err.Error())
			_, _ = fmt.Fprintf(os.Stderr, "%s %s\n", errorMsg, detailMsg)

			suggestion := suggestionStyle.Render(fmt.Sprintf("%s If you are running emulator locally make sure that the emulator was started with the same config as used in this command. \nTry restarting the emulator.", output.TryEmoji()))
			_, _ = fmt.Fprintf(os.Stderr, "%s\n", suggestion)
		} else {
			errorMsg := errorStyle.Render(fmt.Sprintf("%s %s:", output.ErrorEmoji(), description))
			detailMsg := errorMessageStyle.Render(err.Error())
			_, _ = fmt.Fprintf(os.Stderr, "%s %s\n", errorMsg, detailMsg)
		}
	}

	fmt.Println()
	os.Exit(1)
}

// isMissingSignerKeyFile checks if the error indicates a missing signer private key file.
func isMissingSignerKeyFile(err error) bool {
	if err == nil {
		return false
	}

	msg := err.Error()
	// Match common file-not-found patterns originating from key file loading
	return strings.Contains(msg, "no such file or directory") &&
		(strings.Contains(msg, ".pkey") || strings.Contains(msg, "provided location") || strings.Contains(msg, "could not load the key") || strings.Contains(msg, "failed to load key") || strings.Contains(msg, "failed to read private key") || strings.Contains(msg, "open testnet.pkey"))
}

// buildMissingSignerKeySuggestion returns a user-friendly suggestion for resolving missing signer key file errors.
func buildMissingSignerKeySuggestion(err error) string {
	file := extractPathFromOpenErr(err.Error())
	if file == "" {
		// fallback generic message
		return fmt.Sprintf("%s Missing signer private key file. Ensure the file exists or update flow.json to point to a valid key. You can generate one with 'flow keys generate' or create an account with 'flow accounts create --network <network>'.", output.TryEmoji())
	}

	return fmt.Sprintf("%s Missing signer private key file. Ensure '%s' exists (relative to your project) or update 'flow.json' to point to a valid key. Generate a key with 'flow keys generate' or create one with 'flow accounts create --network <network>'. If your project stores keys under 'testnet.pkey' or similar, make sure the file is present or update the path.", output.TryEmoji(), file)
}

// extractPathFromOpenErr tries to extract the file path from an error containing the pattern "open <path>: ...".
func extractPathFromOpenErr(msg string) string {
	idx := strings.Index(msg, "open ")
	if idx == -1 {
		return ""
	}
	rest := msg[idx+len("open "):]
	// path is up to the next colon or end of string
	if colon := strings.Index(rest, ":"); colon != -1 {
		return strings.TrimSpace(rest[:colon])
	}
	return strings.TrimSpace(rest)
}
