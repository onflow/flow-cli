package contracts

import (
	"fmt"
	"path"
	"strings"

	"github.com/onflow/cadence/runtime/ast"
	"github.com/onflow/cadence/runtime/common"
	"github.com/onflow/cadence/runtime/parser2"
	"github.com/onflow/flow-cli/pkg/flowcli/project"
)

type Resolver struct {
	code    []byte
	program *ast.Program
}

func NewResolver(code []byte) (*Resolver, error) {
	program, err := parser2.ParseProgram(string(code))
	if err != nil {
		return nil, err
	}

	return &Resolver{
		code:    code,
		program: program,
	}, nil
}

func (r *Resolver) ResolveImports(
	scriptPath string,
	contracts []project.Contract,
	aliases project.Aliases,
) ([]byte, error) {
	imports := r.parseImports()
	sourceTarget := r.getSourceTarget(contracts, aliases)

	for _, imp := range imports {
		target := sourceTarget[absolutePath(scriptPath, imp)]
		if target != "" {
			r.code = r.replaceImport(imp, target)
		} else {
			return nil, fmt.Errorf("import %s could not be resolved from the configuration", imp)
		}
	}

	return r.code, nil
}

func (r *Resolver) replaceImport(from string, to string) []byte {
	return []byte(strings.Replace(
		string(r.code),
		fmt.Sprintf(`"%s"`, from),
		fmt.Sprintf("0x%s", to),
		1,
	))
}

func (r *Resolver) getSourceTarget(
	contracts []project.Contract,
	aliases project.Aliases,
) map[string]string {
	sourceTarget := make(map[string]string, 0)
	for _, contract := range contracts {
		sourceTarget[path.Clean(contract.Source)] = contract.Target.String()
	}

	for source, target := range aliases {
		sourceTarget[source] = target
	}

	return sourceTarget
}

func (r *Resolver) ImportExists() bool {
	return len(r.parseImports()) > 0
}

func (r *Resolver) parseImports() []string {
	imports := make([]string, 0)

	for _, importDeclaration := range r.program.ImportDeclarations() {
		location, ok := importDeclaration.Location.(common.StringLocation)
		if ok {
			imports = append(imports, location.String())
		}
	}

	return imports
}
