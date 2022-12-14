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
	"github.com/radovskyb/watcher"
	"io/fs"
	"path"
	"path/filepath"
	"time"
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
)

type accountChange struct {
	status int
	name   string
}

type contractChange struct {
	status  int
	path    string
	account string
}

func newProjectFiles(projectDir string) *projectFiles {
	return &projectFiles{
		cadenceDir: path.Join(projectDir, cadenceDir),
		watcher:    watcher.New(),
	}
}

type projectFiles struct {
	cadenceDir string
	watcher    *watcher.Watcher
}

func (f *projectFiles) contracts() ([]string, error) {
	return f.getFilePaths(contractDir)
}

func (f *projectFiles) deployments() (map[string][]string, error) {
	deployments := make(map[string][]string)

	contracts, err := f.contracts()
	if err != nil {
		return nil, err
	}

	for _, file := range contracts {
		accName, _ := accountFromPath(f.cadenceDir, file)
		deployments[accName] = append(deployments[accName], file)
	}

	return deployments, nil
}

func (f *projectFiles) scripts() ([]string, error) {
	return f.getFilePaths(scriptDir)
}

func (f *projectFiles) transactions() ([]string, error) {
	return f.getFilePaths(transactionDir)
}

func (f *projectFiles) watch() (<-chan accountChange, <-chan contractChange, error) {
	err := f.watcher.AddRecursive(contractDir)
	if err != nil {
		return nil, nil, err
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
		}

		for {
			select {
			case event := <-f.watcher.Event:
				name, containsAccount := accountFromPath(f.cadenceDir, event.Path)
				if event.IsDir() && containsAccount {
					// todo handle rename and move
					accounts <- accountChange{
						status: status[event.Op],
						name:   name,
					}
					continue
				}

				if filepath.Ext(event.Path) != cadenceExt { // skip any non cadence files
					continue
				}

				contracts <- contractChange{
					status:  status[event.Op],
					path:    event.Path,
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

func (f *projectFiles) getFilePaths(dir string) ([]string, error) {
	dir = path.Join(f.cadenceDir, dir)
	paths := make([]string, 0)
	err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if path == dir || d.IsDir() { // we only want to get the files in the dir
			return nil
		}
		if filepath.Ext(path) != cadenceExt {
			return nil
		}

		projectDir := filepath.Dir(filepath.Dir(dir)) // we want to include the cadence folder from the project path
		rel, err := filepath.Rel(projectDir, path)    // this will get files relative to project folder including cadence folder
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
	rel, err := filepath.Rel(path.Dir(f.cadenceDir), file)
	if err != nil {
		return "", err
	}
	return rel, nil
}

// accountFromPath returns the account name from provided path if possible, otherwise returns empty and false.
//
// Account name can be extracted from path when the contract folder contains another folder, that in our syntax indicates account name.
func accountFromPath(cadenceDir string, path string) (string, bool) {
	// extract account from path if file path is provided e.g.: /Foo/contracts/[alice]/foo.cdc
	subAccPattern := filepath.Clean(fmt.Sprintf("%s/%s/*/*%s", cadenceDir, contractDir, cadenceExt))
	if match, _ := filepath.Match(subAccPattern, path); match {
		return filepath.Base(filepath.Dir(path)), true
	}
	// extract account from path if dir path is provided e.g.: /Foo/contracts/[alice]
	subAccPattern = filepath.Clean(fmt.Sprintf("%s/%s/*", cadenceDir, contractDir))
	if match, _ := filepath.Match(subAccPattern, path); match {
		if filepath.Ext(path) != "" { // might be a file
			return "", false
		}
		return filepath.Base(path), true
	}

	return "", false
}
