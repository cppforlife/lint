package linter

import (
	"bufio"
	"fmt"
	"go/ast"
	"go/token"
	"io"
	"log"
	"path/filepath"
	"sync"

	gotypes "code.google.com/p/go.tools/go/types"

	"github.com/cppforlife/lint/check"
)

type UI interface {
	DisplayError(error)
}

type PresentableError interface {
	IsPresentable() bool
}

type ErrorWithUnderlyingErrors interface {
	UnderlyingErrs() []error
}

type plainUIMsg int

// Different types of messages presented
const (
	plainUIPackage plainUIMsg = iota
	plainUIFile
	plainUIProblem
	plainUIError
)

type plainUI struct {
	writer    *bufio.Writer
	printLock sync.Mutex

	lastPackage  *gotypes.Package
	lastPosition token.Position

	lastMsg plainUIMsg

	logger *log.Logger
}

func NewPlainUI(writer io.Writer, logger *log.Logger) *plainUI {
	return &plainUI{
		writer: bufio.NewWriter(writer),
		logger: logger,
	}
}

func (ui *plainUI) ReportPackage(pkg *gotypes.Package) {
	ui.printLock.Lock()
	defer ui.printLock.Unlock()

	ui.writeLnAfterLastMsg(plainUIPackage)

	ui.write("Looking at package \"%s\"\n", pkg.Path())
	defer ui.flush()
}

func (ui plainUI) ReportFile(pkg *gotypes.Package, file *ast.File) {}

func (ui *plainUI) ReportProblem(problem check.Problem) {
	ui.printLock.Lock()
	defer ui.printLock.Unlock()

	lastMsg := ui.writeLnAfterLastMsg(plainUIProblem)

	if problem.Package == nil {
		panic(fmt.Sprintf("Missing package for problem: %#v", problem))
	}

	if !ui.lastPosition.IsValid() || ui.lastPosition.Filename != problem.Position.Filename {
		if lastMsg == plainUIProblem {
			ui.write("\n")
		}
		ui.write("-- %s\n", problem.Position.Filename)
	}

	ui.write(
		"%s:%d:%d %s\n",
		filepath.Base(problem.Position.Filename),
		problem.Position.Line,
		problem.Position.Column,
		problem.Text,
	)

	for name, value := range problem.Context {
		ui.write("\t%s = %s\n", name, value)
	}

	for _, diff := range problem.Diff {
		have := diff.Have
		if diff.MissingHave {
			have = "(missing)"
		}
		ui.write("\t%s : %s -> %s\n", diff.Name, have, diff.Want)
	}

	ui.lastPackage = problem.Package
	ui.lastPosition = problem.Position

	defer ui.flush()
}

func (ui *plainUI) DisplayError(err error) {
	ui.printLock.Lock()
	defer ui.printLock.Unlock()

	// Some errors are not worth showing
	if presentableErr, ok := err.(PresentableError); ok {
		if !presentableErr.IsPresentable() {
			return
		}
	}

	ui.writeLnAfterLastMsg(plainUIError)

	ui.write("[error] %s\n", err.Error())

	// If error was caused by additional errors include those here
	if errWithUnderlyingErrs, ok := err.(ErrorWithUnderlyingErrors); ok {
		for _, underlyingErr := range errWithUnderlyingErrs.UnderlyingErrs() {
			ui.write("        - %s\n", underlyingErr.Error())
		}
	}

	defer ui.flush()
}

func (ui *plainUI) writeLnAfterLastMsg(currentMsg plainUIMsg) plainUIMsg {
	lm := ui.lastMsg
	if lm != currentMsg {
		ui.write("\n")
	}

	ui.lastMsg = currentMsg

	return lm
}

func (ui plainUI) write(format string, args ...interface{}) {
	_, err := fmt.Fprintf(ui.writer, format, args...)
	if err != nil {
		ui.logger.Printf("Failed to print UI: %#v", err)
	}
}

func (ui plainUI) flush() {
	err := ui.writer.Flush()
	if err != nil {
		ui.logger.Printf("Failed to flush UI: %#v", err)
	}
}
