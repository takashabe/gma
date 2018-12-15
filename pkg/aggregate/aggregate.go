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

// Aggregator provides aggregate for main and depend file.
type Aggregator struct {
	main    *File
	depends []*File
}

// File represent package and files
type File struct {
	file    *ast.File
	name    string
	imports []string
}

var (
	fset = token.NewFileSet()
)

// New returns initialized Aggregator.
func New() *Aggregator {
	return &Aggregator{}
}

// Invoke aggregate main and depend files.
func (a Aggregator) Invoke(main string, depends []string) (*ast.File, error) {
	mp, err := a.parseFile(main)
	if err != nil {
		return nil, err
	}
	a.main = mp

	for _, dep := range depends {
		dp, err := a.parseFile(dep)
		if err != nil {
			return nil, err
		}
		a.depends = append(a.depends, dp)
	}

	n, err := renameDependPackage(a.main, a.depends)
	if err != nil {
		return nil, err
	}

	f, ok := n.(*ast.File)
	if !ok {
		return nil, errors.Wrapf(ErrInvalidFile, f.Name.Name)
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

func (a Aggregator) parseFile(name string) (*File, error) {
	if !strings.HasSuffix(name, "go") {
		return nil, errors.Wrapf(ErrInvalidFile, "%s", name)
	}
	af, err := parser.ParseFile(fset, name, nil, parser.ParseComments)
	if err != nil {
		return nil, err
	}

	imports := make([]string, 0, len(af.Imports))
	for _, i := range af.Imports {
		imports = append(imports, i.Path.Value)
	}
	return &File{
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

	return restructureImport(file)
}

func restructureImport(n ast.Node) (*ast.File, error) {
	var (
		buf  = bytes.Buffer{}
		fset = token.NewFileSet()
	)
	printer.Fprint(&buf, fset, n)
	a, err := imports.Process("", buf.Bytes(), nil)
	if err != nil {
		return nil, err
	}

	return parser.ParseFile(fset, "", a, parser.AllErrors)
}

func addDependPrefix(file *File) (ast.Node, map[string]string) {
	replaceFuncs := make(map[string]string)

	prefix := fmt.Sprintf("_%s", strings.ToLower(file.name))
	ast.Inspect(file.file, func(n ast.Node) bool {
		switch t := n.(type) {
		case *ast.FuncDecl:
			origin := fmt.Sprintf("%s.%s", file.name, t.Name.Name)
			replaced := fmt.Sprintf("%s_%s", prefix, t.Name.Name)

			replaceFuncs[origin] = replaced
			t.Name.Name = replaced
		case *ast.CallExpr:
			id, ok := t.Fun.(*ast.Ident)
			if !ok {
				return true
			}
			origin := fmt.Sprintf("%s.%s", file.name, id.Name)
			replaced := fmt.Sprintf("%s_%s", prefix, id.Name)

			replaceFuncs[replaced] = origin
			id.Name = replaced
		}
		return true
	})
	return file.file, replaceFuncs
}

func renameDependPackage(main *File, depends []*File) (ast.Node, error) {
	replaceFuncs := make(map[string]string)
	replacePkgs := []string{}
	for _, p := range depends {
		if main.name != p.name {
			replacePkgs = append(replacePkgs, p.name)

			_, fs := addDependPrefix(p)
			for origin, replaced := range fs {
				replaceFuncs[origin] = replaced
			}
		}
	}

	files := []*ast.File{main.file}
	for _, df := range depends {
		files = append(files, df.file)
	}
	mf, err := mergeFiles(files)
	if err != nil {
		return nil, err
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
	return restructureImport(ret)
}
