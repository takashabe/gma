package main

import (
	"flag"
	"fmt"
	"io"

	"github.com/pkg/errors"
)

func main() {
	fmt.Println("not yet implements...")
}

type param struct {
	main    string
	depends []string
}

// CLI represent cli interface.
type CLI struct {
	OutStream io.Writer
	ErrStream io.Writer
}

func (c *CLI) parseArgs(args []string, p *param) error {
	flags := flag.NewFlagSet("param", flag.ContinueOnError)
	flags.SetOutput(c.ErrStream)

	flags.StringVar(&p.main, "main", "", "main file")
	// TODO: Add slice args with Set()

	err := flags.Parse(args)
	if err != nil {
		return errors.Wrapf(err, "failed to parsed args")
	}
	return nil
}
