# Value Marks

----

**Note:** Marks are an optional feature that will not be needed in most
applications. If your application never uses this API then you don't need to
worry about encountering marked values, and you can ignore this document.

----

A `cty.Value` can optionally be _marked_, which causes it to carry around some
additonal metadata along with its value. Marks are just normal Go values that
are value to use as map keys, and are compared by equality.

For example, an application might use marks to track the origin of a particular
value in order to give better error messages, or to present the value in a
different way in a UI.

```go
// Use a named type for all marks, for namespacing purposes.
type fromConfigType int
val fromConfig fromConfigType

return val.Mark(fromConfig)
```

```go
if val.HasMark(fromConfig) {
    // Maybe warn the user that the value is derived from configuration?
    // Or whatever makes sense for the calling application.
}
```

When a value is marked, operation methods using that value will propagate the
marks to any result values. That makes marks "infectious" in the sense that
they propagate through operations and accumulate in the result automatically.

However, marks cannot propagate automatically thruogh _integration_ methods,
and so a calilng application is required to explicitly _unmark_ a value
before using them:

```go
val, valMarks := val.Unmark()
// ...then use integration methods with val,
// eventually producing a result that propgates
// the marks:
return result.WithMarks(valMarks)
```

## Marked Values in Sets

Sets present an interesting problem for marks because marks do not contribute
to equality of two values and thus it would be possible in principle to add
the same value to a given set twice with different marks.

To avoid the complexity of tracking superset marks on a per-element basis,
`cty` instead makes a compromise: sets can never contain marked values, and
if any marked values are passed to `cty.SetVal` then they will be automatically
unmarked and the superset of all marks applied to the resulting set as a whole.

This is lossy about exactly which elements contributed which marks, but is
conservative in the sense that any access to elements in the set will encounter
the superset marks as expected.

## Marks Under Conversion

The `cty/convert` package is aware of marks and will automatically propagate
them through conversions. That includes nested values that are marked, which
will be propagated to the corresponding nested value in the result if possible,
or will be simplified to marks on a container where an exact propagation is not
possible.

## Marks as Function Arguments

The `cty/function` package is aware of marks and will, by default,
automatically unmark all function arguments prior to calling a function and
propagate the argument marks to the result value so that most functions do
not need to worry about handling marks.

A function may opt in to handling marks itself for a particular parameter by
setting `AllowMarks: true` on the definition of that parameter. If a function
opts in, it is therefore responsible for correctly propagating any marks onto
its result.

A function's `Type` implementation will receive automatically-unmarked values
unless `AllowMarks` is set, which means that return-type checking alone will
disregard any marks unless `AllowMarks` is set. Because type checking does not
return a value, there is no way for type checking alone to communicate which
marks it encountered during its work.

If you're using marks in a use-case around obscuring sensitive values, beware
that type checking of some functions could extract information without
preserving the sensitivity mark. For example, if a string marked as sensitive
were passed as the first argument to the stdlib `JSONDecode` function then
type-checking of that function will betray the object property names inside
that value as part of the inferred result type.

## Marks Under Serialization

Marks cannot be represented in either the JSON nor the msgpack serializations
of cty values, and so the `Marshal` functions for those will return errors
if they encounter marked values.

If you need to serialize values that might contain marks, you must explicitly
unmark the whole data structure first (e.g. using `Value.UnmarkDeep`) and then
decide what to do with those marks in order to ensure that if it makes sense
to propagate them through the serialization then they will get represented
somehow.
