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

package generator

import (
	"path/filepath"

	"github.com/onflow/flow-cli/internal/util"
)

const (
	DefaultTestDirectory = "tests"
)

type TestTemplate struct {
	Name         string
	TemplatePath string
	Data         map[string]interface{}
}

var _ TemplateItem = TestTemplate{}

func (o TestTemplate) GetType() string {
	return "test"
}

// GetName returns the name of the script or transaction
func (o TestTemplate) GetName() string {
	return o.Name
}

// GetTemplate returns an empty string for scripts and transactions
func (o TestTemplate) GetTemplatePath() string {
	if o.TemplatePath == "" {
		return "empty_test.cdc.tmpl"
	}

	return o.TemplatePath
}

// GetData returns the data of the script or transaction
func (o TestTemplate) GetData() map[string]interface{} {
	return o.Data
}

func (o TestTemplate) GetTargetPath() string {
	return filepath.Join(DefaultCadenceDirectory, DefaultTestDirectory, util.AddCDCExtension(o.Name))
}
