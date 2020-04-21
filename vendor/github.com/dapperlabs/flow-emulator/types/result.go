package types

import "github.com/onflow/flow-go-sdk"

type StorableTransactionResult struct {
	ErrorCode    int
	ErrorMessage string
	Logs         []string
	Events       []flow.Event
}
