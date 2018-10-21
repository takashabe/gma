package aggregate

import (
	"fmt"
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
	// TODO: implements
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

func (a Aggregator) getSolverNode() (*ast.Ident, bool) {
	var ident *ast.Ident
	ast.Inspect(a.main.files[0], func(n ast.Node) bool {
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

		// TODO: process otherwise
		ex, ok := fd.Recv.List[0].Type.(*ast.StarExpr)
		if !ok {
			return true
		}
		id, ok := ex.X.(*ast.Ident)
		if !ok {
			return true
		}

		ident = id
		a.main.name = fd.Name.Name
		return false
	})
	return ident, ident != nil
}

func (a Aggregator) inejctMain() error {
	node, ok := a.getSolverNode()
	if !ok {
		return errors.New("not exists Solver")
	}
	file, err := parser.ParseFile(token.NewFileSet(), "main", templateMain(node.Name), parser.Mode(0))
	if err != nil {
		return err
	}

	for _, d := range file.Decls {
		fn, ok := d.(*ast.FuncDecl)
		if ok {
			a.main.files[0].Decls = append(a.main.files[0].Decls, fn)
			return nil
		}
	}
	return errors.New("failed to inject main method")
}

func templateMain(solver string) string {
	return fmt.Sprintf("package main; func main() { s:=%s{};s.Solve() }", solver)
}

// TODO: Rename function when import another util packages.
func (a Aggregator) replaceUtilFuncs(files []*ast.File) {
	for _, f := range files {
		pp.Println(f.Name)
	}
}
