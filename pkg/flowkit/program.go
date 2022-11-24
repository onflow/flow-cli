package flowkit

import (
	"fmt"
	"github.com/onflow/cadence/runtime/ast"
	"github.com/onflow/cadence/runtime/common"
	"github.com/onflow/cadence/runtime/parser"
	"strings"
)

type Program struct {
	code       []byte
	astProgram *ast.Program
}

func NewProgram(code []byte) (*Program, error) {
	astProgram, err := parser.ParseProgram(code, nil)
	if err != nil {
		return nil, err
	}

	return &Program{
		code:       code,
		astProgram: astProgram,
	}, nil
}

func (p *Program) Imports() []string {
	imports := make([]string, 0)

	for _, importDeclaration := range p.astProgram.ImportDeclarations() {
		_, isFileImport := importDeclaration.Location.(common.StringLocation)

		if isFileImport {
			imports = append(imports, importDeclaration.Location.String())
		}
	}

	return imports
}

func (p *Program) hasImports() bool {
	return len(p.Imports()) > 0
}

func (p *Program) reload() {
	astProgram, err := parser.ParseProgram(p.code, nil)
	if err != nil {
		return
	}

	p.astProgram = astProgram
}

func (p *Program) replaceImport(from string, to string) {
	p.code = []byte(strings.Replace(
		string(p.code),
		fmt.Sprintf(`"%s"`, from),
		fmt.Sprintf("0x%s", to),
		1,
	))

	p.reload()
}
