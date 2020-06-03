package types

import "github.com/dapperlabs/flow-go/engine/execution/computation/virtualmachine"

type FlowError struct {
	FlowError virtualmachine.FlowError
}

func (f *FlowError) Error() string {
	return f.FlowError.ErrorMessage()
}
