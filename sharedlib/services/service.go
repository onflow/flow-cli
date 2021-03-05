package services

import "github.com/onflow/flow-go-sdk"

type Service interface {
	GetAccount(address string) (*flow.Account, error)
}
