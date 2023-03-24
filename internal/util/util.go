package util

import (
	"fmt"
	"github.com/onflow/flow-cli/pkg/flowkit"
	"os"
	"path"
)

const EnvPrefix = "FLOW"

func Exit(code int, msg string) {
	fmt.Fprintln(os.Stderr, msg)
	os.Exit(code)
}

// AddToGitIgnore adds a new line to the .gitignore if one doesn't exist it creates it.
func AddToGitIgnore(filename string, loader flowkit.ReaderWriter) error {
	currentWd, err := os.Getwd()
	if err != nil {
		return err
	}
	gitIgnorePath := path.Join(currentWd, ".gitignore")
	gitIgnoreFiles := ""
	filePermissions := os.FileMode(0644)

	fileStat, err := os.Stat(gitIgnorePath)
	if !os.IsNotExist(err) { // if gitignore exists
		gitIgnoreFilesRaw, err := loader.ReadFile(gitIgnorePath)
		if err != nil {
			return err
		}
		gitIgnoreFiles = string(gitIgnoreFilesRaw)
		filePermissions = fileStat.Mode().Perm()
	}
	return loader.WriteFile(
		gitIgnorePath,
		[]byte(fmt.Sprintf("%s\n%s", gitIgnoreFiles, filename)),
		filePermissions,
	)
}
