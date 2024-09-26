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
	Data         map[string]interface{}
}

var _ TemplateItem = TransactionTemplate{}

// GetName returns the name of the script or transaction
func (o TransactionTemplate) GetType() string {
	return "transaction"
}

// GetTemplate returns an empty string for scripts and transactions
func (o TransactionTemplate) GetTemplatePath() string {
	if o.TemplatePath == "" {
		return "transaction_init.cdc.tmpl"
	}

	return o.TemplatePath
}

// GetData returns the data of the script or transaction
func (o TransactionTemplate) GetData() map[string]interface{} {
	return o.Data
}

func (o TransactionTemplate) GetTargetPath() string {
	return filepath.Join(DefaultCadenceDirectory, DefaultTransactionDirectory, util.AddCDCExtension(o.Name))
}
