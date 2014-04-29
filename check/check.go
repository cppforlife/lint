package check

import (
	"go/ast"
	"go/token"

	goloader "code.google.com/p/go.tools/go/loader"
)

type AstNodeEvaler func(ast.Node) bool
type AstWalker func(AstNodeEvaler)

type Finder interface {
	FindInAST(AstWalker, *goloader.PackageInfo, *ast.File, *token.FileSet) []Check
}

type Check interface {
	Check() []Problem
}
