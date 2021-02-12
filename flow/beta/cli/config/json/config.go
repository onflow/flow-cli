package json

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/onflow/flow-cli/flow/beta/cli/config"
)

type jsonConfig struct {
	Contracts jsonContracts `json:"contracts"`
	Networks  jsonNetworks  `json:"networks"`
	Accounts  jsonAccounts  `json:"accounts"`
	Deploys   jsonDeploys   `json:"deploys"`
}

func (j jsonConfig) transformToConfig() *config.Config {
	return &config.Config{
		Contracts: j.Contracts.transformToConfig(),
		Networks:  j.Networks.transformToConfig(),
		Accounts:  j.Accounts.transformToConfig(),
		Deploys:   j.Deploys.transformToConfig(),
	}
}

// Save saves the configuration to the specified path file in JSON format.
func Save(conf *config.Config, path string) error {
	data, err := json.MarshalIndent(conf, "", "\t")
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(path, data, 0777)
	if err != nil {
		return err
	}

	return nil
}

// ErrDoesNotExist is error to be returned when config file does not exists
var ErrDoesNotExist = errors.New("project config file does not exist")

func Load(path string) (*config.Config, error) {
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, ErrDoesNotExist
		}

		return nil, err
	}

	d := json.NewDecoder(f)
	conf := new(jsonConfig)
	err = d.Decode(conf)

	if err != nil {
		fmt.Printf("%s contains invalid json: %s\n", path, err.Error())
		os.Exit(1)
	}

	return conf.transformToConfig(), nil
}

// Exists checks if file exists on the specified path
func Exists(path string) bool {
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}
