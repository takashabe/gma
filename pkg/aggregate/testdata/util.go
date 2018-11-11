package solve

import "fmt"

func Foo() error {
	_, err := fmt.Println("solve.Foo")
	return err
}
