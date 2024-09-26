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
