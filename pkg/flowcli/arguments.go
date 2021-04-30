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

package flowcli

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/onflow/cadence"
	jsoncdc "github.com/onflow/cadence/encoding/json"
)

type CadenceArgument struct {
	Value cadence.Value
}

func (v CadenceArgument) MarshalJSON() ([]byte, error) {
	return jsoncdc.Encode(v.Value)
}

func (v *CadenceArgument) UnmarshalJSON(b []byte) (err error) {
	v.Value, err = jsoncdc.Decode(b)
	if err != nil {
		return err
	}
	return nil
}

func ParseArgumentsJSON(input string) ([]cadence.Value, error) {
	var args []CadenceArgument
	b := []byte(input)
	err := json.Unmarshal(b, &args)

	if err != nil {
		return nil, err
	}

	cadenceArgs := make([]cadence.Value, len(args))
	for i, arg := range args {
		cadenceArgs[i] = arg.Value
	}
	return cadenceArgs, nil
}

func ParseArgumentsCommaSplit(input []string) ([]cadence.Value, error) {
	args := make([]map[string]interface{}, 0)

	if len(input) == 0 {
		return make([]cadence.Value, 0), nil
	}

	for _, in := range input {
		argInput := strings.Split(in, ":")
		argType := argInput[0]
		argValue := argInput[1]

		if len(argInput) != 2 {
			return nil, fmt.Errorf(
				"argument not passed in correct format, correct format is: Type:Value, got %s",
				in,
			)
		}

		args = append(args, map[string]interface{}{
			"value": processValue(argType, argValue),
			"type":  argType,
		})
	}
	jsonArgs, _ := json.Marshal(args)
	cadenceArgs, err := ParseArgumentsJSON(string(jsonArgs))

	return cadenceArgs, err
}

// sanitizeAddressArg sanitize address and make sure it has 0x prefix
func processValue(argType string, argValue string) interface{} {
	if argType == "Address" && !strings.Contains(argValue, "0x") {
		return fmt.Sprintf("0x%s", argValue)
	} else if argType == "Bool" {
		converted, _ := strconv.ParseBool(argValue)
		return converted
	}

	return argValue
}

func ParseArguments(args []string, argsJSON string) (scriptArgs []cadence.Value, err error) {
	if argsJSON != "" {
		scriptArgs, err = ParseArgumentsJSON(argsJSON)
	} else {
		scriptArgs, err = ParseArgumentsCommaSplit(args)
	}
	if err != nil {
		return nil, err
	}

	return
}
