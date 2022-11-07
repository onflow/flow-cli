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

	"github.com/spf13/viper"
)

const settingsFile = "flow-cli.settings"
const settingsDir = "/FlowCLI"
const settingsType = "yaml"

func SettingsFile() string {
	return settingsFile + "." + settingsType
}

// Init is called to initialize global settings
func Init() {
	_ = viperInit()
}

// Set updates settings file with new value for provided key
func Set(key string, val interface{}) {
	viper.Set(key, val)
	if err := viper.WriteConfig(); err != nil {
		fmt.Println("Failed to update " + SettingsFile())
	}
}

func Get(key string) interface{} {
	return viper.Get(key)
}

func GetBool(key string) bool {
	return viper.GetBool(key)
}

func GetString(key string) string {
	return viper.GetString(key)
}

func GetInt(key string) int {
	return viper.GetInt(key)
}

// viperInit initializes global settings with viper and reads in the settings file
func viperInit() error {
	viper.SetConfigName(settingsFile)
	viper.SetConfigType(settingsType)

	// Set path to settings file
	dir, err := os.UserConfigDir()
	if err != nil {
		return err
	}
	viper.AddConfigPath(dir + settingsDir)

	err = viper.MergeConfigMap(defaults)
	if err != nil {
		fmt.Println("Failed to set default settings: ", err.Error())
	}

	if err := viper.ReadInConfig(); err != nil {
		fmt.Println("No " + settingsFile + " found. Using default settings")
		_ = viper.SafeWriteConfig()
	}

	return nil
}
