package check

import (
	"go/ast"
	"go/token"
	"path/filepath"
	"strings"

	goloader "code.google.com/p/go.tools/go/loader"

	"github.com/cppforlife/lint/check/fix"
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

func (c packageDirName) Check() ([]Problem, error) {
	pkgPos := c.fset.Position(c.file.Package)
	dirName := filepath.Base(filepath.Dir(pkgPos.Filename))
	pkgName := c.pkg.Pkg.Name()

	expectedPkgName := strings.TrimSuffix(pkgName, "_test")
	isTestPkgName := strings.HasSuffix(pkgName, "_test")

	// Ignore main package since it corresponds to app-named directory
	if expectedPkgName == "main" {
		return []Problem{}, nil
	}

	var problems []Problem

	if dirName != expectedPkgName {
		problem := Problem{
			Package:  c.pkg.Pkg,
			Position: pkgPos,
			Context: Context{
				"dirName": dirName,
			},
		}

		if isTestPkgName {
			problem.Text = "Test package name should match directory name with _text suffix"
			problem.Fixes = []fix.Fix{
				fix.NewPackageRename(
					fix.SimpleDiff{
						Name:    "package",
						Current: pkgName,
						Desired: dirName + "_test",
					},
					c.file,
					c.fset,
				),
			}
		} else {
			problem.Text = "Package name should match directory name"
			problem.Fixes = []fix.Fix{
				fix.NewPackageRename(
					fix.SimpleDiff{
						Name:    "package",
						Current: pkgName,
						Desired: dirName,
					},
					c.file,
					c.fset,
				),
			}
		}

		problems = append(problems, problem)
	}

	return problems, nil
}
