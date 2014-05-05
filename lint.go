package main

import (
	"flag"
	"io"
	"log"
	"os"

	"github.com/cppforlife/lint/linter"
)

var (
	debugOpt = flag.Bool("debug", false, "show debugging information")
	fixOpt   = flag.Bool("fix", false, "fix problems that can be fixed automatically")
)

func main() {
	flag.Parse()

	logger := buildLogger(*debugOpt)

	ui := linter.NewPlainUI(os.Stdout, logger)

	loader, err := linter.NewLoaderFromArgs(os.Getenv("GOPATH"), flag.Args(), logger)
	if err != nil {
		ui.DisplayError(err)
		os.Exit(1)
	}

	reporter := ui

	l := linter.NewLinter(reporter, logger)

	cli := linter.NewCLI(ui, loader, l, logger)

	err = cli.Run(*fixOpt)
	if err != nil {
		ui.DisplayError(err)
		os.Exit(1)
	}
}

func buildLogger(debug bool) *log.Logger {
	var logDevice io.Writer

	if debug {
		logDevice = os.Stderr
	} else {
		var err error

		logDevice, err = os.Open(os.DevNull)
		if err != nil {
			os.Exit(1)
		}
	}

	return log.New(logDevice, "[debug] ", 0)
}
