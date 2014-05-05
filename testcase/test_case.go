package testcase

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type testCase struct {
	name        string
	packageName string
	outputPath  string
}

func newTestCaseFromGoPath(path, walkDir, goPath string) testCase {
	goSrc := filepath.Join(goPath, "src")
	return testCase{
		name:        withoutSep(strings.TrimPrefix(path, walkDir)),
		packageName: withoutSep(strings.TrimPrefix(path, goSrc)),
		outputPath:  filepath.Join(path, "out.txt"),
	}
}

func (tc testCase) Name() string        { return tc.name }
func (tc testCase) PackageName() string { return tc.packageName }
func (tc testCase) OutputPath() string  { return tc.outputPath }

func withoutSep(path string) string {
	return strings.TrimPrefix(path, string(filepath.Separator))
}

func FindTestCases(goPath string) ([]testCase, error) {
	var testCases []testCase

	pwd, err := os.Getwd()
	if err != nil {
		return testCases, fmt.Errorf("Getwd %v", err)
	}

	walkDir := filepath.Join(pwd, "testcase")

	walkFunc := func(path string, info os.FileInfo, err error) error {
		if err == nil && info.IsDir() && path != walkDir {
			files, err := filepath.Glob(filepath.Join(path, "*.go"))
			if err != nil {
				return err
			}

			if len(files) > 0 {
				testCases = append(testCases, newTestCaseFromGoPath(path, walkDir, goPath))
			}
		}
		return err
	}

	err = filepath.Walk(walkDir, walkFunc)
	if err != nil {
		return testCases, fmt.Errorf("Walk %v", err)
	}

	return testCases, nil
}
