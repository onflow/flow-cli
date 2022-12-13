package project

import (
	"fmt"
	"github.com/onflow/cadence/runtime/ast"
	"github.com/onflow/cadence/runtime/common"
	"github.com/onflow/cadence/runtime/parser"
	"regexp"
)

type Program struct {
	script     Scripter
	astProgram *ast.Program
}

type Scripter interface {
	Code() []byte
	SetCode([]byte)
	Location() string
}

func NewProgram(script Scripter) (*Program, error) {
	astProgram, err := parser.ParseProgram(script.Code(), nil)
	if err != nil {
		return nil, err
	}

	return &Program{
		script:     script,
		astProgram: astProgram,
	}, nil
}

func (p *Program) Imports() []string {
	imports := make([]string, 0)

	for _, importDeclaration := range p.astProgram.ImportDeclarations() {
		_, isIdentifierImport := importDeclaration.Location.(common.IdentifierLocation)
		if isIdentifierImport {
			location := importDeclaration.Location.String()
			if location == "Crypto" { // skip core library
				continue
			}
			imports = append(imports, location)
		}

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
	code := string(p.Code())

	pathRegex := regexp.MustCompile(fmt.Sprintf(`import (\w+) from "%s"`, from))
	identifierRegex := regexp.MustCompile(fmt.Sprintf("import (%s)", from))

	replacement := fmt.Sprintf(`import $1 from 0x%s`, to)
	code = pathRegex.ReplaceAllString(code, replacement)
	code = identifierRegex.ReplaceAllString(code, replacement)

	p.script.SetCode([]byte(code))
	p.reload()
	return p
}

func (p *Program) Location() string {
	return p.script.Location()
}

func (p *Program) Code() []byte {
	return p.script.Code()
}

func (p *Program) Name() (string, error) {
	if len(p.astProgram.CompositeDeclarations()) > 1 || len(p.astProgram.InterfaceDeclarations()) > 1 ||
		len(p.astProgram.CompositeDeclarations())+len(p.astProgram.InterfaceDeclarations()) > 1 {
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
	astProgram, err := parser.ParseProgram(p.script.Code(), nil)
	if err != nil {
		return
	}

	p.astProgram = astProgram
}
