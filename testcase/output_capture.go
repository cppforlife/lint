package testcase

import (
	"bytes"
	"log"
	"os"
	"testing"

	"github.com/cppforlife/lint/linter"
)

func CaptureLintOutput(t *testing.T, goPath, packageName string) ([]byte, error) {
	buf := &bytes.Buffer{}

	logDevice, err := os.Open(os.DevNull)
	if err != nil {
		t.Fatalf("Open %v", err)
	}

	logger := log.New(logDevice, "[debug] ", 0)

	loader, err := linter.NewLoaderFromArgs(goPath, []string{packageName}, logger)
	if err != nil {
		t.Fatalf("NewLoaderFromArgs %v", err)
	}

	ui := linter.NewPlainUI(buf, logger)

	l := linter.NewLinter(ui, logger)

	cli := linter.NewCLI(ui, loader, l, logger)

	err = cli.Run(false)
	if err != nil {
		if _, ok := err.(linter.FoundProblemsError); !ok {
			t.Fatalf("Run %v", err)
		} else {
			err = nil
		}
	}

	return buf.Bytes(), err
}
