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

// Package build contains information about the build that injected at build-time.
//
// To use this package, simply import it in your program, then add build
// arguments like the following:
//
//   go build -ldflags "-X github.com/onflow/flow-go/version.semver=v1.0.0"
package build

// Default value for build-time-injected version strings.
const undefined = "undefined"

// The following variables are injected at build-time using ldflags.
var (
	semver string
	commit string
)

// Semver returns the semantic version of this build.
func Semver() string {
	return semver
}

// Commit returns the commit at which this build was created.
func Commit() string {
	return commit
}

// IsDefined determines whether a version string is defined. Inputs should
// have been produced from this package.
func IsDefined(v string) bool {
	return v != undefined
}

// If any of the build-time-injected variables are empty at initialization,
// mark them as undefined.
func init() {
	if len(semver) == 0 {
		semver = undefined
	}
	if len(commit) == 0 {
		commit = undefined
	}
}
