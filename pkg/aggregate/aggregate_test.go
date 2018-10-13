package aggregate

import (
	"testing"

	"github.com/k0kubun/pp"
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
		p, err := aggregator.parseFiles(tt.names)
		if err != nil {
			assert.EqualError(t, err, tt.expectErr.Error())
			continue
		}
		pp.Println(p)
	}
}
