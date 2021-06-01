/*
 * Flow CLI
 *
 * Copyright 2019-2021 Dapper Labs, Inc.
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

package services

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/onflow/flow-cli/pkg/flowcli/gateway"
	"github.com/onflow/flow-cli/pkg/flowcli/output"
	"github.com/onflow/flow-cli/pkg/flowcli/project"
	"github.com/spf13/afero"
)

const stateDir = "./states"

// State is a service that handles state
type State struct {
	gateway gateway.Gateway
	project *project.Project
	logger  output.Logger
}

// NewState returns a new state service
func NewState(
	gateway gateway.Gateway,
	project *project.Project,
	logger output.Logger,
) *State {
	return &State{
		gateway: gateway,
		project: project,
		logger:  logger,
	}
}

func (s *State) Save(name string) error {
	af := afero.Afero{ // todo refactor to pass instance as arg
		Fs: afero.NewOsFs(),
	}

	currentStatePath := path.Join(stateDir, "main") // todo get from config
	newStatePath := path.Join(stateDir, name)
	newStateUsagePath := path.Join(newStatePath, "main")
	newStateSnapshotPath := path.Join(newStatePath, "snapshot")

	exists, err := af.DirExists(newStatePath)
	if err != nil {
		return err
	} else if exists {
		return fmt.Errorf("state named: %s already exists") // todo overwrite option
	}

	err = copy(currentStatePath, newStateUsagePath)
	if err != nil {
		return err
	}
	err = copy(currentStatePath, newStateSnapshotPath)
	if err != nil {
		return err
	}

	s.logger.Info(fmt.Sprintf("state [%s] saved, to load the state execute state use command"))

	return nil
}

func copy(source, destination string) error {
	af := afero.Afero{
		Fs: afero.NewOsFs(),
	}

	err := af.Walk(source, func(path string, info os.FileInfo, err error) error {
		relPath := strings.Replace(path, source, "", 1)
		if relPath == "" {
			return nil
		}
		if info.IsDir() {
			return af.Mkdir(filepath.Join(destination, relPath), 0755)
		} else {
			data, err := af.ReadFile(filepath.Join(source, relPath))
			if err != nil {
				return err
			}

			return af.WriteFile(filepath.Join(destination, relPath), data, 0777)
		}
	})

	return err
}
