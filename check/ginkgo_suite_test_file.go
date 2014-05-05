package check

import (
	"go/ast"
	"go/token"
	"path/filepath"
	"reflect"
	"strings"

	goloader "code.google.com/p/go.tools/go/loader"

	"github.com/cppforlife/lint/check/fix"
)

type gingkoSuiteTestFileFinder struct{}

func NewGingkoSuiteTestFileFinder() gingkoSuiteTestFileFinder {
	return gingkoSuiteTestFileFinder{}
}

func (c gingkoSuiteTestFileFinder) FindInAST(
	walker AstWalker,
	pkg *goloader.PackageInfo,
	file *ast.File,
	fset *token.FileSet,
) []Check {
	return []Check{NewGingkoSuiteTestFile(pkg, file, fset)}
}

type gingkoSuiteTestFile struct {
	pkg  *goloader.PackageInfo
	file *ast.File
	fset *token.FileSet
}

// NewGingkoSuiteTestFile constructs a check
// to make sure if ginkgo test library is used in *_test.go files,
// *_suite_test.go files exists in directories with those files
func NewGingkoSuiteTestFile(
	pkg *goloader.PackageInfo,
	file *ast.File,
	fset *token.FileSet,
) gingkoSuiteTestFile {
	return gingkoSuiteTestFile{
		pkg:  pkg,
		file: file,
		fset: fset,
	}
}

func (c gingkoSuiteTestFile) Check() ([]Problem, error) {
	pkgPos := c.fset.Position(c.file.Package)
	fileName := filepath.Base(pkgPos.Filename)
	dirPath := filepath.Dir(pkgPos.Filename)

	isTestFile := strings.HasSuffix(fileName, "_test.go")
	isSuiteTestFile := strings.HasSuffix(fileName, "suite_test.go")

	if isTestFile && !isSuiteTestFile && c.importsGinkgo() {
		foundFileNames, err := c.suiteTestFileNames(dirPath)
		if err != nil {
			return []Problem{}, err
		}

		expectedFileName := filepath.Base(dirPath) + "_suite_test.go"

		return c.compare(dirPath, expectedFileName, foundFileNames), nil
	}

	return []Problem{}, nil
}

func (c gingkoSuiteTestFile) compare(dirPath string, expectedFileName string, foundFileNames []string) []Problem {
	if reflect.DeepEqual([]string{expectedFileName}, foundFileNames) {
		return []Problem{}
	}

	problem := Problem{
		Package:  c.pkg.Pkg,
		Position: c.fset.Position(c.file.Package),
	}

	switch len(foundFileNames) {
	case 0:
		problem.Text = "Missing ginkgo suite test file"
		problem.Fixes = []fix.Fix{
			fix.FileRename{
				Diff: fix.SimpleDiff{
					Name:           "suiteTestFileName",
					Desired:        expectedFileName,
					MissingCurrent: true,
				},
				DirPath: dirPath,
			},
		}

	case 1:
		problem.Text = "Ginkgo suite test file name should match directory name"
		problem.Fixes = []fix.Fix{
			fix.FileRename{
				Diff: fix.SimpleDiff{
					Name:    "suiteTestFileName",
					Current: foundFileNames[0],
					Desired: expectedFileName,
				},
				DirPath: dirPath,
			},
		}

	default: // >1
		problem.Text = "Ginkgo suite test file name should match directory name"
		problem.Diffs = []fix.Diff{
			fix.SimpleDiff{
				Name:    "suiteTestFileName",
				Current: strings.Join(foundFileNames, ", "),
				Desired: expectedFileName,
			},
		}
	}

	return []Problem{problem}
}

// importsGinkgo determines if current file imports
// any package that looks like ginkgo test library
func (c gingkoSuiteTestFile) importsGinkgo() bool {
	for _, importSpec := range c.file.Imports {
		// import path is a quoted string
		if strings.HasSuffix(importSpec.Path.Value, `/ginkgo"`) {
			return true
		}
	}
	return false
}

// suiteTestFiles returns list of possible suite test files
// found in current file's directory
// e.g. suite_test.go, pkg_suite_test.go
func (c gingkoSuiteTestFile) suiteTestFileNames(dirPath string) ([]string, error) {
	paths, err := filepath.Glob(filepath.Join(dirPath, "*suite_test.go"))
	if err != nil {
		return []string{}, err
	}

	var fileNames []string

	for _, path := range paths {
		fileNames = append(fileNames, filepath.Base(path))
	}

	return fileNames, nil
}
