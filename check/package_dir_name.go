package check

import (
	"go/ast"
	"go/token"
	"path/filepath"
	"strings"

	goloader "code.google.com/p/go.tools/go/loader"
)

type packageDirNameFinder struct{}

func NewPackageDirNameFinder() packageDirNameFinder {
	return packageDirNameFinder{}
}

func (c packageDirNameFinder) FindInAST(
	walker AstWalker,
	pkg *goloader.PackageInfo,
	file *ast.File,
	fset *token.FileSet,
) []Check {
	return []Check{NewPackageDirName(pkg, file, fset)}
}

type packageDirName struct {
	pkg  *goloader.PackageInfo
	file *ast.File
	fset *token.FileSet
}

// NewPackageDirName constructs a check
// to make sure directory name matches package name
func NewPackageDirName(
	pkg *goloader.PackageInfo,
	file *ast.File,
	fset *token.FileSet,
) packageDirName {
	return packageDirName{
		pkg:  pkg,
		file: file,
		fset: fset,
	}
}

func (c packageDirName) Check() []Problem {
	var problems []Problem

	pkgPos := c.fset.Position(c.file.Package)
	dirName := filepath.Base(filepath.Dir(pkgPos.Filename))
	pkgName := c.pkg.Pkg.Name()

	expectedPkgName := strings.TrimSuffix(pkgName, "_test")
	isTestPkgName := strings.HasSuffix(pkgName, "_test")

	// Ignore main package since it corresponds to app-named directory
	if expectedPkgName == "main" {
		return problems
	}

	if dirName != expectedPkgName {
		problem := Problem{
			Package:  c.pkg.Pkg,
			Position: pkgPos,
			Context: ProblemContext{
				"dirName": dirName,
			},
		}

		if isTestPkgName {
			problem.Text = "Test package name should match directory name with _text suffix"
			problem.Diff = []ProblemDiff{
				{Name: "package", Have: pkgName, Want: dirName + "_test"},
			}
		} else {
			problem.Text = "Package name should match directory name"
			problem.Diff = []ProblemDiff{
				{Name: "package", Have: pkgName, Want: dirName},
			}
		}

		problems = append(problems, problem)
	}

	return problems
}
