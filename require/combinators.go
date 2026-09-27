package require

// All runs every check against the same value. Each check carries its own
// declaration site, so a failure still points at the line that declared it.
// Plain testify functions mix in: All[any](require.NotNil, Len[any](2)).
//
// require semantics still apply: the first failure aborts the subtest.
func All[T any](cs ...GenericAssertionFunc[T]) GenericAssertionFunc[T] {
	return func(t TestingT, actual T, called ...any) {
		if h, ok := t.(tHelper); ok {
			h.Helper()
		}
		for _, c := range cs {
			if c == nil {
				continue
			}
			c(t, actual, called...)
		}
	}
}

// On focuses a check on a part of a larger value -- the getter half of a lens.
// Both sides are statically typed; there is no assertion and no failure mode.
//
//	On(func(p *corev1.Pod) string { return p.Name }, Equal("counter"))
func On[P, C any](get func(P) C, c GenericAssertionFunc[C]) GenericAssertionFunc[P] {
	return func(t TestingT, parent P, called ...any) {
		if h, ok := t.(tHelper); ok {
			h.Helper()
		}
		if c == nil {
			return
		}
		c(t, get(parent), called...)
	}
}

// Nothing asserts nothing, for table entries that do not care when the loop
// calls the check unconditionally.
func Nothing[T any]() GenericAssertionFunc[T] {
	return func(TestingT, T, ...any) {}
}
