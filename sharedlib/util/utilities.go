package util

import (
	"fmt"
	"io/ioutil"
)

func LoadFile(filename string) ([]byte, error) {
	var code []byte
	var err error

	if filename != "" {
		code, err = ioutil.ReadFile(filename)
		if err != nil {
			return nil, fmt.Errorf("Failed to load file: %s", filename)
		}
	}

	return code, nil
}

func IsByteSlice(v interface{}) bool {
	slice, isSlice := v.([]interface{})
	if !isSlice {
		return false
	}
	_, isBytes := slice[0].(byte)
	return isBytes
}
