package flowkit

import "github.com/onflow/cadence"

// Script includes Cadence code and optional arguments and filename.
//
// Filename is only required to be passed if you want to resolve imports.
type Script struct {
	code     []byte
	Args     []cadence.Value
	location string
}

func NewScript(code []byte, args []cadence.Value, location string) *Script {
	return &Script{
		code:     code,
		Args:     args,
		location: location,
	}
}

func (s *Script) Code() []byte {
	return s.code
}

func (s *Script) SetCode(code []byte) {
	s.code = code
}

func (s *Script) Location() string {
	return s.location
}
