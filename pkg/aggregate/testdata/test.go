package solve

import (
	"fmt"

	"./util2"
)

type P001 struct{}

func (o *P001) Solve() {
	fmt.Println("main")
}

func (p *P001) callExportFunc() {
	util2.Foo()
}
