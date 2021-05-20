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
