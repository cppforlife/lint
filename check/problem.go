package check

import (
	"go/token"

	gotypes "code.google.com/p/go.tools/go/types"

	"github.com/cppforlife/lint/check/fix"
)

type Problem struct {
	Text string

	Package  *gotypes.Package
	Position token.Position

	Context Context

	Diffs []fix.Diff
	Fixes []fix.Fix
}

type Context map[string]string
