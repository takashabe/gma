package util2

import "fmt"

func Foo() error {
	bar()
	return nil
}

func bar() {
	fmt.Println("bar")
}

func foo() {
	fmt.Println("foo")
}
