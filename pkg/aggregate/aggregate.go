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

	"github.com/pkg/errors"
	"golang.org/x/tools/go/ast/astutil"
	"golang.org/x/tools/imports"
)

// Aggregator provides aggregate solver
type Aggregator struct {
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

	n := a.replaceUtilFuncs()
	f, ok := n.(*ast.File)
	if !ok {
		return nil, errors.New("invalid depends files")
	}
	return f, nil
}

// Fprint wrapped printer.Print
func Fprint(w io.Writer, f *ast.File) error {
	config := printer.Config{
		Mode:     printer.TabIndent,
		Tabwidth: 8,
		Indent:   0,
	}
	config.Fprint(w, token.NewFileSet(), f)
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

func addDependPrefix(pkg *Package) (ast.Node, map[string]string) {
	replaceFuncs := make(map[string]string)

	prefix := fmt.Sprintf("_%s", strings.ToLower(pkg.name))
	ast.Inspect(pkg.file, func(n ast.Node) bool {
		switch t := n.(type) {
		case *ast.FuncDecl:
			origin := fmt.Sprintf("%s.%s", pkg.name, t.Name.Name)
			replaced := fmt.Sprintf("%s_%s", prefix, t.Name.Name)

			replaceFuncs[origin] = replaced
			t.Name.Name = replaced
		case *ast.CallExpr:
			id, ok := t.Fun.(*ast.Ident)
			if !ok {
				return true
			}
			origin := fmt.Sprintf("%s.%s", pkg.name, id.Name)
			replaced := fmt.Sprintf("%s_%s", prefix, id.Name)

			replaceFuncs[replaced] = origin
			id.Name = replaced
		}
		return true
	})
	return pkg.file, replaceFuncs
}

// TODO: Rename function when import another util packages.
func (a Aggregator) replaceUtilFuncs() ast.Node {
	// collect util package and method list
	replaceFuncs := make(map[string]string)
	replacePkgs := []string{}
	for _, p := range a.depends {
		if a.main.name != p.name {
			replacePkgs = append(replacePkgs, p.name)

			_, fs := addDependPrefix(p)
			for origin, replaced := range fs {
				replaceFuncs[origin] = replaced
			}
		}
	}

	files := []*ast.File{a.main.file}
	for _, df := range a.depends {
		files = append(files, df.file)
	}
	mf, err := mergeFiles(files)
	if err != nil {
		panic("failed to mergeFiles")
	}

	replaceFuncNodes := make(map[string]*ast.FuncDecl)
	ast.Inspect(mf, func(n ast.Node) bool {
		fn, ok := n.(*ast.FuncDecl)
		if !ok {
			return true
		}

		for origin, replaced := range replaceFuncs {
			if fn.Name.Name == replaced {
				replaceFuncNodes[origin] = fn
				return true
			}
		}
		return true
	})

	pre := func(c *astutil.Cursor) bool {
		n := c.Node()

		callExpr, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}
		selector, ok := callExpr.Fun.(*ast.SelectorExpr)
		if !ok {
			return true
		}
		x, ok := selector.X.(*ast.Ident)
		if !ok {
			return true
		}

		repOk := false
		for _, rp := range replacePkgs {
			if rp == x.Name {
				repOk = true
				break
			}
		}
		if !repOk {
			return true
		}

		fn := fmt.Sprintf("%s.%s", x.Name, selector.Sel.Name)
		repNode, ok := replaceFuncNodes[fn]
		if !ok {
			return false
		}

		cn := &ast.CallExpr{
			Fun:  repNode.Name,
			Args: callExpr.Args,
		}
		c.Replace(cn)
		return true
	}

	ret := astutil.Apply(mf, pre, nil)
	return ret
}
