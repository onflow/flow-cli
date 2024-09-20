package generator

import (
	"path/filepath"

	"github.com/onflow/flow-cli/internal/util"
	"github.com/onflow/flowkit/v2"
)

// ScriptTemplate contains only a name property for scripts and transactions
type ScriptTemplate struct {
	Name         string
	TemplatePath string
	Data         map[string]interface{}
}

var _ TemplateItem = ScriptTemplate{}

// GetName returns the name of the script or transaction
func (o ScriptTemplate) GetName() string {
	return o.Name
}

// GetTemplate returns an empty string for scripts and transactions
func (o ScriptTemplate) GetTemplatePath() string {
	if o.TemplatePath == "" {
		return "script_init.cdc.tmpl"
	}

	return o.TemplatePath
}

// GetData returns the data of the script or transaction
func (o ScriptTemplate) GetData() map[string]interface{} {
	return o.Data
}

func (o ScriptTemplate) GetTargetPath() string {
	return filepath.Join(DefaultCadenceDirectory, "scripts", util.AddCDCExtension(o.Name))
}

func (o ScriptTemplate) UpdateState(state *flowkit.State) error {
	return nil
}
