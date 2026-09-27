package require

import (
	"fmt"
	"runtime"
	"strings"
	"sync"
)

// Declaration sites are found by walking the stack past every package that is
// part of this library, rather than by a fixed runtime.Caller skip count.
//
// A fixed skip is wrong the moment anything wraps a constructor -- a re-export
// facade, a project-local helper, a convenience alias. Those are all one extra
// frame, and the reported line silently becomes the wrapper's.
var (
	selfMu   sync.RWMutex
	selfPkgs = map[string]bool{}
)

func init() { registerPkg(1) }

// RegisterWrapper marks the calling package as a pass-through layer, so
// declaration sites are attributed to its callers instead. Call it from an
// init() in any package that wraps or re-exports these constructors.
func RegisterWrapper() { registerPkg(2) }

func registerPkg(skip int) {
	// skip counts frames above registerPkg itself.
	pc, _, _, ok := runtime.Caller(skip)
	if !ok {
		return
	}
	f := runtime.FuncForPC(pc)
	if f == nil {
		return
	}
	selfMu.Lock()
	selfPkgs[pkgOf(f.Name())] = true
	selfMu.Unlock()
}

// pkgOf turns "github.com/a/b/require.Equal[go.shape.string]" into
// "github.com/a/b/require". Also handles methods and closures.
func pkgOf(fn string) string {
	slash := strings.LastIndexByte(fn, '/')
	dot := strings.IndexByte(fn[slash+1:], '.')
	if dot < 0 {
		return fn
	}
	return fn[:slash+1+dot]
}

func loc() string {
	const maxDepth = 32
	pcs := make([]uintptr, maxDepth)
	n := runtime.Callers(2, pcs)
	frames := runtime.CallersFrames(pcs[:n])
	selfMu.RLock()
	defer selfMu.RUnlock()
	for {
		f, more := frames.Next()
		if !selfPkgs[pkgOf(f.Function)] {
			return fmt.Sprintf("%s:%d", f.File, f.Line)
		}
		if !more {
			return "?"
		}
	}
}
