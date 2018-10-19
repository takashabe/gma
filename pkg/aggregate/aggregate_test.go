package aggregate

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParsePackage(t *testing.T) {
	tests := []struct {
		names     []string
		expectErr error
	}{
		{[]string{"testdata/test.go"}, nil},
		// {[]string{"testdata/test.go", "testdata/util.go"}, nil},
		// {[]string{""}, errors.New("Not found correctly go files")},
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

func TestDetectSolver(t *testing.T) {
	tests := []struct {
		name  string
		exist bool
	}{
		{"testdata/test.go", true},
		{"testdata/util.go", false},
	}
	for _, tt := range tests {
		a := &Aggregator{}
		pkg, err := a.parsePackage([]string{tt.name})
		assert.NoError(t, err)
		assert.Len(t, pkg.files, 1)

		a.main = pkg
		_, ok := a.detectSolver()
		assert.Equal(t, tt.exist, ok)
	}
}

func TestReplaceUtilFuncs(t *testing.T) {
	tests := []struct {
		name  string
		exist bool
	}{
		{"testdata/test.go", true},
		{"testdata/util.go", false},
		{"testdata/util2/util2.go", false},
	}
	for _, tt := range tests {
		a := &Aggregator{}
		pkg, err := a.parsePackage([]string{tt.name})
		assert.NoError(t, err)
		assert.Len(t, pkg.files, 1)
	}
}
