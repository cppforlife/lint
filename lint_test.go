package main_test

import (
	"fmt"
	"io/ioutil"
	"os"
	"reflect"
	"testing"

	"github.com/cppforlife/lint/testcase"
)

func TestAll(t *testing.T) {
	goPath := os.Getenv("GOPATH")
	if goPath == "" {
		t.Fatalf("GOPATH must be specified")
	}

	testCases, err := testcase.FindTestCases(goPath)
	if err != nil {
		t.Fatalf("FindTestCases %v", err)
	}

	var actualTestCaseNames, expectedTestCaseNames []string

	for _, tc := range testCases {
		actualTestCaseNames = append(actualTestCaseNames, tc.Name())

		actualOutput, err := testcase.CaptureLintOutput(t, goPath, tc.PackageName())
		if err != nil {
			t.Fatalf("CaptureLintOutput %v", err)
		}

		expectedOutput, err := ioutil.ReadFile(tc.OutputPath())
		if err != nil {
			t.Fatalf("ReadFile %v", err)
		}

		test := testcase.NewOutputComparison(actualOutput, expectedOutput, goPath)
		if !test.Match() {
			desc := fmt.Sprintf("%s did not match expected output.", tc.PackageName())
			details := fmt.Sprintf("Actual:\n_%s_\n\nExpected:\n_%s_\n", test.Actual(), test.Expected())
			t.Errorf("%s\n%s\n%s\n", desc, test.DiffLines(), details)
		}
	}

	expectedTestCaseNames = []string{
		"errorassignment",

		"ginkgosuitetestfile/invalid",
		"ginkgosuitetestfile/missing",
		"ginkgosuitetestfile/valid",

		"packagedirname/main",
		"packagedirname/other",

		"testpackagesuffix",
	}

	// Make sure all expected test cases are exercised
	if !reflect.DeepEqual(actualTestCaseNames, expectedTestCaseNames) {
		t.Fatalf("DeepEqual %v != %v", actualTestCaseNames, expectedTestCaseNames)
	}
}
