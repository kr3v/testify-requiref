package requiref_test

import (
	"os/exec"
	"regexp"
	"strings"
	"testing"
)

// The clean "Error Trace" depends on an implementation detail of testify:
// assert.CallerInfo drops frames whose parent directory is named
// assert/require/mock. Nothing in testify's API promises that.
//
// This runs the example suite for real and fails if any frame from this
// library shows up in the reported trace, so a testify upgrade that changes
// CallerInfo is caught here rather than in every downstream test's output.
func TestNoLibraryFramesInErrorTrace(t *testing.T) {
	if testing.Short() {
		t.Skip("runs go test in a subprocess")
	}

	out, err := exec.Command("go", "test", "-count=1",
		"github.com/kr3v/requiref/example",
		"github.com/kr3v/requiref/example2",
	).CombinedOutput()
	if err == nil {
		t.Fatal("example suites are expected to fail on purpose; they passed")
	}

	// Collect the Error Trace blocks: the "Error Trace:" line plus its
	// continuation lines, up to the next labeled line.
	var (
		trace   []string
		inTrace bool
		label   = regexp.MustCompile(`^\s*\w[\w ]*:\s`)
	)
	for _, line := range strings.Split(string(out), "\n") {
		switch {
		case strings.Contains(line, "Error Trace:"):
			inTrace = true
			trace = append(trace, line)
		case inTrace && label.MatchString(line):
			inTrace = false
		case inTrace:
			trace = append(trace, line)
		}
	}
	if len(trace) == 0 {
		t.Fatalf("no Error Trace found in output:\n%s", out)
	}

	for _, line := range trace {
		for _, bad := range []string{"/require/", "/requiref/facade.go"} {
			if strings.Contains(line, bad) {
				t.Errorf("library frame leaked into Error Trace: %q\n\nfull trace:\n%s",
					strings.TrimSpace(line), strings.Join(trace, "\n"))
			}
		}
	}
}
