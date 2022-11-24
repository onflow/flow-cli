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

func (p *Program) HasImports() bool {
	return len(p.Imports()) > 0
}

func (p *Program) ReplaceImport(from string, to string) *Program {
	p.code = []byte(strings.Replace(
		string(p.code),
		fmt.Sprintf(`"%s"`, from),
		fmt.Sprintf("0x%s", to),
		1,
	))

	p.reload()
	return p
}

func (p *Program) Code() []byte {
	return p.code
}

func (p *Program) Name() (string, error) {
	if len(p.astProgram.CompositeDeclarations())+len(p.astProgram.InterfaceDeclarations()) != 1 {
		return "", fmt.Errorf("the code must declare exactly one contract or contract interface")
	}

	for _, compositeDeclaration := range p.astProgram.CompositeDeclarations() {
		if compositeDeclaration.CompositeKind == common.CompositeKindContract {
			return compositeDeclaration.Identifier.Identifier, nil
		}
	}

	for _, interfaceDeclaration := range p.astProgram.InterfaceDeclarations() {
		if interfaceDeclaration.CompositeKind == common.CompositeKindContract {
			return interfaceDeclaration.Identifier.Identifier, nil
		}
	}

	return "", fmt.Errorf("unable to determine contract name")
}

func (p *Program) reload() {
	astProgram, err := parser.ParseProgram(p.code, nil)
	if err != nil {
		return
	}

	p.astProgram = astProgram
}
