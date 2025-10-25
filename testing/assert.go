// Package testing provides some useful testing mechanisms.
package testing

import "testing"

var enableOkMessages = true

func Assert(condition bool, errorMessage, successMessage string, test *testing.T) {
	if !condition {
		test.Errorf("Assertion failure: %v", errorMessage)
	} else if enableOkMessages {
		test.Logf("Assertion ok: %v", successMessage)
	}
}

func EnableOkMessagesSet(on bool) {
	enableOkMessages = on
}
