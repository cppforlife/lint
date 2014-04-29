package main_test

import (
	"fmt"
	"io/ioutil"
	"os"
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

	for _, tc := range testCases {
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
}
