package util

import (
	"fmt"
	"os"
)

const EnvPrefix = "FLOW"

func Exit(code int, msg string) {
	fmt.Fprintln(os.Stderr, msg)
	os.Exit(code)
}
