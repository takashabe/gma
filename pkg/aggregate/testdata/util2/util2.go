package util2

import "fmt"

func Foo() error {
	bar()
	return nil
}

func Echo(s string) string {
	fmt.Println(s)
	return s
}

func bar() {
	fmt.Println("bar_string")
}

func foo() {
	fmt.Println("foo_string")
}
