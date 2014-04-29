package linter

import (
	"fmt"
	"go/ast"
	"log"

	goloader "code.google.com/p/go.tools/go/loader"

	"github.com/cppforlife/lint/check"
)

type Linter interface {
	Run(program *goloader.Program) error
}

type FoundProblemsError struct {
	count int
}

func (e FoundProblemsError) Error() string {
	errorText := "error"
	if e.count == 0 || e.count > 1 {
		errorText += "s"
	}
	return fmt.Sprintf("%d %s found", e.count, errorText)
}

func (e FoundProblemsError) IsPresentable() bool { return false }

type linter struct {
	reporter Reporter
	logger   *log.Logger
}

func NewLinter(reporter Reporter, logger *log.Logger) linter {
	return linter{reporter, logger}
}

func (l linter) Run(program *goloader.Program) error {
	var checks []check.Check
	var problems []check.Problem

	finders := []check.Finder{
		check.NewErrorAssignmentsFinder(),
		check.NewTestPackageSuffixFinder(),
		check.NewPackageDirNameFinder(),
	}

	numPkgs, numFiles := 0, 0

	for _, pkg := range program.InitialPackages() {
		numPkgs++
		l.reporter.ReportPackage(pkg.Pkg)

		for _, file := range pkg.Files {
			numFiles++
			l.reporter.ReportFile(pkg.Pkg, file)

			astWalker := func(e check.AstNodeEvaler) { ast.Inspect(file, e) }

			for _, finder := range finders {
				checks = append(checks, finder.FindInAST(astWalker, pkg, file, program.Fset)...)
			}
		}
	}

	for _, check := range checks {
		problems = append(problems, check.Check()...)
	}

	for _, problem := range problems {
		l.reporter.ReportProblem(problem)
	}

	if len(problems) > 0 {
		return FoundProblemsError{count: len(problems)}
	}

	return nil
}
