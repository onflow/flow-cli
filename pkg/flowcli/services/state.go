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

	"github.com/otiai10/copy"
	"github.com/spf13/afero"

	"github.com/onflow/flow-cli/pkg/flowcli/config"
	"github.com/onflow/flow-cli/pkg/flowcli/gateway"
	"github.com/onflow/flow-cli/pkg/flowcli/output"
	"github.com/onflow/flow-cli/pkg/flowcli/project"
)

var mainPath = path.Join(config.StateDir, config.MainState)

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

// Snapshot takes a snapshot of the current state
func (s *State) Snapshot(name string) error {
	// todo refactor to pass instance as arg
	af := afero.Afero{
		Fs: afero.NewOsFs(),
	}

	snapshotPath := path.Join(config.StateDir, name)

	exists, err := af.DirExists(snapshotPath)
	if err != nil {
		return err
	} else if exists {
		return fmt.Errorf("snapshot named: %s already exist", name) // todo implement overwrite option
	}

	err = copy.Copy(mainPath, snapshotPath)
	if err != nil {
		return err
	}

	s.logger.Info(fmt.Sprintf("snapshot of state [%s] saved, use restore command to restore from the snapshot", name))

	return nil
}

// Restore restores the snapshot by the name
func (s *State) Restore(name string) error {
	// todo refactor to pass instance as arg
	af := afero.Afero{
		Fs: afero.NewOsFs(),
	}

	snapshotPath := path.Join(config.StateDir, name)

	exists, err := af.DirExists(snapshotPath)
	if err != nil {
		return err
	} else if !exists {
		return fmt.Errorf("snapshot named: %s doesn't exist", name) // todo implement overwrite option
	}

	err = copy.Copy(snapshotPath, mainPath)
	if err != nil {
		return err
	}

	s.logger.Info(fmt.Sprintf("state was restored to the snapshot state with name: %s", name))

	return nil
}

func (s *State) Remove(name string) error {
	// todo refactor to pass instance as arg
	af := afero.Afero{
		Fs: afero.NewOsFs(),
	}

	snapshotPath := path.Join(config.StateDir, name)

	exists, err := af.DirExists(snapshotPath)
	if err != nil {
		return err
	} else if !exists {
		return fmt.Errorf("snapshot named: %s doesn't exist", name) // todo implement overwrite option
	}

	err = os.RemoveAll(snapshotPath)
	if err != nil {
		return err
	}

	s.logger.Info(fmt.Sprintf("snapshot with name: %s was removed", name))

	return nil
}

// List lists all snapshots
func (s *State) List() ([]string, error) {
	// todo refactor to pass instance as arg
	af := afero.Afero{
		Fs: afero.NewOsFs(),
	}

	var files []string
	fileInfo, err := af.ReadDir(config.StateDir)
	if err != nil {
		return nil, err
	}

	for _, file := range fileInfo {
		files = append(files, file.Name())
	}
	return files, nil
}
