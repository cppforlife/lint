package check

import (
	"go/ast"
	"go/token"
	"path/filepath"
	"strings"

	goloader "code.google.com/p/go.tools/go/loader"

	"github.com/cppforlife/lint/check/fix"
)

type testPackageSuffixFinder struct{}

func NewTestPackageSuffixFinder() testPackageSuffixFinder {
	return testPackageSuffixFinder{}
}

func (c testPackageSuffixFinder) FindInAST(
	walker AstWalker,
	pkg *goloader.PackageInfo,
	file *ast.File,
	fset *token.FileSet,
) []Check {
	return []Check{NewTestPackageSuffix(pkg, file, fset)}
}

type testPackageSuffix struct {
	pkg  *goloader.PackageInfo
	file *ast.File
	fset *token.FileSet
}

// NewTestPackageSuffix constructs a check
// to make sure test files belong to a _test package
// instead of the same package that is being tested
func NewTestPackageSuffix(
	pkg *goloader.PackageInfo,
	file *ast.File,
	fset *token.FileSet,
) testPackageSuffix {
	return testPackageSuffix{
		pkg:  pkg,
		file: file,
		fset: fset,
	}
}

func (c testPackageSuffix) Check() ([]Problem, error) {
	var problems []Problem

	pkgPos := c.fset.Position(c.file.Package)
	fileName := filepath.Base(pkgPos.Filename)
	pkgName := c.pkg.Pkg.Name()

	isTestFile := strings.HasSuffix(fileName, "_test.go")
	isTestPkg := strings.HasSuffix(pkgName, "_test")

	if isTestFile && !isTestPkg {
		problems = append(problems, Problem{
			Text:     "Test file should be in a corresponding test package",
			Package:  c.pkg.Pkg,
			Position: pkgPos,
			Context: Context{
				"fileName": fileName,
			},
			Fixes: []fix.Fix{
				fix.NewPackageRename(
					fix.SimpleDiff{
						Name:    "package",
						Current: pkgName,
						Desired: pkgName + "_test",
					},
					c.file,
					c.fset,
				),
			},
		})
	}

	return problems, nil
}
