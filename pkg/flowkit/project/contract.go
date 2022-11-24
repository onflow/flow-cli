package project

import (
	"github.com/onflow/cadence"
	"github.com/onflow/flow-go-sdk"
)

// Contract is a Cadence contract definition for a project.
type Contract struct {
	Name           string
	Location       string
	AccountAddress flow.Address
	AccountName    string
	Args           []cadence.Value
}
