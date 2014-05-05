package check

import (
	"go/ast"
	"go/token"
	"path/filepath"
	"reflect"
	"strings"

	goloader "code.google.com/p/go.tools/go/loader"
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

func (c gingkoSuiteTestFile) Check() []Problem {
	pkgPos := c.fset.Position(c.file.Package)
	fileName := filepath.Base(pkgPos.Filename)
	dirName := filepath.Base(filepath.Dir(pkgPos.Filename))

	isTestFile := strings.HasSuffix(fileName, "_test.go")
	isSuiteTestFile := strings.HasSuffix(fileName, "suite_test.go")

	// Ignore non-test files and suite test files
	if !isTestFile || isSuiteTestFile {
		return []Problem{}
	}

	// Ignore test files that do not use ginkgo
	if !c.importsGinkgo() {
		return []Problem{}
	}

	foundFileNames := c.suiteTestFileNames()
	expectedFileName := dirName + "_suite_test.go"

	if reflect.DeepEqual(foundFileNames, []string{expectedFileName}) {
		return []Problem{}
	}

	var problems []Problem

	switch len(foundFileNames) {
	case 0:
		problems = append(problems, Problem{
			Text:     "Missing ginkgo suite test file",
			Package:  c.pkg.Pkg,
			Position: pkgPos,
			Diff: []ProblemDiff{
				{Name: "suiteTestFileName", MissingHave: true, Want: expectedFileName},
			},
		})
	case 1:
		problems = append(problems, Problem{
			Text:     "Ginkgo suite test file name should match directory name",
			Package:  c.pkg.Pkg,
			Position: pkgPos,
			Diff: []ProblemDiff{
				{Name: "suiteTestFileName", Have: foundFileNames[0], Want: expectedFileName},
			},
		})
	default: // >1
		problems = append(problems, Problem{
			Text:     "Ginkgo suite test file name should match directory name",
			Package:  c.pkg.Pkg,
			Position: pkgPos,
			Diff: []ProblemDiff{{
				Name: "suiteTestFileName",
				Have: strings.Join(foundFileNames, ", "),
				Want: expectedFileName,
			}},
		})
	}

	return problems
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
func (c gingkoSuiteTestFile) suiteTestFileNames() []string {
	pkgPos := c.fset.Position(c.file.Package)
	dirPath := filepath.Dir(pkgPos.Filename)

	paths, err := filepath.Glob(filepath.Join(dirPath, "*suite_test.go"))
	if err != nil {
		return []string{}
	}

	var fileNames []string

	for _, path := range paths {
		fileNames = append(fileNames, filepath.Base(path))
	}

	return fileNames
}
