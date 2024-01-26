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

package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDependencies_ByName(t *testing.T) {
	dependencies := Dependencies{
		Dependency{Name: "mydep"},
	}

	dep := dependencies.ByName("mydep")
	assert.NotNil(t, dep)
}

func TestDependencies_AddOrUpdate(t *testing.T) {
	dependencies := Dependencies{}
	dependencies.AddOrUpdate(Dependency{Name: "mydep"})

	assert.Len(t, dependencies, 1)

	dep := dependencies.ByName("mydep")
	assert.NotNil(t, dep)
}
