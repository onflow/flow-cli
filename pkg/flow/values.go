package flow

import (
	"github.com/onflow/cadence"
)

func NewStakingInfoFromValue(value cadence.Value) map[string]interface{} {
	stakingInfo := make(map[string]interface{})
	arrayValue := value.(cadence.Array)

	if len(arrayValue.Values) == 0 {
		return stakingInfo
	}

	keys := make([]string, 0)
	for _, field := range arrayValue.Values[0].(cadence.Struct).StructType.Fields {
		keys = append(keys, field.Identifier)
	}

	for i, value := range arrayValue.Values[0].(cadence.Struct).Fields {
		stakingInfo[keys[i]] = value
	}

	return stakingInfo
}
