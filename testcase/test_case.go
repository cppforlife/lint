package testcase

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type testCase struct {
	packageName string
	outputPath  string
}

func newTestCaseFromGoPath(path, goPath string) testCase {
	goSrc := filepath.Join(goPath, "src")
	return testCase{
		packageName: strings.TrimPrefix(path, goSrc),
		outputPath:  filepath.Join(path, "out.txt"),
	}
}

func (tc testCase) PackageName() string { return tc.packageName }
func (tc testCase) OutputPath() string  { return tc.outputPath }

func FindTestCases(goPath string) ([]testCase, error) {
	var testCases []testCase

	pwd, err := os.Getwd()
	if err != nil {
		return testCases, fmt.Errorf("Getwd %v", err)
	}

	walkDir := filepath.Join(pwd, "testcase")

	walkFunc := func(path string, info os.FileInfo, err error) error {
		if err == nil && info.IsDir() && path != walkDir {
			testCases = append(testCases, newTestCaseFromGoPath(path, goPath))
		}
		return err
	}

	err = filepath.Walk(walkDir, walkFunc)
	if err != nil {
		return testCases, fmt.Errorf("Walk %v", err)
	}

	return testCases, nil
}
