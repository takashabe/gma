package aggregate

import (
	"go/ast"
	"go/printer"
	"go/token"
	"os"
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

func dummyAggregator(t *testing.T, f string) *Aggregator {
	a := &Aggregator{}
	pkg, err := a.parsePackage(f)
	assert.NoError(t, err)
	a.main = pkg
	return a
}

func TestParsePackage(t *testing.T) {
	tests := []struct {
		name          string
		expectPkgName string
		expectErr     error
	}{
		{"testdata/test.go", "main", nil},
		{"testdata/util2/util2.go", "util2", nil},
		{"foo", "solve", errors.New("not exists .go file foo")},
	}
	for _, tt := range tests {
		a := &Aggregator{}
		p, err := a.parsePackage(tt.name)
		if err != nil {
			assert.EqualError(t, err, tt.expectErr.Error())
			continue
		}
		assert.Equal(t, tt.expectPkgName, p.name)
	}
}

func TestAddDependPrefix(t *testing.T) {
	tests := []struct {
		depend string
	}{
		{"testdata/test2.go"},
		{"testdata/util.go"},
		{"testdata/util2/util2.go"},
	}
	for _, tt := range tests {
		a := Aggregator{}
		pkg, err := a.parsePackage(tt.depend)
		assert.NoError(t, err)

		addDependPrefix(pkg)
	}
}

func TestReplaceFuncs(t *testing.T) {
	tests := []struct {
		main string
		deps []string
	}{
		{
			main: "testdata/test.go",
			deps: []string{
				"testdata/util2/util2.go",
			},
		},
	}
	for _, tt := range tests {
		a := dummyAggregator(t, tt.main)

		for _, d := range tt.deps {
			pkg, err := a.parsePackage(d)
			assert.NoError(t, err)
			a.depends = append(a.depends, pkg)
		}

		n := a.replaceUtilFuncs()
		printer.Fprint(os.Stdout, token.NewFileSet(), n)
	}
}

func TestMergeFiles(t *testing.T) {
	tests := []struct {
		files []string
	}{
		{
			[]string{
				"testdata/util2/util.go",
				"testdata/util2/util2.go",
			},
		},
		{
			[]string{
				"testdata/util.go",
				"testdata/util2/util2.go",
			},
		},
	}
	for _, tt := range tests {
		pkgs := make([]*Package, 0, len(tt.files))
		for _, f := range tt.files {
			a := Aggregator{}
			p, err := a.parsePackage(f)
			assert.NoError(t, err)

			pkgs = append(pkgs, p)
		}

		files := make([]*ast.File, 0, len(pkgs))
		for _, pkg := range pkgs {
			files = append(files, pkg.file)
		}

		_, err := mergeFiles(files)
		assert.NoError(t, err)
	}
}

func TestAggregate(t *testing.T) {
	tests := []struct {
		main string
		deps []string
	}{
		{
			"testdata/test.go",
			[]string{
				"testdata/util2/util.go",
				"testdata/util2/util2.go",
			},
		},
		{
			"testdata/test2.go",
			[]string{
				"testdata/util2/util2.go",
			},
		},
	}
	for _, tt := range tests {
		actual, err := New().Invoke(tt.main, tt.deps)
		assert.NoError(t, err)

		printer.Fprint(os.Stdout, token.NewFileSet(), actual)
	}
}
