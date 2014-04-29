package testcase

import (
	"fmt"
	"strings"
)

type outputComparison struct {
	actual   string
	expected string
}

func NewOutputComparison(
	actualOutput, expectedOutput []byte,
	goPath string,
) outputComparison {
	return outputComparison{
		actual:   normalizeOutput(actualOutput, goPath),
		expected: normalizeOutput(expectedOutput, goPath),
	}
}

func (c outputComparison) Actual() string   { return c.actual }
func (c outputComparison) Expected() string { return c.expected }
func (c outputComparison) Match() bool      { return c.actual == c.expected }

func (c outputComparison) DiffLines() string {
	actualLines := strings.Split(c.actual, "\n")
	expectedLines := strings.Split(c.expected, "\n")

	for i, line := range expectedLines {
		if len(actualLines) > i {
			if actualLines[i] != line {
				return fmt.Sprintf(
					"Difference on line %d:\n%s\n%s\n",
					i, string(actualLines[i]), string(expectedLines[i]))
			}
		} else {
			return fmt.Sprintf("Actual ends early")
		}
	}

	return fmt.Sprintf("Expected ends early")
}

func normalizeOutput(output []byte, goPath string) string {
	out := string(output)
	out = strings.Replace(out, goPath, "$GOPATH", -1)
	out = strings.Replace(out, "\t", "  ", -1)
	return strings.TrimSpace(out)
}
