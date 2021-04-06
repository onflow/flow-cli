package contracts

import (
	"fmt"
	"path"
	"path/filepath"
	"strings"

	"github.com/onflow/cadence/runtime/ast"
	"github.com/onflow/cadence/runtime/common"
	"github.com/onflow/cadence/runtime/parser2"
	"github.com/onflow/flow-cli/pkg/flowcli/project"
)

type Resolver struct {
	code    string
	program *ast.Program
}

func NewResolver(code []byte) (*Resolver, error) {
	program, err := parser2.ParseProgram(string(code))
	if err != nil {
		return nil, err
	}

	return &Resolver{
		code:    string(code),
		program: program,
	}, nil
}

func (r *Resolver) ResolveImports(
	contracts []project.Contract,
	aliases project.Aliases,
) string {
	imports := r.parseImports()
	sourceTarget := r.getSourceTarget(contracts, aliases)

	fmt.Println(absolutePath("./tests/Foo.cdc", "./Foo.cdc"), path.Clean("./tests/Foo.cdc"))
	fmt.Println(filepath.Abs("./Foo.cdc"))
	fmt.Println(filepath.Abs("./tests/Foo.cdc"))

	for _, imp := range imports {
		fmt.Println(absolutePath(imp, sourceTarget[imp]))
		fmt.Println(imp, sourceTarget[imp], sourceTarget)

		if sourceTarget[imp] != "" {
			r.code = r.replaceImport(imp, sourceTarget[imp])
		}
	}

	return r.code
}

func (r *Resolver) replaceImport(from string, to string) string {
	return strings.Replace(
		r.code,
		fmt.Sprintf(`"%s"`, from),
		fmt.Sprintf("0x%s", to),
		1,
	)
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
