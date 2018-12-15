package util2

import (
	"fmt"
	"os"
)

func Foo() error {
	bar()
	return nil
}

func Echo(s string) string {
	fmt.Println(s)
	return s
}

func bar() {
	fmt.Fprintf(os.Stdout, "bar_string")
}

func foo() {
	fmt.Println("foo_string")
}

const Const = "go"
