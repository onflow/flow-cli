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

package globalSettings

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// GlobalSettings contains the global settings configurations
//
// Use this to import/ export the settings file and update user settings
var GlobalSettings = flowCliSettings{}

// flowCliSettings contains settings for flow-cli to be persisted in settings.yaml
type flowCliSettings struct {
	MetricsEnabled bool `default:"true"`
	// Add other fields to persist here
}

const settingsName = "flow-cli.settings.yaml"
const settingsDir = "/FlowCLI"
const settingsFilePath = settingsDir + "/" + settingsName

// GetFileName returns the settings file name
func (cfg *flowCliSettings) GetFileName() string {
	return settingsName
}

// GetFilePath returns the settings file path
func (cfg *flowCliSettings) GetFilePath() string {
	return settingsFilePath
}

// Import is called on GlobalSettings to import the settings file
func (cfg *flowCliSettings) Import() {
	if err := cfg.importFile(); err != nil {
		fmt.Println("No " + cfg.GetFileName() + " file found. Using default settings.")
	}
}

// Export is called on GlobalSettings to write to the settings file
func (cfg *flowCliSettings) Export() {
	if err := cfg.exportFile(); err != nil {
		fmt.Println("Failed to export " + cfg.GetFileName())
	}
}

// importFile imports the settings file into flowCliSettings
func (cfg *flowCliSettings) importFile() error {
	dir, err := os.UserConfigDir()
	if err != nil {
		return err
	}

	f, err := os.ReadFile(dir + settingsFilePath)
	if err != nil {
		return err
	}

	if err := yaml.Unmarshal(f, cfg); err != nil {
		return err
	}

	return nil
}

// exportFile creates persisted settings file from flowCliSettings
func (cfg *flowCliSettings) exportFile() error {
	output, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}

	dir, err := os.UserConfigDir()
	if err != nil {
		return err
	}

	_ = os.Mkdir(dir+settingsDir, os.ModePerm)

	file, err := os.Create(dir + settingsFilePath)
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
