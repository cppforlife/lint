package linter

import (
	"fmt"
	"log"
	"runtime"

	goloader "code.google.com/p/go.tools/go/loader"
)

type cli struct {
	ui     UI
	loader Loader
	linter Linter
	logger *log.Logger
}

func NewCLI(ui UI, loader Loader, linter Linter, logger *log.Logger) cli {
	return cli{ui, loader, linter, logger}
}

func (c cli) Run() error {
	c.setGOMAXPROCS()

	programsCh, loaderErrsCh, err := c.loader.Programs()
	if err != nil {
		return fmt.Errorf("Loading programs %#v", err)
	}

	numPrograms := 0

	linterErrsCh := make(chan error)
	defer close(linterErrsCh)

	for program := range programsCh {
		numPrograms++
		go func(program *goloader.Program) {
			linterErrsCh <- c.linter.Run(program)
		}(program)
	}

	var firstLoaderErr, firstLinterErr error

	// Drain possible loader errs
	for err := range loaderErrsCh {
		if err != nil {
			if firstLoaderErr == nil {
				firstLoaderErr = err
			}
			c.ui.DisplayError(err)
		}
	}

	// Drain possible linter errors
	for i := 0; i < numPrograms; i++ {
		err := <-linterErrsCh
		if err != nil {
			if firstLinterErr == nil {
				firstLinterErr = err
			}
			c.ui.DisplayError(err)
		}
	}

	if firstLoaderErr != nil {
		return firstLoaderErr
	}

	return firstLinterErr
}

func (c cli) setGOMAXPROCS() {
	numCPU := runtime.NumCPU()
	c.logger.Printf("Setting GOMAXPROCS=%d\n", numCPU)
	runtime.GOMAXPROCS(numCPU)
}
