package lib

import (
	"encoding/json"
	"fmt"
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
	args := make([]map[string]string, 0)

	if len(input) == 0 {
		return make([]cadence.Value, 0), nil
	}

	for _, in := range input {
		argInput := strings.Split(in, ":")
		if len(argInput) != 2 {
			return nil, fmt.Errorf("Argument not passed in correct format, correct format is: Type:Value, got %s", in)
		}

		args = append(args, map[string]string{
			"value": string(argInput[1]),
			"type":  string(argInput[0]),
		})
	}
	jsonArgs, _ := json.Marshal(args)
	cadenceArgs, err := ParseArgumentsJSON(string(jsonArgs))

	return cadenceArgs, err
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
