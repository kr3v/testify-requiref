package example_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"

	requiref "github.com/kr3v/testify-requiref/require"
)

var errNotFound = errors.New("not found")

func lookup(key string) error {
	if key == "" {
		return errNotFound
	}
	return nil
}

// GenericAssertionFunc[T] is a defined func type, so testify's own assertions are already
// Check values. Nothing here is adapted, wrapped or converted.
func TestInterop(t *testing.T) {
	var (
		_ requiref.GenericAssertionFunc[error]  = require.NoError
		_ requiref.GenericAssertionFunc[error]  = require.Error
		_ requiref.GenericAssertionFunc[any]    = require.Empty
		_ requiref.GenericAssertionFunc[any]    = require.NotNil
		_ requiref.GenericAssertionFunc[bool]   = require.True
		_ requiref.GenericAssertionFunc[string] = require.FileExists
	)

	tests := []struct {
		name string
		key  string
		want requiref.GenericAssertionFunc[error]
	}{
		{name: "plain testify", key: "k", want: require.NoError},
		{name: "partially applied", key: "", want: requiref.ErrorIs(errNotFound)},
		{name: "deliberately failing", key: "k", want: requiref.ErrorIs(errNotFound)},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) { tc.want(t, lookup(tc.key)) })
	}
}
