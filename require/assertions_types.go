package require

import "time"

// Prebuilt subject binders for common types, so nullary checks need no
// instantiation: String.Empty() rather than Empty[string]().
// For anything else write Assertions[T]{} inline, or bind it once in the test.
//
// Empty structs -- no state, no cost, nothing to race on.
var (
	String   = Assertions[string]{}
	Bool     = Assertions[bool]{}
	Int      = Assertions[int]{}
	Int32    = Assertions[int32]{}
	Int64    = Assertions[int64]{}
	Float64  = Assertions[float64]{}
	Bytes    = Assertions[[]byte]{}
	Duration = Assertions[time.Duration]{}
	Time     = Assertions[time.Time]{}
	Err      = Assertions[error]{} // not Error: that collides with the Error check
	Any      = Assertions[any]{}

	Strings   = Assertions[[]string]{}
	Ints      = Assertions[[]int]{}
	StringMap = Assertions[map[string]string]{}
)
