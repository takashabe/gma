package aggregate

import "testing"

func TestParsePackageWithText(t *testing.T) {
	tests := []struct {
		text string
	}{
		{"foo"},
		{"package main\nimport \"fmt\"\nfunc main() { fmt.Println(1) }"},
	}
	for _, tt := range tests {
		aggregator := Aggregator{}
		names := []string{"foo.go", "foo.md"}
		aggregator.parsePackage(names, tt.text)
	}
}

func TestParsePackageWithFile(t *testing.T) {
	tests := []struct {
		names []string
	}{
		{[]string{""}},
	}
	for _, tt := range tests {
		t.Errorf("not implements %v", tt.names)
	}
}
