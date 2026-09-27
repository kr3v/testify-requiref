package example2_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/kr3v/testify-requiref/requiref"
)

func TestFacade(t *testing.T) {
	tests := []struct {
		name string
		got  string
		want requiref.GenericAssertionFunc[string]
	}{
		{name: "binder re-export", got: "x", want: requiref.String.Empty()},
		{name: "generic wrapper", got: "x", want: requiref.Empty[string]()},
		{name: "generic wrapper with arg", got: "x", want: requiref.Equal("y")},
		{name: "plain testify in a typed column", got: "x", want: require.FileExists},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) { tc.want(t, tc.got) })
	}
}
