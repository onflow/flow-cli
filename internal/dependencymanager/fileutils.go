package dependencymanager

import (
	"os"
	"path/filepath"
)

func contractFileExists(address, contractName string) bool {
	path := filepath.Join("imports", address, contractName)
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

func createContractFile(address, contractName, data string) error {
	path := filepath.Join("imports", address, contractName)

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		return err
	}

	return os.WriteFile(path, []byte(data), 0644)
}
