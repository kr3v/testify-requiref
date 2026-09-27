package require

// Assertions binds the subject type once, so nullary checks need no explicit
// instantiation: Assertions[string]{}.Empty() rather than Empty[string]().
//
// It is the same move testify makes with its own Assertions, one parameter over.
// testify cannot make `t` implicit in a free function, so require.New(t) parks `t`
// on a receiver. Go cannot infer a type parameter from an assignment target,
// so this parks `T` on a receiver.
//
// The methods differ in kind, though, and that is worth knowing:
// testify's Assertions methods *run* an assertion, while these *build* one.
//
//	require.New(t).Equal("a", got)          // asserts now
//	requiref.Assertions[string]{}.Equal("a") // returns a GenericAssertionFunc[string]
//
// Binding T also makes untyped constants convert the way they would in any normal
// assignment: Assertions[int64]{}.Equal(0) compiles, while Equal(0) infers
// GenericAssertionFunc[int] from the literal and will not fit an int64 column.
//
// The methods are generated into assertions_generated.go.
type Assertions[T any] struct{}
