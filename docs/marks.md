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
var fromConfig fromConfigType

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

## Provenance Marks

Any comparable value can be used as a mark, but there is a subcategory of marks
called "provenance marks", which are those that are of struct types with
[`ctymarks.ProvenanceMark`](https://pkg.go.dev/github.com/zclconf/go-cty/cty/ctymarks#ProvenanceMark)
embedded within them.

Plain (i.e. non-provenance) marks are propagated through operations on a
"best-effort" basis when unknown values are present, meaning that any result
that is unknown may omit some marks that would have appeared if the input values
had already been known. In other words, you should generally assume that the
marks for an unknown value are themselves un-finalized, and that new marks
might appear if the same operands were re-evaluated with more of its operands
being known.

Provenance marks instead opt in to more conservative behavior where callers can
expect them to be propagated more conservatively: if it's unclear whether a
known result from the same operation would have a provenance mark then the
unknown value should always have that mark, and then the mark might _disappear_
once the operands are more known.

The general guideline for when to use each subcategory of marks is:

- Use plain (i.e. non-provenance) marks for situations where the presence of the
  mark causes your application to produce some sort of dynamic runtime error,
  and where it's better for that error to be withheld if there's some chance
  that the final known value would not have that mark, or if the mark is used
  purely to change the cosmetic presentation of a value in your application's
  UI and where that transformation is not important for unknown values.

    For example, OpenTofu uses a plain mark to represent that a value is
    "sensitive", which is a purely-UI-driven concern where sensitive-marked
    values are not rendered in cleartext in the UI. OpenTofu's design considers
    it okay for unknown values to appear in the UI even if they would be treated
    as sensitive in a later phase, because the rendering of an unknown value
    in its UI does not disclose any significant sensitive information.

- Use provenance marks for situations where the presence of the mark causes
  some important non-error behavior that must be handled correctly in order for
  later evaluation to produce a correct result.

    For example, if you're using marks to perform dynamic analysis of
    dependencies between objects by marking each object then testing for
    those marks on other objects then that's likely a good use-case for
    provenance marks because failing to propagate such a mark when unknown
    values are present would cause the resulting dependency graph to be
    incomplete, and thus potentially cause operations to be performed in the
    wrong order once known values are introduced in a later phase.

The unfortunate distinction between plain mark and provenance marks is a
pragmatic concession to backward-compatiblity. Although marks were originally
intended to be used for "provenance-related" use-cases, in practice their
earliest uses were purely for cosmetic tracking of sensitive values for UI
purposes (where unknown values don't tend to reveal any information anyway)
or for raising errors that ought to be deferred to when values were more known,
and so the initial implementation had some gaps that calling applications came
to rely on.

The concept of provenance marks is a compromise to allow retrofitting the more
conservative tracking behavior without breaking existing callers whose existing
marks are now considered to be "plain" by default. The most notable differences
in handling for provenance marks are:

- When calling `Value.Index` or `Value.GetAttrByValue` with an unknown key or
  attribute name, thereby making it unclear which one of the collection/object's
  nested values would be selected, plain marks on the nested elements are
  ignored, while provenance marks are preserved on the unknown result.
- When transferring marks between two values of compatible types using
  `Value.UnmarkDeepWithPaths` and `Value.MarkWithPaths`, if the new value has
  unknown value placeholders for collection or structural types and there are
  marks referring to nested paths under those unknown values then only
  provenance marks are applied as "nested marks" on the resulting unknown
  values.
- Any application which uses provenance marks, or which explicitly constructs
  unknown values with nested marks using `UnknownValWithNestedMarks`, must be
  prepared for functions like `Walk`, `Transform`, and `Value.WrangleMarksDeep`
  to produce paths that have an "unknown tail", which is a placeholder for
  all possible paths underneath an unknown value. A value at any path with
  an unknown tail is always unknown but has marks that represent which marks
  could potentially appear at any point inside the nested data structure.
- Any application which uses provenance marks, or which explicitly constructs
  unknown values with nested marks using `UnknownValWithNestedMarks`, must be
  prepared to encounter unknown values of unknown type that are not trivially
  equal (using `==`) to `cty.DynamicVal`, because they may be carrying nested
  marks.

The above are just some examples. The general rule is that for any operation
whose operands are unknown values there must be no known values that could
validly substitute for those unknown values which would cause additional
provenance marks to appear on the result. If there are any such situations
today then that is a bug and calling applications should expect it to be fixed
in a future patch release, rather than depending on the current incorrect
behavior. This rule is not necessarily upheld for "plain" marks, although for
new features added in future it's likely that both subcategories of marks will
be treated more equally.

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

## Relationship to "Refinements"

The idea of annotating a value with additional information has some overlap
with the concept of [Refinements](refinements.md). However, the two have
different purposes and so different design details and tradeoffs.

For more details, see
[the corresponding section in the Refinements documentation](refinements.md#relationship-to-marks).
