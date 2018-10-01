package aggregate

import (
	"go/ast"
	"go/parser"
	"go/token"
	"strings"
)

// Aggregator provide
type Aggregator struct{}

type Package struct {
	main *ast.File
	subs []*ast.File
}

func (a *Aggregator) parsePackage(names []string) error {
	var astFiles []*ast.File

	fs := token.NewFileSet()
	for _, name := range names {
		if !strings.HasSuffix(name, "go") {
			continue
		}
		af, err := parser.ParseFile(fs, name, nil, parser.ParseComments)
		if err != nil {
			return err
		}
		astFiles = append(astFiles, af)
	}
	return nil
}
