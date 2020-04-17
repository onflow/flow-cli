package types

import "github.com/dapperlabs/flow-go-sdk"

type StorableTransactionResult struct {
	ErrorCode    int
	ErrorMessage string
	Logs         []string
	Events       []flow.Event
}
