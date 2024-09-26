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
	DefaultScriptDirectory = "scripts"
)

type ScriptTemplate struct {
	Name         string
	TemplatePath string
	Data         map[string]interface{}
}

var _ TemplateItem = ScriptTemplate{}

func (o ScriptTemplate) GetType() string {
	return "script"
}

func (o ScriptTemplate) GetTemplatePath() string {
	if o.TemplatePath == "" {
		return "script_init.cdc.tmpl"
	}

	return o.TemplatePath
}

func (o ScriptTemplate) GetData() map[string]interface{} {
	return o.Data
}

func (o ScriptTemplate) GetTargetPath() string {
	return filepath.Join(DefaultCadenceDirectory, DefaultScriptDirectory, util.AddCDCExtension(o.Name))
}
