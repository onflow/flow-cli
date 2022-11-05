package util

import (
	"fmt"
	"gopkg.in/yaml.v3"
	"os"
)

// configFile is the global flow-cli config for the config file
//
// imported is set to true after the config file has been loaded
var configFile struct {
	imported bool `default:"false"`
	flowCliConfig
}

// GetConfigFile is called globally to access the config file
func GetConfigFile() *flowCliConfig {
	if !configFile.imported {
		err := configFile.importConfig()
		if err != nil {
			fmt.Println("Failed to import config file from" + configFilePath + ". Using default config.")
		}
		configFile.imported = true
	}

	return &configFile.flowCliConfig
}

// ExportConfigFile is called to create / update the persisted config file
func ExportConfigFile() {
	err := configFile.exportConfig()
	if err != nil {
		fmt.Println("Failed to export config file " + configName)
	}
}

const configName = "flow-cli.config.yaml"
const configDir = "/FlowCLI"
const configFilePath = configDir + "/" + configName

// flowCliConfig configurations for flow-cli to be persisted
type flowCliConfig struct {
	MetricsEnabled bool `default:"false"`
	// Add other fields to persist here
}

// exportConfig creates persisted config file from flowCliConfig
func (cfg *flowCliConfig) exportConfig() error {
	output, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}

	dir, err := os.UserConfigDir()
	if err != nil {
		return err
	}

	_ = os.Mkdir(dir+configDir, os.ModePerm)

	file, err := os.Create(dir + configFilePath)
	if err != nil {
		return err
	}

	defer func() {
		if err := file.Close(); err != nil {
			return
		}
	}()

	_, err = file.Write(output)
	if err != nil {
		return err
	}

	return nil
}

// importConfig imports the config file into flowCliConfig
func (cfg *flowCliConfig) importConfig() error {
	dir, err := os.UserConfigDir()
	if err != nil {
		return err
	}

	f, err := os.ReadFile(dir + configFilePath)
	if err != nil {
		return err
	}

	if err := yaml.Unmarshal(f, cfg); err != nil {
		return err
	}

	return nil
}
