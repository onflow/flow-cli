package tools

import (
	"errors"
	"fmt"
	"os/user"
	"path"
	"runtime"
)

var errorPlatformNotSupported = errors.New("OS not supported, only supporting Windows and Mac OS")

const (
	darwin  = "darwin"
	windows = "windows"
)

// GetDefaultInstallDir returns default installation directory based on the OS.
func GetDefaultInstallDir() (string, error) {
	switch runtime.GOOS {
	case darwin:
		return "/Applications", nil
	case windows:
		// TODO: Search in common install directories
		// https://superuser.com/questions/1327037/what-choices-do-i-have-about-where-to-install-software-on-windows-10
		user, err := user.Current()
		if err != nil {
			return "", fmt.Errorf("could not find user information", err)
		}
		return path.Join(user.HomeDir, "AppData", "Local"), nil
	default:
		return "", errorPlatformNotSupported
	}
}
