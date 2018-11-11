package aggregate

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"strings"

	"golang.org/x/tools/go/ast/astutil"

	"github.com/k0kubun/pp"
	"github.com/pkg/errors"
)

// Aggregator provides aggregate solver
type Aggregator struct {
	solver  string
	main    *Package
	depends []*Package
}

// Package represent package and files
type Package struct {
	file    *ast.File
	name    string
	imports []string
}

var (
	fset = token.NewFileSet()
)

// Aggregate aggregate main and sub files.
func (a Aggregator) Aggregate(mainFile string, subFiles []string) error {
	// TODO: implements
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

	imports := []*ast.ImportSpec{}
	seen := make(map[string]struct{})
	for _, file := range files {
		for _, imp := range file.Imports {
			p := imp.Path.Value
			if _, ok := seen[p]; ok {
				continue
			}
			imports = append(imports, imp)
			seen[p] = struct{}{}
		}
	}

	// Collect decls
	decls := []ast.Decl{}
	for _, file := range files {
		for _, d := range file.Decls {
			g, ok := d.(*ast.GenDecl)
			// TODO: All import decls through for debug.
			if ok && g.Tok == token.IMPORT {
				pp.Println(g)
				continue
			}
			decls = append(decls, d)
		}
	}

	// Make GenDecl for imports
	if len(imports) > 0 {
		imps := make([]ast.Spec, 0, len(imports))
		for _, i := range imports {
			imps = append(imps, i)
		}

		g := &ast.GenDecl{
			TokPos: files[0].Package,
			Tok:    token.IMPORT,
			Specs:  imps,
		}
		decls = append(decls, g)
	}

	// TODO(takashabe): Breaked when has the multi package files. Add parameter a primary package.
	pos := files[0].Package
	scope := files[0].Scope
	name := files[0].Name

	file := &ast.File{
		Package: pos,
		Name:    name,
		Decls:   decls,
		Imports: imports,
		Scope:   scope,
	}

	return file, nil
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

		if call, ok := n.(*ast.CallExpr); ok {
			pp.Println(call)
			return true
		}

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
		return true
	}

	return astutil.Apply(a.main.file, pre, nil)
}

func funcName(f *ast.FuncDecl) string {
	receiver := f.Recv
	if receiver == nil || len(receiver.List) != 1 {
		return f.Name.Name
	}

	t := receiver.List[0].Type
	if p, _ := t.(*ast.StarExpr); p != nil {
		t = p.X
	}

	// reciever name + func name
	if p, _ := t.(*ast.Ident); p != nil {
		return p.Name + "." + f.Name.Name
	}

	return f.Name.Name
}
