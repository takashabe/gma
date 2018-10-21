package aggregate

import (
	"bytes"
	"fmt"
	"go/printer"
	"go/token"
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

func dummyAggregator(t *testing.T, f string) *Aggregator {
	a := &Aggregator{}
	pkg, err := a.parsePackage([]string{f})
	assert.NoError(t, err)
	assert.Len(t, pkg.files, 1)
	a.main = pkg
	return a
}

func TestParsePackage(t *testing.T) {
	tests := []struct {
		names     []string
		expectErr error
	}{
		{[]string{"testdata/test.go"}, nil},
		{[]string{"testdata/test.go", "testdata/util.go"}, nil},
		{[]string{""}, errors.New("Not found correctly go files")},
	}
	for _, tt := range tests {
		aggregator := &Aggregator{}
		_, err := aggregator.parsePackage(tt.names)
		if err != nil {
			assert.EqualError(t, err, tt.expectErr.Error())
			continue
		}
	}
}

func TestGetSolverNode(t *testing.T) {
	tests := []struct {
		name  string
		exist bool
	}{
		{"testdata/test.go", true},
		{"testdata/util.go", false},
	}
	for _, tt := range tests {
		a := dummyAggregator(t, tt.name)

		_, ok := a.getSolverNode()
		assert.Equal(t, tt.exist, ok)
	}
}

func TestInjectMain(t *testing.T) {
	tests := []struct {
		name string
	}{
		{"testdata/test.go"},
	}
	for _, tt := range tests {
		a := dummyAggregator(t, tt.name)
		err := a.inejctMain()
		assert.NoError(t, err)

		var buf bytes.Buffer
		p := printer.Config{Mode: printer.TabIndent, Tabwidth: 4}
		p.Fprint(&buf, token.NewFileSet(), a.main.files[0])

		// TODO: Compare expect code string. Probably use comparing with AST converted string.
		fmt.Println(buf.String())
	}
}
