package main

import (
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/pkg/errors"
	"github.com/takashabe/go-main-aggregator/pkg/aggregate"
)

func main() {
	c := &CLI{
		OutStream: os.Stdout,
		ErrStream: os.Stderr,
	}
	os.Exit(c.Run(os.Args))
}

// Exit codes.
const (
	ExitCodeOK = iota

	// TODO: define specific error codes.
	ExitCodeError
	ExitCodeParseError
)

type param struct {
	main    string
	depends depends
}

type depends []string

func (d *depends) String() string {
	return fmt.Sprint(*d)
}

func (d *depends) Set(s string) error {
	*d = append(*d, s)
	return nil
}

// CLI represent cli interface.
type CLI struct {
	OutStream io.Writer
	ErrStream io.Writer
}

// Run invokes the CLI with the given arguments.
func (c *CLI) Run(args []string) int {
	p := &param{}
	if err := c.parseArgs(args[1:], p); err != nil {
		c.printError("failed to parse args. %v\n", err)
		return ExitCodeParseError
	}

	a := aggregate.New()
	ret, err := a.Invoke(p.main, p.depends)
	if err != nil {
		c.printError("failed to aggregate process. %v\n", err)
		return ExitCodeError
	}
	if err := aggregate.Fprint(c.OutStream, ret); err != nil {
		c.printError("failed to print aggregated file. %v\n", err)
		return ExitCodeError
	}

	return ExitCodeOK
}

func (c *CLI) parseArgs(args []string, p *param) error {
	flags := flag.NewFlagSet("param", flag.ContinueOnError)
	flags.SetOutput(c.ErrStream)

	flags.StringVar(&p.main, "main", "", "require: main file")
	flags.Var(&p.depends, "depends", "optional: depend files")

	err := flags.Parse(args)
	if err != nil {
		return errors.Wrapf(err, "failed to parsed args")
	}
	return nil
}

func (c *CLI) printError(format string, args ...interface{}) {
	fmt.Fprintf(c.ErrStream, format, args...)
}
