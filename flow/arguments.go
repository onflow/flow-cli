package cli

import (
	"encoding/json"
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
func ParseArguments(input string) ([]cadence.Value, error) {
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
