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
	"path"
	"path/filepath"
)

const (
	cadenceDir     = "cadence"
	contractDir    = "contracts"
	scriptDir      = "scripts"
	transactionDir = "transactions"
	cadenceExt     = ".cdc"
)

func newProjectFiles(projectDir string) *projectFiles {
	return &projectFiles{cadenceDir: path.Join(projectDir, cadenceDir)}
}

type projectFiles struct {
	cadenceDir string
}

func (f *projectFiles) Contracts() ([]string, error) {
	return getFilePaths(path.Join(f.cadenceDir, contractDir))
}

func (f *projectFiles) Deployments() (map[string][]string, error) {
	deployments := make(map[string][]string)

	contracts, err := f.Contracts()
	if err != nil {
		return nil, err
	}

	for _, file := range contracts {
		accName := ""
		subAccPattern := fmt.Sprintf("**/%s/**/*%s", contractDir, cadenceExt)
		if match, _ := filepath.Match(subAccPattern, file); match {
			accName = filepath.Base(filepath.Dir(file))
		}
		deployments[accName] = append(deployments[accName], file)
	}

	return deployments, nil
}

func (f *projectFiles) Scripts() ([]string, error) {
	return getFilePaths(path.Join(f.cadenceDir, scriptDir))
}

func (f *projectFiles) Transactions() ([]string, error) {
	return getFilePaths(path.Join(f.cadenceDir, transactionDir))
}

func getFilePaths(dir string) ([]string, error) {
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
