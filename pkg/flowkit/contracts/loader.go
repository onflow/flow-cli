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

package contracts

import (
	"github.com/onflow/flow-cli/pkg/flowkit"
)

// Loader defines methods for loading contract resource.
type Loader interface {
	Load(source string) ([]byte, error)
	Normalize(base, relative string) string
}

// FilesystemLoader defines contract loader from files.
type FilesystemLoader struct {
	Reader flowkit.ReaderWriter
}

func (f FilesystemLoader) Load(source string) ([]byte, error) {
	codeBytes, err := f.Reader.ReadFile(source)
	if err != nil {
		return nil, err
	}

	return codeBytes, nil
}

func (f FilesystemLoader) Normalize(base, relative string) string {
	return absolutePath(base, relative)
}
