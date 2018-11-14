package aggregate

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
	"io"
	"strings"

	"golang.org/x/tools/go/ast/astutil"

	"github.com/pkg/errors"
	"golang.org/x/tools/imports"
)

// Aggregator provides aggregate solver
type Aggregator struct {
	solver  string
	main    *Package
	depends []*Package
}

// Package represent package and files
// TODO: Package rename to `File`, and file field to `ast`
type Package struct {
	file    *ast.File
	name    string
	imports []string
}

var (
	fset = token.NewFileSet()
)

// Aggregate aggregate main and sub files.
func Aggregate(mainFile string, subFiles []string) (*ast.File, error) {
	a := Aggregator{}
	mp, err := a.parsePackage(mainFile)
	if err != nil {
		return nil, err
	}
	a.main = mp

	for _, dep := range subFiles {
		dp, err := a.parsePackage(dep)
		if err != nil {
			return nil, err
		}
		a.depends = append(a.depends, dp)
	}

	// todo: impl
	if err := a.inejctMain(); err != nil {
		return nil, err
	}
	n := a.replaceUtilFuncs()
	f, ok := n.(*ast.File)
	if !ok {
		return nil, errors.New("invalid depends files")
	}
	a.main.file = f

	files := make([]*ast.File, 0)
	files = append(files, a.main.file)
	for _, d := range a.depends {
		files = append(files, d.file)
	}
	return mergeFiles(files)
}

// Fprint print Aggregator hold main file.
// Todo: Add print mode arg support.
func (a Aggregator) Fprint(w io.Writer, _ int) error {
	return nil
}

func (a Aggregator) parsePackage(name string) (*Package, error) {
	if !strings.HasSuffix(name, "go") {
		return nil, errors.Errorf("not exists .go file %s", name)
	}
	af, err := parser.ParseFile(fset, name, nil, parser.ParseComments)
	if err != nil {
		return nil, err
	}

	imports := make([]string, 0, len(af.Imports))
	for _, i := range af.Imports {
		imports = append(imports, i.Path.Value)
	}
	return &Package{
		file:    af,
		name:    af.Name.Name,
		imports: imports,
	}, nil
}

func (a Aggregator) getSolverNode() (*ast.Object, bool) {
	var obj *ast.Object
	ast.Inspect(a.main.file, func(n ast.Node) bool {
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

		switch t := fd.Recv.List[0].Type.(type) {
		case *ast.StarExpr:
			id, ok := t.X.(*ast.Ident)
			if !ok {
				return true
			}
			obj = id.Obj
		case *ast.Ident:
			obj = t.Obj
		default:
			return true
		}

		a.main.name = fd.Name.Name
		return false
	})
	return obj, obj != nil
}

func (a Aggregator) inejctMain() error {
	node, ok := a.getSolverNode()
	if !ok {
		return errors.New("not exists Solver")
	}
	file, err := parser.ParseFile(fset, "main", templateMain(node.Name), parser.Mode(0))
	if err != nil {
		return err
	}

	for _, d := range file.Decls {
		fn, ok := d.(*ast.FuncDecl)
		if ok {
			a.main.file.Decls = append(a.main.file.Decls, fn)
			return nil
		}
	}
	return errors.New("failed to inject main method")
}

func mergeFiles(files []*ast.File) (*ast.File, error) {
	switch len(files) {
	case 1:
		return files[0], nil
	case 0:
		return nil, errors.New("not found merge files")
	}

	// Collect decls
	decls := []ast.Decl{}
	for _, file := range files {
		for _, d := range file.Decls {
			g, ok := d.(*ast.GenDecl)
			// TODO: All import decls through for debug.
			if ok && g.Tok == token.IMPORT {
				continue
			}
			decls = append(decls, d)
		}
	}

	// TODO(takashabe): Breaked when has the multi package files. Add parameter a primary package.
	pos := files[0].Package
	name := files[0].Name

	file := &ast.File{
		Package: pos,
		Name:    name,
		Decls:   decls,
	}

	var (
		buf  = bytes.Buffer{}
		fset = token.NewFileSet()
	)
	printer.Fprint(&buf, fset, file)
	a, err := imports.Process("", buf.Bytes(), nil)
	if err != nil {
		return nil, err
	}

	return parser.ParseFile(fset, "", a, parser.AllErrors)
}

func addDependPrefix(pkg *Package) ast.Node {
	prefix := fmt.Sprintf("_%s", strings.ToLower(pkg.name))
	ast.Inspect(pkg.file, func(n ast.Node) bool {
		switch t := n.(type) {
		case *ast.FuncDecl:
			t.Name.Name = fmt.Sprintf("%s_%s", prefix, t.Name.Name)
		case *ast.CallExpr:
			id, ok := t.Fun.(*ast.Ident)
			if !ok {
				return true
			}
			id.Name = fmt.Sprintf("%s_%s", prefix, id.Name)
		}
		return true
	})
	return pkg.file
}

func templateMain(solver string) string {
	return fmt.Sprintf("package main; func main() { s:=%s{};s.Solve() }", solver)
}

// TODO: Rename function when import another util packages.
func (a Aggregator) replaceUtilFuncs() ast.Node {
	// collect util package and method list
	replacePkgs := []string{}
	for _, p := range a.depends {
		if a.main.name != p.name {
			addDependPrefix(p)
			replacePkgs = append(replacePkgs, p.name)
		}
	}

	pre := func(c *astutil.Cursor) bool {
		n := c.Node()

		selector, ok := n.(*ast.SelectorExpr)
		if !ok {
			return true
		}
		ident, ok := selector.X.(*ast.Ident)
		if !ok {
			return true
		}

		repOk := false
		for _, rp := range replacePkgs {
			if rp == ident.Name {
				repOk = true
				break
			}
		}
		if !repOk {
			return true
		}

		// NOTE: ident.Name == call replace package name
		// TODO: Implements depend funcs rename in the main file
		return true
	}

	return astutil.Apply(a.main.file, pre, nil)
}
