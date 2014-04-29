package linter

import (
	"go/ast"

	gotypes "code.google.com/p/go.tools/go/types"

	"github.com/cppforlife/lint/check"
)

type Reporter interface {
	ReportPackage(*gotypes.Package)
	ReportFile(*gotypes.Package, *ast.File)
	ReportProblem(check.Problem)
}
