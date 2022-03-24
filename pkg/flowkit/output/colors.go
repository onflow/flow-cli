/*
 * Flow CLI
 *
 * Copyright 2019 Dapper Labs, Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *   http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

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
