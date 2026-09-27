// Package requiref re-exports the internal require package.
//
// The wrapped package must live in a directory named "require" for testify's
// CallerInfo to elide its frames. This facade exists only to offer a name that
// does not collide with testify's own require; an import alias does the same
// job for one line per file, which is why this is optional.
//
// Go has no generic function values, so the 42 generic constructors need real
// wrapper functions -- `var Empty = require.Empty` does not compile. Only the
// types, the binders and the 28 concrete constructors are plain re-exports.
package requiref

import "github.com/kr3v/requiref/require"

func init() { require.RegisterWrapper() }

// Type aliases: identical types, not new ones.
type (
	GenericAssertionFunc[T any] = require.GenericAssertionFunc[T]
	Assertions[T any]           = require.Assertions[T]
	TestingT                    = require.TestingT
)

// Plain re-exports: binders, and every concrete-subject constructor.
var (
	String = require.String
	Int64  = require.Int64
	Err    = require.Err

	NoError = require.NoError
	Error   = require.Error
	ErrorIs = require.ErrorIs
	True    = require.True
	False   = require.False
)

// Wrappers, forced by the lack of generic function values. RegisterWrapper
// above is what keeps the reported declaration site on the caller's line
// rather than on these.
func Empty[T any](msgAndArgs ...any) GenericAssertionFunc[T] { return require.Empty[T](msgAndArgs...) }

func Equal[T any](expected T, msgAndArgs ...any) GenericAssertionFunc[T] {
	return require.Equal(expected, msgAndArgs...)
}

// ...and so on for the other 40. The generator can emit them.
