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
	DefaultTransactionDirectory = "transactions"
)

// TransactionTemplate contains only a name property for scripts and transactions
type TransactionTemplate struct {
	Name         string
	TemplatePath string
	Data         map[string]any
}

var _ TemplateItem = TransactionTemplate{}

func (o TransactionTemplate) GetType() string {
	return "transaction"
}

// GetTemplatePath returns an empty string for scripts and transactions
func (o TransactionTemplate) GetTemplatePath() string {
	if o.TemplatePath == "" {
		return "transaction_init.cdc.tmpl"
	}

	return o.TemplatePath
}

// GetData returns the data of the script or transaction
func (o TransactionTemplate) GetData() map[string]any {
	return o.Data
}

func (o TransactionTemplate) GetTargetPath() string {
	return filepath.Join(DefaultCadenceDirectory, DefaultTransactionDirectory, util.AddCDCExtension(o.Name))
}
