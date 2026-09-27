// Package require builds partially applied testify assertions that remember
// where they were declared.
//
// GenericAssertionFunc is a plain func type,
// so testify's own assertions are already GenericAssertionFunc values and the two mix freely in one column:
//
//	want: require.NoError,            // plain testify, assignable as-is
//	want: requiref.ErrorIs(ErrFoo),   // partially applied, remembers its line
//
// Failures are reported in stock testify format. The declaration site arrives
// as testify's own Messages field -- there is no custom TestingT, no overridden
// Errorf, and no custom output.
//
// The directory MUST be named "require" (or "assert"): testify's CallerInfo
// drops every stack frame whose parent directory is named assert/require/mock,
// which is what keeps this package out of the reported "Error Trace".
package require

//go:generate go run ../internal/gen .

import (
	"fmt"

	"github.com/stretchr/testify/require"
)

// GenericAssertionFunc is a testify assertion with everything but the actual value applied.
//
// It sits alongside testify's own ValueAssertionFunc, BoolAssertionFunc and ErrorAssertionFunc,
// and differs from them only in that the subject is typed:
// a GenericAssertionFunc[int64] column rejects a GenericAssertionFunc[int] at compile time.
//
// It is a defined func type rather than a struct,
// which is what keeps stock testify assignable to it:
// a func value has an unnamed type, so require.NoError is a GenericAssertionFunc[error]
// and require.Empty is a GenericAssertionFunc[any], with no adapter.
// Going the other way is a conversion: require.ValueAssertionFunc(c).
//
// The zero value is nil.
// A table column that may be empty needs either a nil guard in the loop, or Nothing[T]().
type GenericAssertionFunc[T any] func(t TestingT, actual T, msgAndArgs ...any)

// TestingT is testify's TestingT, re-exported as an alias so signatures here
// read without a second import. Identical type, not a new one.
type TestingT = require.TestingT

// tHelper mirrors testify's own unexported interface of the same name.
//
// Every generated closure starts with `if h, ok := t.(tHelper); ok { h.Helper() }`
// inlined rather than calling a shared helper: Helper marks whichever function
// calls it, so it has to run in the closure's own frame. Without it the failure
// is attributed to a line in generated.go instead of the test.
type tHelper interface{ Helper() }

// declaredAt folds the declaration site into testify's msgAndArgs, which is
// why no custom TestingT is needed. `built` is what the check was constructed
// with, `called` is what the call site passed; testify reads only one message,
// so the more specific one wins.
func declaredAt(loc string, built, called []any) []any {
	msg := message(built)
	if msg == "" {
		msg = message(called)
	}
	if msg == "" {
		return []any{"declared at " + loc}
	}
	return []any{msg + " (declared at " + loc + ")"}
}

// message mirrors testify's own messageFromMsgAndArgs.
func message(msgAndArgs []any) string {
	switch len(msgAndArgs) {
	case 0:
		return ""
	case 1:
		if s, ok := msgAndArgs[0].(string); ok {
			return s
		}
		return fmt.Sprintf("%+v", msgAndArgs[0])
	default:
		if s, ok := msgAndArgs[0].(string); ok {
			return fmt.Sprintf(s, msgAndArgs[1:]...)
		}
		return fmt.Sprintf("%+v", msgAndArgs)
	}
}
