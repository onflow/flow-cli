package types

import (
	"github.com/onflow/flow-go/fvm"
)

type FlowError struct {
	FlowError fvm.Error
}

func (f *FlowError) Error() string {
	return f.FlowError.Error()
}
