package output

import (
	"fmt"
	"runtime"
)

var red = "\033[31m"
var green = "\033[32m"
var bold = "\033[1m"
var reset = "\033[0m"

func printColor(msg string, color string) string {
	if runtime.GOOS == "windows" {
		return msg
	}

	return fmt.Sprintf("%s%s%s", color, msg, reset)
}

func Red(msg string) string {
	return printColor(msg, red)
}

func Green(msg string) string {
	return printColor(msg, green)
}

func Bold(msg string) string {
	return printColor(msg, bold)
}
