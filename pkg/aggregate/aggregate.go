package aggregate

import (
	"go/ast"
	"go/parser"
	"go/token"
	"strings"

	"github.com/k0kubun/pp"
	"github.com/pkg/errors"
)

// Aggregator provides aggregate solver
type Aggregator struct {
	solver string
	main   *Package
	subs   []*Package
}

// Package represent package and files
type Package struct {
	files []*ast.File
	name  string
}

// Aggregate aggregate main and sub files.
func (a Aggregator) Aggregate(mainFile string, subFiles []string) error {
	return nil
}

func (Aggregator) parsePackage(names []string) (*Package, error) {
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
	return &Package{
		files: astFiles,
		name:  astFiles[0].Name.Name,
	}, nil
}

func (a Aggregator) detectSolver() (ast.Node, bool) {
	var expr ast.Expr
	ast.Inspect(a.main.files[0], func(n ast.Node) bool {
		// TODO: Detect implements Solver() struct.
		fd, ok := n.(*ast.FuncDecl)
		if !ok {
			return true
		}
		if fd.Name.Name != "Solve" {
			return true
		}
		if fd.Recv == nil {
			return true
		}

		expr = fd.Recv.List[0].Type
		a.main.name = fd.Name.Name
		return false
	})
	return expr, expr != nil
}

// TODO: Rename function when import another util packages.
func (a Aggregator) replaceUtilFuncs(files []*ast.File) {
	for _, f := range files {
		pp.Println(f.Name)
	}
}
