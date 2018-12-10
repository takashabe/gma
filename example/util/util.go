package util

import "fmt"

type P struct {
	Name string
	Age  int
}

func Person(name string, age int) P {
	return P{
		Name: name,
		Age:  age,
	}
}

func (p P) Say() {
	fmt.Printf("%s %d\n", p.Name, p.Age)
}
