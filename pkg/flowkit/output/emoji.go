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

import "runtime"

func printEmoji(emoji string) string {
	if runtime.GOOS == "windows" {
		return ""
	}

	return emoji
}

func ErrorEmoji() string {
	return printEmoji("❌")
}

func TryEmoji() string {
	return printEmoji("🙏")
}

func WarningEmoji() string {
	return printEmoji("⚠️")
}

func SaveEmoji() string {
	return printEmoji("💾")
}

func StopEmoji() string {
	return printEmoji("🔴️")
}

func GoEmoji() string {
	return printEmoji("🟢")
}

func OkEmoji() string {
	return printEmoji("✅")
}

func SuccessEmoji() string {
	return printEmoji("✨")
}
