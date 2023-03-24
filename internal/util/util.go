package util

import (
	"fmt"
	"os"
)

func Exit(code int, msg string) {
	fmt.Fprintln(os.Stderr, msg)
	os.Exit(code)
}
