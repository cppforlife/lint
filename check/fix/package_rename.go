package fix

import (
	"go/ast"
	"go/token"
)

func NewPackageRename(diff Diff, file *ast.File, fset *token.FileSet) fileRewrite {
	fx := func() error {
		file.Name.Name = diff.DesiredStr()
		return nil
	}

	return fileRewrite{
		Diff: diff,
		Func: fx,
		File: file,
		Fset: fset,
	}
}
