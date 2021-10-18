// ------------------------------------------------------ {COPYRIGHT-TOP} ---
// IBM Confidential
// Automated Tests
// Copyright IBM Corp. 2021
// ------------------------------------------------------ {COPYRIGHT-END} ---
package testing

import (
	"fmt"
	"strings"

	gomegatypes "github.com/onsi/gomega/types"
	"github.com/r3labs/diff/v2"
)

// Resemble is a Gomega matcher that lists the structural differences between 2 Go structs
func Resemble(expected interface{}) gomegatypes.GomegaMatcher {
	return &differ{
		expected: expected,
	}
}

type differ struct {
	expected interface{}
	changes  diff.Changelog
}

// Match checks if the actual resource matches the expected resource
func (d *differ) Match(actual interface{}) (success bool, err error) {
	d.changes, err = diff.Diff(d.expected, actual)
	return err == nil && len(d.changes) == 0, err
}

// FailureMessage formats the message in the result of a match failure
func (d *differ) FailureMessage(actual interface{}) (message string) {
	var result string = "Actual value differs from expected as follows:\n"
	for i, c := range d.changes {
		result += fmt.Sprintf("%d. %s:\n   Expected: %v\n   Actual: %v\n", i+1, strings.Join(c.Path, "."), c.From, c.To)
	}
	return result
}

// NegatedFailureMessage returns the inverse of the failure message
func (d *differ) NegatedFailureMessage(actual interface{}) (message string) {
	return fmt.Sprintf("Expected the following not to occur:\n%s", d.FailureMessage(actual))
}
