# requiref -- partially applied testify assertions

Replaces the `assert: func(t *testing.T, got string) { ... }` column in table tests with a
typed value that carries the assertion *and* the line where the table entry declared it.

```
go test ./hack/dima-stuff/requiref/...
```

The `example` and `example2` packages fail on purpose, so you can see the output.

## Typed, and testify still drops straight in

```go
type GenericAssertionFunc[T any] func(t TestingT, actual T, msgAndArgs ...any)
```

A defined *func* type, not a struct. That one decision buys both things at once.

Typed, so the compiler catches what testify cannot:

```go
want requiref.GenericAssertionFunc[int64]
{want: requiref.Int64.Equal(0)}   // ok
{want: requiref.Equal(0)}         // NO: infers GenericAssertionFunc[int]
```

`require.Equal(t, "us-east-1a", 42)` compiles under stock testify. A `GenericAssertionFunc[string]`
column rejects it.

And still open to plain testify, because a func value has an unnamed type and is
therefore assignable to any defined type with the same shape. No adapter, no conversion:

```go
var (
	_ requiref.GenericAssertionFunc[error]  = require.NoError
	_ requiref.GenericAssertionFunc[error]  = require.Error
	_ requiref.GenericAssertionFunc[any]    = require.Empty
	_ requiref.GenericAssertionFunc[any]    = require.NotNil
	_ requiref.GenericAssertionFunc[bool]   = require.True
	_ requiref.GenericAssertionFunc[string] = require.FileExists
)
```

All of those compile -- see `example/interop_test.go`, which also runs a mixed column.
Going the other way is a one-token conversion: `require.ValueAssertionFunc(c)`.

A struct-shaped `GenericAssertionFunc` would have closed that door; that was the first version, and
this is why it is gone.

## Usage

```go
tests := []struct {
	name    string
	pvZones []string
	srcZone string
	want    requiref.GenericAssertionFunc[string]
}{
	{name: "no PVCs returns empty", want: requiref.String.Empty()},
	{name: "single-zone PV",        want: requiref.Equal("us-east-1a")},
	{name: "raw closure fits too",  want: func(t requiref.TestingT, got string, _ ...any) {
		require.Equal(t, "ap-south-1a", got)
	}},
}

for _, tc := range tests {
	t.Run(tc.name, func(t *testing.T) {
		tc.want(t, requiredZone(tc.pvZones, tc.srcZone))
	})
}
```

The call is `tc.want(t, got)` -- the same shape as testify's own
`ValueAssertionFunc` idiom, no method call.

The zero `GenericAssertionFunc` is nil. A column that may be empty needs a nil guard in the loop,
or `Nothing[T]()`.

## Output is stock testify

```
--- FAIL: TestRequiredZone/pod_with_multi-zone_PV_returns_empty (0.00s)
    zone_test.go:59:
        	Error Trace:	/.../example/zone_test.go:59
        	Error:      	Should be empty, but was us-east-1a
        	Test:       	TestRequiredZone/pod_with_multi-zone_PV_returns_empty
        	Messages:   	declared at /.../example/zone_test.go:45
```

`Error Trace` is the call site in the loop body. `Messages` is the table entry.
No frame from this package appears, and there is no custom output format --
the declaration site rides in testify's own `msgAndArgs`.

## The four things that make it work

### 1. testify elides frames from directories named assert/require/mock

`assert.CallerInfo` drops any frame whose *parent directory* is named `assert`,
`require` or `mock` (the `dir != "assert" && dir != "mock" && dir != "require"` check
in `assert/assertions.go`).

The directory must literally be `require`. `requiref` does NOT work -- verified,
two wrapper frames come back:

```
Error Trace:	/.../check/requiref/generated.go:14
            	/.../check/requiref/check.go:29
            	/.../ex2/zone_test.go:54
```

So: directory `require`, package `require`, and `requiref` is the *import alias* --
which you need anyway in files that still import testify's `require`.

### 2. The declaration site goes in msgAndArgs

`declaredAt` folds the captured location into the `msgAndArgs` testify already accepts.
No custom `TestingT`, no overridden `Errorf`, no `fmt.Print` to stdout.

That last point matters: the real `pkg/compatibility/migration_test.go` table runs
`t.Parallel()`, so writing to stdout would interleave across subtests and confuse
`go test -json`.

### 3. Every closure inlines the Helper check

Each generated closure starts with `if h, ok := t.(tHelper); ok { h.Helper() }`,
inlined rather than calling a shared helper -- `Helper` marks whichever function calls
it, so it has to run in the closure's own frame. Same reason testify inlines it.

Without it the failure is attributed to `generated.go:59` instead of the test file.
Measured both ways.

### 4. The *testing.T is bound at assert time

`tc.want(t, got)` passes `t` in. An earlier shape captured `t` when the check was built;
if that was the parent subtest and the assertion ran in a nested `t.Run`, testify's
`FailNow` hit the parent and Go printed

```
test executed panic(nil) or runtime.Goexit: subtest may have called FailNow on a parent test
```

with the inner subtest running on past the failure.

## Type inference, and why Assertions[T] exists

Go infers type parameters from arguments, never from the assignment target.
`Equal("us-east-1a")` infers `GenericAssertionFunc[string]`; `Empty()` has nothing to infer from.

`Assertions[T]` binds the subject type once. Methods on a generic type carry `T` already,
so nothing needs inferring:

```go
want: requiref.String.Empty(),       // prebuilt binder, require/assertions_types.go
want: requiref.Assertions[Zone]{}.Empty(),  // anything not prebuilt
want: requiref.Empty[string](),      // or just instantiate
want: requiref.Equal("us-east-1a"),  // arg-taking: inference works
```

### Prefer the binder even when inference would work

Inference reads the *argument's* type, not the column's. Both verified:

```go
type introw struct{ want requiref.GenericAssertionFunc[int64] }
{want: requiref.Equal(0)}        // NO: untyped const defaults to int
{want: requiref.Int64.Equal(0)}  // ok: method param is int64

type Zone string
{want: requiref.Equal("us-east-1a")}             // NO: infers GenericAssertionFunc[string]
{want: requiref.Assertions[Zone]{}.Equal("us-east-1a")} // ok
```

The binder puts the column's type on the method parameter, so untyped constants convert
as they would in any normal assignment. This matters constantly with k8s API types,
which are mostly named strings.

### Same name as testify, different binding

`require.Assertions` binds `t`; `requiref.Assertions[T]` binds `T`. Same move, one
parameter over -- testify cannot make `t` implicit in a free function, Go cannot infer a
type parameter from an assignment target.

The methods differ in kind, which is the one thing to know:

```go
require.New(t).Equal("a", got)           // asserts now
requiref.Assertions[string]{}.Equal("a") // returns a GenericAssertionFunc[string]
```

### Why not variance

Verified on go1.27.1 -- generic types are invariant in both directions, and plain func
types are too:

```
cannot use GenericAssertionFunc[any]{}    as GenericAssertionFunc[string]
cannot use GenericAssertionFunc[string]{} as GenericAssertionFunc[any]
cannot use fAny (func(any)) as func(string)
```

The direction that would help is *contra*variance, since `T` sits in a parameter
position. Generic interfaces are not an escape either -- method sets need exact
signatures. Go has variance nowhere except interface satisfaction.

## Composition

```go
want: requiref.All(
	requiref.On(func(p pod) string { return p.Name }, requiref.Equal("counter")),
	requiref.On(func(p pod) []string { return p.Containers }, requiref.Strings.Len(2)),
),
```

`On` is the getter half of a lens -- contramap. Both sides are statically typed, so
unlike an `any`-based version there is no assertion and no runtime failure mode.
It is also the only useful half: assertions only read.

`All` is trivial now that each check carries its own site internally.
`Nothing[T]()` asserts nothing.

Negation combinators are not worth building: testify ships `Not*` variants, and there
is no way to invert an arbitrary assertion's failure message into a readable one.

Caveat: `All` and `On` add stack frames. testify's `CallerInfo` reads the stack in
batches of 10 and skips frames across batch boundaries, so deep nesting can add a stray
`runtime/asm_amd64.s` line to `Error Trace`. Cosmetic only.

## loc() walks instead of counting

A fixed `runtime.Caller` skip breaks the moment anything wraps a constructor -- a facade,
a project-local helper, an alias. Measured, before the fix:

```
requiref.String.Empty()     declared at facade_test.go:15   (no extra frame)
requiref.Empty[string]()    declared at facade.go:24        (wrapper, off by one)
```

`loc()` walks the stack and returns the first frame outside any registered package.
The library registers itself in `init()`; a wrapper registers itself with
`require.RegisterWrapper()` from its own `init()`.

This also kills the bug class from `test_located_v2.go`, where `Equal`, `NoError` and
`New` each needed a different skip and two of them were wrong.

## Codegen

```
go generate ./require/     # regenerate
./verify-codegen.sh        # fail if stale
```

`internal/gen` reads the testify version this module already depends on
(`go list -m -f {{.Dir}} github.com/stretchr/testify`), so the output cannot drift.

```
generated 70 checks (42 generic, 28 concrete) and 42 Assertions[T] methods
```

Selection: exported, first parameter `TestingT`, returns `bool`, name does not end in
`f`. That is 76 functions; 6 are skipped, leaving 70.

The hole -- the parameter the actual value fills in -- follows a rule:

1. the parameter named `actual`, else
2. the parameter named `object`, else
3. parameter 0.

Right for 72 of 76. The rest is `internal/gen/config.go`, about 30 lines:

- `skip` -- `Fail`, `FailNow`, `Condition`, `Eventually`, `EventuallyWithT`, `Never`.
  None has a single actual value to partially apply.
- `holeOverride` -- `Regexp` and `NotRegexp` take `(rx, str)`, `PanicsWithError` and
  `PanicsWithValue` take `(x, f)`. The subject is the last parameter in all four.
- `bindToT` / `looseExpected` -- which sibling parameters share the subject's type.
  `expected` by default, plus `e2` for the Greater/Less family and `listB` for
  `ElementsMatch`; `EqualValues`, `InDelta` and `InEpsilon` compare deliberately across
  types, so theirs stays `any`.

Generic or concrete falls straight out of the hole's declared type. `interface{}` hole
(42) gives `X[T any](...) GenericAssertionFunc[T]` plus an `Assertions[T]` method; concrete hole (28) gives
`GenericAssertionFunc[error]`, `GenericAssertionFunc[string]`, `GenericAssertionFunc[bool]`, `GenericAssertionFunc[assert.PanicTestFunc]`,
`GenericAssertionFunc[http.HandlerFunc]`, `GenericAssertionFunc[time.Time]`.

Concrete-hole checks get no `Assertions[T]` method: `NoError()` is already fully determined,
so a bound subject would just be a `T` the method ignores.

## As a library

Separate module, test-only, one dependency. This repo already tags subdirectory modules
(the `cli/snapshot-cli/v0.0.x` convention).

```
<module root>/
  go.mod
  require/                -- MUST be this directory name; package require
    assertion_func.go loc.go combinators.go for.go for_types.go
    generated.go assertions_generated.go
  requiref/               -- optional re-export facade
  internal/gen/           -- the generator
  verify-codegen.sh
  elision_test.go
```

Import as `requiref "…/require"`.

### The facade is optional

Go has no generic function values, so `var Empty = require.Empty` does not compile:

```
cannot use generic function require.Empty without instantiation
```

Types, binders and the 28 concrete constructors are plain re-exports; the 42 generic
ones need real wrapper functions. Since an import alias costs one line per file,
the facade is not worth maintaining unless you want the name.

`RegisterWrapper` is what keeps the declaration site correct through those wrappers.

### Guard test

The clean `Error Trace` rests on an implementation detail of testify -- the dir-name
check in `assert.CallerInfo`. Nothing in testify's API promises it.

`elision_test.go` runs the example suites in a subprocess and fails if any library path
appears in a reported trace. It is verified to catch a real leak, not to pass vacuously:
adding `/example/` to its deny list makes it fail as expected.

That test plus `verify-codegen.sh` is what makes a testify upgrade safe.

## Files

- `require/check.go` -- `GenericAssertionFunc[T]`, `TestingT`, `tHelper`, `declaredAt`
- `require/loc.go` -- declaration-site lookup by stack walk, and `RegisterWrapper`
- `require/assertions.go`, `require/assertions_types.go` -- the subject binder and prebuilt ones
- `require/combinators.go` -- `All`, `On`, `Nothing`
- `require/generated.go`, `require/assertions_generated.go` -- generated
- `internal/gen/` -- the generator; `config.go` holds the hand-written exceptions
- `verify-codegen.sh`, `elision_test.go` -- drift check and elision guard
- `example/zone_test.go` -- `pkg/compatibility/migration_test.go` TestRequiredZone, rewritten
- `example/interop_test.go` -- plain testify functions in typed columns
- `example/compose_test.go` -- `All` / `On` / `Nothing`
- `example2/facade_test.go` -- the re-export facade
