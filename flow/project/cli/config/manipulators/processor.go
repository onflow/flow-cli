/*
 * Flow CLI
 *
 * Copyright 2019-2020 Dapper Labs, Inc.
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

package manipulators

import (
	"os"
	"regexp"
	"strings"
)

var (
	envRegex      *regexp.Regexp = regexp.MustCompile(`\$\{env\:(.+)\}`)
	fileRegex     *regexp.Regexp = regexp.MustCompile(`"([^"]*)"\s*:\s*{\s*"fromFile"\s*:\s*"([^"]*)"\s*},?`)
	trailingComma *regexp.Regexp = regexp.MustCompile(`\,\s*}`)
)

// Preprocessor is used to pre-process configuration file
type Preprocessor struct {
	composer *Composer
}

// NewPreprocessor creates new instance of preprocessor
func NewPreprocessor(composer *Composer) *Preprocessor {
	return &Preprocessor{
		composer: composer,
	}
}

// Run all pre-processors
func (p *Preprocessor) Run(raw []byte) []byte {
	rawString := string(raw)
	rawString = p.processEnv(rawString)
	rawString = p.processFile(rawString)

	return []byte(rawString)
}

// processEnv finds env variables and insert env values
func (p *Preprocessor) processEnv(raw string) string {
	envMatches := envRegex.FindAllStringSubmatch(raw, -1)

	for _, match := range envMatches {
		raw = strings.ReplaceAll(
			raw,
			match[0],
			p.getEnvVariable(match[1]),
		)
	}

	return raw
}

// processFile finds file variables and insert content
func (p *Preprocessor) processFile(raw string) string {
	fileMatches := fileRegex.FindAllStringSubmatch(raw, -1)

	for _, match := range fileMatches {
		if len(match) < 3 {
			continue
		}

		p.composer.AddAccountFromFile(match[2], match[1])

		// remove from config
		raw = strings.ReplaceAll(raw, match[0], "")

		// remove possible trailing comma
		raw = trailingComma.ReplaceAllString(raw, "}")
	}

	return raw
}

// get environment variable by name
func (p *Preprocessor) getEnvVariable(name string) string {
	return os.Getenv(name)
}
