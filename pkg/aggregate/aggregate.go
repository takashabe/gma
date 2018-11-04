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
	files   *ast.File
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
		files:   af,
		name:    af.Name.Name,
		imports: imports,
	}, nil
}

func (a Aggregator) getSolverNode() (*ast.Object, bool) {
	var obj *ast.Object
	ast.Inspect(a.main.files, func(n ast.Node) bool {
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
			a.main.files.Decls = append(a.main.files.Decls, fn)
			return nil
		}
	}
	return errors.New("failed to inject main method")
}

func mergeFiles(ns []ast.Node) ast.Node {
	pre := func(c *astutil.Cursor) bool {
		n := c.Node()
		g, ok := n.(*ast.GenDecl)
		if !ok {
			return true
		}
		pp.Println(g)
		return true
	}

	for _, n := range ns {
		// var buf bytes.Buffer
		// printer.Fprint(&buf, fset, d.files)

		a := astutil.Apply(n, pre, nil)
		pp.Println(a)
	}
	return nil
}

func addDependPrefix(pkg *Package) ast.Node {
	prefix := fmt.Sprintf("_%s", strings.ToLower(pkg.name))
	ast.Inspect(pkg.files, func(n ast.Node) bool {
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
	return pkg.files
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

	return astutil.Apply(a.main.files, pre, nil)
}
