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
	"io/fs"
	"path"
	"path/filepath"
)

const (
	contractDir    = "contracts"
	scriptDir      = "scripts"
	transactionDir = "transactions"
	cadenceExt     = ".cdc"
)

type projectFiles struct {
	projectDir string
}

func (f *projectFiles) Contracts() ([]string, error) {
	return getFilePaths(path.Join(f.projectDir, contractDir))
}

func (f *projectFiles) Accounts() ([]string, error) {
	paths := make([]string, 0)
	dir := filepath.Join(f.projectDir, contractDir)
	err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if path == dir || !d.IsDir() { // we only want to get directories
			return nil
		}
		if filepath.Ext(path) != cadenceExt {
			return nil
		}

		// we only need the folder name
		paths = append(paths, filepath.Base(path))
		return nil
	})
	if err != nil {
		return nil, err
	}

	return paths, nil
}

func (f *projectFiles) Scripts() ([]string, error) {
	return getFilePaths(path.Join(f.projectDir, scriptDir))
}

func (f *projectFiles) Transactions() ([]string, error) {
	return getFilePaths(path.Join(f.projectDir, transactionDir))
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

		projectDir := filepath.Dir(dir) // we want to include the folder containing files in the path
		rel, err := filepath.Rel(projectDir, path)
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
