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

	// TODO: define codes.
	ExitCodeError
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
	if err := c.parseArgs(args[2:], p); err != nil {
		return ExitCodeError
	}

	ret, err := aggregate.Aggregator(p.main, p.depends)
	if err != nil {
		return ExitCodeError
	}
	if err := aggregate.Fprint(c.OutStream, ret); err != nil {
		return ExitCodeError
	}

	return ExitCodeOK
}

func (c *CLI) parseArgs(args []string, p *param) error {
	flags := flag.NewFlagSet("param", flag.ContinueOnError)
	flags.SetOutput(c.ErrStream)

	flags.StringVar(&p.main, "main", "", "main file")
	flags.Var(&p.depends, "depends", "", "depend files.")

	err := flags.Parse(args)
	if err != nil {
		return errors.Wrapf(err, "failed to parsed args")
	}
	return nil
}

func (c *CLI) error(format string, args ...interface{}) {
	//todo: wrapped fmt.Fprintf()
}
