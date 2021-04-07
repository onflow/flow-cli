package contracts

import (
	"fmt"
	"path"
	"strings"

	"github.com/onflow/flow-go-sdk"

	"github.com/onflow/cadence/runtime/ast"
	"github.com/onflow/cadence/runtime/common"
	"github.com/onflow/cadence/runtime/parser2"
	"github.com/onflow/flow-cli/pkg/flowcli/project"
)

// Resolver handles resolving imports in Cadence code
type Resolver struct {
	code    []byte
	program *ast.Program
}

// NewResolver creates a new resolver
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

// ResolveImports resolves imports in code to addresses
//
// resolving is done based on code file path and is resolved to
// addresses defined in configuration for contracts or their aliases
//
func (r *Resolver) ResolveImports(
	codePath string,
	contracts []project.Contract,
	aliases project.Aliases,
) ([]byte, error) {
	imports := r.parseImports()
	sourceTarget := r.getSourceTarget(contracts, aliases)

	for _, imp := range imports {
		target := sourceTarget[absolutePath(codePath, imp)]
		if target != "" {
			r.code = r.replaceImport(imp, target)
		} else {
			return nil, fmt.Errorf("import %s could not be resolved from the configuration", imp)
		}
	}

	return r.code, nil
}

// replaceImport replaces import from path to address
func (r *Resolver) replaceImport(from string, to string) []byte {
	return []byte(strings.Replace(
		string(r.code),
		fmt.Sprintf(`"%s"`, from),
		fmt.Sprintf("0x%s", to),
		1,
	))
}

// getSourceTarget return a map with contract paths as keys and addresses as values
func (r *Resolver) getSourceTarget(
	contracts []project.Contract,
	aliases project.Aliases,
) map[string]string {
	sourceTarget := make(map[string]string, 0)
	for _, contract := range contracts {
		sourceTarget[path.Clean(contract.Source)] = contract.Target.String()
	}

	for source, target := range aliases {
		sourceTarget[path.Clean(source)] = target
	}

	return sourceTarget
}

// ImportExists checks if there is an import statement present in Cadence code
func (r *Resolver) ImportExists() bool {
	return len(r.parseImports()) > 0
}

// parseImports returns all imports from Cadence code as an array
func (r *Resolver) parseImports() []string {
	imports := make([]string, 0)

	for _, importDeclaration := range r.program.ImportDeclarations() {
		location, ok := importDeclaration.Location.(common.StringLocation)
		if ok && flow.HexToAddress(location.String()) == flow.EmptyAddress {
			imports = append(imports, location.String())
		}
	}

	return imports
}
