package aggregate

import (
	"go/ast"
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

func dummyAggregator(t *testing.T, s string) *Aggregator {
	a := &Aggregator{}
	f, err := a.parseFile(s)
	assert.NoError(t, err)
	a.main = f
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
		{"foo", "solve", ErrInvalidFile},
	}
	for _, tt := range tests {
		a := &Aggregator{}
		p, err := a.parseFile(tt.name)
		if err != nil {
			assert.Equal(t, tt.expectErr, errors.Cause(err))
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
		f, err := a.parseFile(tt.depend)
		assert.NoError(t, err)

		addDependPrefix(f)

		// TODO: check node
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
			pkg, err := a.parseFile(d)
			assert.NoError(t, err)
			a.depends = append(a.depends, pkg)
		}

		_, err := renameDependPackage(a.main, a.depends)
		assert.NoError(t, err)

		// TODO: check return value
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
		pkgs := make([]*File, 0, len(tt.files))
		for _, f := range tt.files {
			a := Aggregator{}
			p, err := a.parseFile(f)
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
				"testdata/util.go",
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
		_, err := New().Invoke(tt.main, tt.deps)
		assert.NoError(t, err)

		// TODO: check return value
	}
}
