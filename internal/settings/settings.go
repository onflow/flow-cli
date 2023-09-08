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

package settings

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

const settingsFile = "flow-cli.settings"

const settingsDir = "flow-cli"

const settingsType = "yaml"

// viperLoaded only load settings file once
var viperLoaded = false

func init() {
	viper.SetConfigName(settingsFile)
	viper.SetConfigType(settingsType)
	viper.AddConfigPath(FileDir())
}

func FileName() string {
	return fmt.Sprintf("%s.%s", settingsFile, settingsType)
}

func FileDir() string {
	dir, err := os.UserConfigDir()
	if err != nil {
		dir = "."
	}
	return filepath.Join(dir, settingsDir)
}

// Set updates settings file with new value for provided key
func Set(key string, val any) error {
	if err := loadViper(); err != nil {
		return err
	}

	viper.Set(key, val)
	if err := viper.WriteConfig(); err != nil {
		return err
	}

	return nil
}

// loadViper loads the global settings file
func loadViper() error {
	if viperLoaded {
		return nil
	}
	viperLoaded = true

	if err := createSettingsDir(); err != nil {
		return err
	}

	err := viper.MergeConfigMap(defaults)
	if err != nil {
		return err
	}

	// Load settings file
	if err := viper.MergeInConfig(); err != nil {
		switch err.(type) {
		case viper.ConfigFileNotFoundError:
			// Create settings file for the first time
			if err = viper.SafeWriteConfig(); err != nil {
				return err
			}
		default:
			return err
		}
	}

	return nil
}

// createSettingsDir creates settings dir if it doesn't exist
func createSettingsDir() error {
	if _, err := os.Stat(FileDir()); errors.Is(err, os.ErrNotExist) {
		err = os.Mkdir(FileDir(), os.ModePerm)
		if err != nil {
			return err
		}
	}
	return nil
}

// GetFlowserPath gets set Flowser install path with sensible default.
func GetFlowserPath() (string, error) {
	if err := loadViper(); err != nil {
		return "", err
	}
	return viper.GetString(flowserPath), nil
}

func SetFlowserPath(path string) error {
	return Set(flowserPath, path)
}

// MetricsEnabled checks whether metric tracking is enabled.
func MetricsEnabled() bool {
	if err := loadViper(); err != nil {
		return true
	}
	return viper.GetBool(metricsEnabled)
}
