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

package config

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/spf13/afero"
)

var (
	envRegex  *regexp.Regexp = regexp.MustCompile(`\$\{env\:(.+)\}`)
	fileRegex *regexp.Regexp = regexp.MustCompile(`\"\$\{file\:(.+)\}\"`)
)

// Preprocessor is used to pre-process configuration file
type Preprocessor struct {
	af *afero.Afero
}

// NewPreprocessor creates new instance of preprocessor
func NewPreprocessor(filesystem afero.Fs) *Preprocessor {
	af := &afero.Afero{Fs: filesystem}

	return &Preprocessor{af: af}
}

// Run all pre-processors
func (p *Preprocessor) Run(raw string) string {
	raw = p.processEnv(raw)
	raw = p.processFile(raw)
	return raw
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
		configFileRaw, err := p.af.ReadFile(match[1])

		if err != nil {
			fmt.Printf("Config file %s not found. \n", match[1])
			os.Exit(1)
		}

		raw = strings.ReplaceAll(
			raw,
			match[0],
			string(configFileRaw),
		)
	}

	return raw
}

// get environment variable by name
func (p *Preprocessor) getEnvVariable(name string) string {
	return os.Getenv(name)
}
