package check

import (
	"go/token"

	gotypes "code.google.com/p/go.tools/go/types"
)

type ProblemContext map[string]string

type ProblemDiff struct {
	Name string
	Have string
	Want string
}

type Problem struct {
	Text string

	Package  *gotypes.Package
	Position token.Position

	Context ProblemContext

	Diff []ProblemDiff
}
