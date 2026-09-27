package example_test

import (
	"testing"

	requiref "github.com/kr3v/testify-requiref/require"
)

type pod struct {
	Name       string
	Containers []string
}

func TestCompose(t *testing.T) {
	tests := []struct {
		name string
		got  pod
		want requiref.GenericAssertionFunc[pod]
	}{
		{
			name: "lens-style field checks",
			got:  pod{Name: "counter", Containers: []string{"a"}},
			want: requiref.All(
				requiref.On(func(p pod) string { return p.Name }, requiref.Equal("counter-XXX")),
				requiref.On(func(p pod) []string { return p.Containers }, requiref.Strings.Len(2)),
			),
		},
		{
			name: "Nothing asserts nothing",
			got:  pod{},
			want: requiref.Nothing[pod](),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) { tc.want(t, tc.got) })
	}
}
