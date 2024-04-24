package util

import (
	"fmt"
	"runtime"
)

func PrintEmoji(emoji string) string {
	if runtime.GOOS == "windows" {
		return ""
	}

	return emoji
}

func MessageWithEmojiPrefix(emoji string, message string) string {
	return fmt.Sprintf("%s%s", PrintEmoji(emoji+" "), message)
}
