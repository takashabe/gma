package util2

import "fmt"

func Bar() error {
	fmt.Println("bar")
	bar()
	return nil
}
