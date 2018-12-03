package main

import (
	"fmt"

	"./util2"
)

func main() {
	solver := &P001{}
	solver.Solve()
}

type P001 struct{}

func (p *P001) Solve() {
	fmt.Println("main")
	localFn("main")
}

func (p *P001) callExportFunc() {
	util2.Foo()
}

func (p *P001) callExportFunc2() {
	localFn(util2.Echo(util2.Echo("p001")))
}

func localFn(_ string) {}
