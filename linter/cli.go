package linter

import (
	"fmt"
	"log"
	"runtime"

	goloader "code.google.com/p/go.tools/go/loader"

	"github.com/cppforlife/lint/check"
	"github.com/cppforlife/lint/check/fix"
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

func (c cli) Run(shouldFixProblems bool) error {
	c.setGOMAXPROCS()

	programsCh, loaderErrsCh, err := c.loader.Programs()
	if err != nil {
		return fmt.Errorf("Loading programs %#v", err)
	}

	numPrograms := 0

	problemssCh := make(chan []check.Problem)
	linterErrsCh := make(chan error)

	for program := range programsCh {
		numPrograms++

		go func(program *goloader.Program) {
			problems, err := c.linter.Run(program)
			linterErrsCh <- err
			problemssCh <- problems
		}(program)
	}

	var lastErr error

	err = c.drainLoaderErrs(loaderErrsCh)
	if err != nil {
		lastErr = err
	}

	err = c.drainLinterErrs(linterErrsCh, numPrograms)
	if err != nil {
		lastErr = err
	}

	fixes := c.drainProblems(problemssCh, numPrograms)

	if shouldFixProblems {
		err = c.applyFixes(fixes)
		if err != nil {
			lastErr = err
		}
	}

	return lastErr
}

func (c cli) setGOMAXPROCS() {
	numCPU := runtime.NumCPU()

	c.logger.Printf("Setting GOMAXPROCS=%d\n", numCPU)

	runtime.GOMAXPROCS(numCPU)
}

func (c cli) drainLoaderErrs(errsCh <-chan error) error {
	var lastErr error

	for err := range errsCh {
		if err != nil {
			lastErr = err
			c.ui.DisplayError(err)
		}
	}

	return lastErr
}

func (c cli) drainLinterErrs(errsCh chan error, numPrograms int) error {
	var lastErr error

	for i := 0; i < numPrograms; i++ {
		err := <-errsCh
		if err != nil {
			lastErr = err
			c.ui.DisplayError(err)
		}
	}

	return lastErr
}

func (c cli) drainProblems(problemssCh chan []check.Problem, numPrograms int) []fix.Fix {
	var fixes []fix.Fix

	for i := 0; i < numPrograms; i++ {
		for _, problem := range <-problemssCh {
			fixes = append(fixes, problem.Fixes...)
		}
	}

	return fixes
}

func (c cli) applyFixes(fixes []fix.Fix) error {
	var lastErr error

	for _, fix := range fixes {
		err := fix.Fix()
		if err != nil {
			lastErr = err
			c.ui.DisplayError(err)
		}
	}

	return lastErr
}
