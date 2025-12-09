/*
 * Flow CLI
 *
 * Copyright Flow Foundation
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

package util

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/onflow/cadence/common"
)

func AddCDCExtension(name string) string {
	if strings.HasSuffix(name, ".cdc") {
		return name
	}
	return fmt.Sprintf("%s.cdc", name)
}

func StripCDCExtension(name string) string {
	return strings.TrimSuffix(name, filepath.Ext(name))
}

// AbsolutePath resolves a relative path against a base file path.
// If the relative path is already absolute, it returns it as-is.
// Otherwise, it joins the relative path to the parent directory of the base path.
func AbsolutePath(basePath, relativePath string) string {
	if filepath.IsAbs(relativePath) {
		return relativePath
	}
	return filepath.Join(filepath.Dir(basePath), relativePath)
}

// IsPathLocation returns true if the location is a file path (contains .cdc)
func IsPathLocation(location common.Location) bool {
	stringLocation, ok := location.(common.StringLocation)
	if !ok {
		return false
	}
	return strings.Contains(stringLocation.String(), ".cdc")
}

// NormalizePathLocation normalizes a relative path import against a base location
func NormalizePathLocation(base, relative common.Location) common.Location {
	baseString, baseOk := base.(common.StringLocation)
	relativeString, relativeOk := relative.(common.StringLocation)

	if !baseOk || !relativeOk {
		return relative
	}

	normalizedPath := AbsolutePath(baseString.String(), relativeString.String())
	return common.StringLocation(normalizedPath)
}
