package flowkit

import "github.com/onflow/cadence"

// Script includes Cadence code and optional arguments and filename.
//
// Filename is only required to be passed if you want to resolve imports.
type Script struct {
	Code     []byte
	Args     []cadence.Value
	Location string
}
