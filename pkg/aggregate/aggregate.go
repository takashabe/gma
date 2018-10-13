package aggregate

import (
	"go/ast"
	"go/parser"
	"go/token"
	"strings"

	"github.com/pkg/errors"
)

// Aggregator provides aggregate solver
type Aggregator struct {
}

func (Aggregator) parseFiles(names []string) ([]*ast.File, error) {
	var astFiles []*ast.File

	fs := token.NewFileSet()
	for _, name := range names {
		if !strings.HasSuffix(name, "go") {
			continue
		}
		af, err := parser.ParseFile(fs, name, nil, parser.ParseComments)
		if err != nil {
			return nil, err
		}
		astFiles = append(astFiles, af)
	}
	if len(astFiles) == 0 {
		return nil, errors.New("Not found correctly go files")
	}
	return astFiles, nil
}

func (Aggregator) walk(f *ast.File) {
	ast.Inspect(f, func(n ast.Node) bool {
		// TODO: Rename function when import another util packages.

		// TODO: Detect implements Solver() struct.
	})
}
