package example_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	requiref "github.com/kr3v/requiref/require"
)

// mimics compatibility.RequiredZone -- deliberately wrong for two cases.
func requiredZone(pvZones []string, srcZone string) string {
	if len(pvZones) == 1 {
		return pvZones[0]
	}
	return srcZone // BUG: should be "" for multi-zone
}

func TestRequiredZone(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		pvZones []string
		srcZone string

		// Typed: a GenericAssertionFunc[int] would not compile here.
		want requiref.GenericAssertionFunc[string]
	}{
		{
			name:    "pod with no PVCs returns empty",
			srcZone: "us-east-1a",
			want:    requiref.String.Empty(),
		},
		{
			name:    "pod with single-zone PV returns that zone",
			pvZones: []string{"us-east-1a"},
			srcZone: "us-east-1a",
			want:    requiref.Equal("us-east-1a"),
		},
		{
			name:    "pod with multi-zone PV returns empty",
			pvZones: []string{"us-east-1a", "us-east-1b"},
			srcZone: "us-east-1a",
			want:    requiref.String.Empty(),
		},
		{
			name:    "a raw closure fits the same column",
			srcZone: "ap-south-1a",
			want: func(t requiref.TestingT, got string, _ ...any) {
				require.Equal(t, "ap-south-1a", got)
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			tc.want(t, requiredZone(tc.pvZones, tc.srcZone))
		})
	}
}
