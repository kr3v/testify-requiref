package main

// Everything the hole rule cannot work out on its own.
// Measured against testify v1.11.1: 76 candidates, the rule is right for 72.

// skip: no single "actual" value to partially apply.
var skip = map[string]bool{
	"Fail":            true, // unconditional failure, not an assertion
	"FailNow":         true,
	"Condition":       true, // takes a Comparison func, nothing to feed it
	"Eventually":      true, // takes a condition func + timing
	"EventuallyWithT": true,
	"Never":           true,
}

// holeOverride: the parameter the actual value fills in, when the rule picks wrong.
// -1 means the last parameter.
var holeOverride = map[string]int{
	"Regexp":          -1, // (rx, str) -- the subject is str
	"NotRegexp":       -1,
	"PanicsWithError": -1, // (errString, f) -- the subject is f
	"PanicsWithValue": -1, // (expected, f)
}

// looseExpected: generic checks whose `expected` must stay `any` rather than
// being bound to T, because comparing across types is the whole point.
var looseExpected = map[string]bool{
	"EqualValues":      true,
	"NotEqualValues":   true,
	"InDelta":          true,
	"InDeltaSlice":     true,
	"InDeltaMapValues": true,
	"InEpsilon":        true,
	"InEpsilonSlice":   true,
}

// bindToT: extra parameters that are the same type as the subject, and so
// should be typed T rather than left as any. The hole rule cannot infer this;
// it is a semantic property of each assertion.
var bindToT = map[string][]string{
	"Greater":          {"e2"},
	"GreaterOrEqual":   {"e2"},
	"Less":             {"e2"},
	"LessOrEqual":      {"e2"},
	"ElementsMatch":    {"listB"},
	"NotElementsMatch": {"listB"},
}
