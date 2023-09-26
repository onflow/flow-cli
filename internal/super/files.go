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

package super

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"time"

	"github.com/pkg/errors"
	"github.com/radovskyb/watcher"

	"github.com/onflow/flow-cli/flowkit/config"
)

const (
	cadenceDir     = "cadence"
	contractDir    = "contracts"
	scriptDir      = "scripts"
	transactionDir = "transactions"
	cadenceExt     = ".cdc"
	created        = 1
	removed        = 2
	changed        = 3
	renamed        = 4
)

type accountChange struct {
	status int
	name   string
}

type contractChange struct {
	status  int
	path    string
	oldPath string
	account string
}

func newProjectFiles(projectPath string) *projectFiles {
	return &projectFiles{
		cadencePath: filepath.Join(projectPath, cadenceDir),
		watcher:     watcher.New(),
	}
}

type projectFiles struct {
	cadencePath string
	watcher     *watcher.Watcher
}

// exist checks if current directory contains all project files required.
func (f *projectFiles) exist() error {
	if _, err := os.Stat(f.cadencePath); errors.Is(err, os.ErrNotExist) {
		if _, err := os.Stat(contractDir); errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("required cadence folder or contract folder does not exist")
		} else if err == nil {
			f.cadencePath = "" // in case there's no cadence folder we just use contracts folder directly
		}
	}
	if _, err := os.Stat(config.DefaultPath); errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("required project configuration ('flow.json') does not exist")
	}

	return nil
}

// contracts returns a list of contracts in project.
func (f *projectFiles) contracts() ([]string, error) {
	return f.getCadenceFilepaths(contractDir)
}

// contracts returns a map of deployments in project, mapping account name to all contracts saved to that account.
func (f *projectFiles) deployments() (map[string][]string, error) {
	deployments := make(map[string][]string)

	contracts, err := f.contracts()
	if err != nil {
		return nil, errors.Wrap(err, "failed to get contracts in deployment")
	}

	for _, file := range contracts {
		accName, _ := accountFromPath(file)
		deployments[accName] = append(deployments[accName], file)
	}

	return deployments, nil
}

// contracts returns a list of scripts in project.
func (f *projectFiles) scripts() ([]string, error) {
	return f.getCadenceFilepaths(scriptDir)
}

// contracts returns a list of transactions in project.
func (f *projectFiles) transactions() ([]string, error) {
	return f.getCadenceFilepaths(transactionDir)
}

// watch for file changes in the contract folder and signal any changes through channel.
//
// This function returns two channels, accountChange which reports any changes on the accounts folders and
// contractChange which reports any changes to the contract files.
func (f *projectFiles) watch() (<-chan accountChange, <-chan contractChange, error) {
	err := f.watcher.AddRecursive(filepath.Join(f.cadencePath, contractDir))
	if err != nil {
		return nil, nil, errors.Wrap(err, "add recursive files failed")
	}

	go func() {
		err = f.watcher.Start(500 * time.Millisecond)
		if err != nil {
			panic(err)
		}
	}()

	accounts := make(chan accountChange)
	contracts := make(chan contractChange)

	go func() {
		status := map[watcher.Op]int{
			watcher.Create: created,
			watcher.Remove: removed,
			watcher.Write:  changed,
			watcher.Rename: renamed,
		}

		for {
			select {
			case event := <-f.watcher.Event:
				rel, err := f.relProjectPath(event.Path)
				if err != nil { // skip if failed
					continue
				}

				name, containsAccount := accountFromPath(rel)
				if event.IsDir() && containsAccount {
					// TODO(sideninja) handle moving of files
					accounts <- accountChange{
						status: status[event.Op],
						name:   name,
					}
					continue
				}

				if filepath.Ext(rel) != cadenceExt { // skip any non cadence files
					continue
				}

				oldPath := ""
				if event.Op == watcher.Rename { // add relative path in case of rename
					oldPath, err = f.relProjectPath(event.OldPath)
					if err != nil {
						continue
					}
				}

				contracts <- contractChange{
					status:  status[event.Op],
					path:    rel,
					oldPath: oldPath,
					account: name,
				}
			case <-f.watcher.Closed:
				close(contracts)
				close(accounts)
				return
			}
		}
	}()

	return accounts, contracts, nil
}

// getFilePaths returns a list of only Cadence files that are inside the provided directory.
func (f *projectFiles) getCadenceFilepaths(dir string) ([]string, error) {
	dir = filepath.Join(f.cadencePath, dir)
	paths := make([]string, 0)
	err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if path == dir || d.IsDir() || filepath.Ext(path) != cadenceExt { // we only want to get .cdc files in the dir
			return nil
		}

		rel, err := f.relProjectPath(path)
		if err != nil {
			return err
		}

		paths = append(paths, rel)
		return nil
	})
	if err != nil {
		return nil, err
	}

	return paths, nil
}

// relProjectPath gets a filepath relative to the project directory including the base cadence directory.
// eg. a path /Users/Mike/Dev/project/cadence/contracts/foo.cdc will become cadence/contracts/foo.cdc
func (f *projectFiles) relProjectPath(file string) (string, error) {
	rel, err := filepath.Rel(filepath.Dir(f.cadencePath), file)
	if err != nil {
		return "", errors.Wrap(err, "failed getting project relative path")
	}
	return rel, nil
}

// accountFromPath returns the account name from provided path if possible, otherwise returns empty and false.
//
// Account name can be extracted from path when the contract folder contains another folder, that in our syntax indicates account name.
func accountFromPath(path string) (string, bool) {
	// extract account from path if file path is provided e.g.: cadence/contracts/[alice]/foo.cdc
	subAccPattern := filepath.Clean(fmt.Sprintf("%s/%s/*/*%s", cadenceDir, contractDir, cadenceExt))
	if match, _ := filepath.Match(subAccPattern, path); match {
		return filepath.Base(filepath.Dir(path)), true
	}
	// extract account from path if dir path is provided e.g.: cadence/contracts/[alice]
	subAccPattern = filepath.Clean(fmt.Sprintf("%s/%s/*", cadenceDir, contractDir))
	if match, _ := filepath.Match(subAccPattern, path); match {
		if filepath.Ext(path) != "" { // might be a file
			return "", false
		}
		return filepath.Base(path), true
	}

	return "", false
}
