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

package config

import (
	"regexp"
	"strings"

	"github.com/joho/godotenv"

	"github.com/a8m/envsubst"
)

var (
	fileRegex     = regexp.MustCompile(`"([^"]*)"\s*:\s*{\s*"fromFile"\s*:\s*"([^"]*)"\s*},?`)
	trailingComma = regexp.MustCompile(`,\s*}`)
)

// ProcessorRun all pre-processors.
func ProcessorRun(raw []byte) ([]byte, map[string]string) {
	rawString := string(raw)
	rawString = processEnv(rawString)
	rawString, accountFromFiles := processFile(rawString)

	return []byte(rawString), accountFromFiles
}

// processEnv finds env variables and insert env values.
func processEnv(raw string) string {
	_ = godotenv.Load() // try to load .env file

	raw, _ = envsubst.String(raw)
	return raw
}

// processFile finds file variables and insert content.
func processFile(raw string) (string, map[string]string) {
	fileMatches := fileRegex.FindAllStringSubmatch(raw, -1)
	accountFromFiles := map[string]string{}

	for _, match := range fileMatches {
		if len(match) < 3 {
			continue
		}

		accountFromFiles[match[1]] = match[2]

		// remove fromFile from config after we add that to composer
		raw = strings.ReplaceAll(raw, match[0], "")

		// remove possible trailing comma
		raw = trailingComma.ReplaceAllString(raw, "}")
	}

	return raw, accountFromFiles
}
