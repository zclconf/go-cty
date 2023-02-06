# Value Refinements

_Refinements_ are dynamic annotations associated with unknown values that
each shrink the range of possible values futher than can be represented by
type constraint alone.

When an unknown value is refined, it allows certain operations against that
unknown value to produce a known result, and allows some operations to fail
earlier than they would with a fully-unknown value by detecting that a valid
result is impossible using just the refinement information.

Refinements always _shrink_ the range of an unknown value, and never grow it.
That makes it valid for some operations to ignore refinements and just treat
an unknown value as representing any possible value of its type constraint,
which is important to avoid burdening all downstream callers of `cty` from
handling all refinements and from immediately adding support for new kinds of
refinement if this model gets extended in future releases.

However, note that `Value.RawEquals` _does_ take into account refinements, so
any tests that assert against the exact final value of an operation may need
to be updated after adopting a new version of `cty` which makes increased use
of refinements. `Value.RawEquals` is not intended as part of the _user model_
of `cty` and so this should not negatively impact the end-user-visible behavior
of an application using `cty`, although of course they might benefit from
more specific results from operations that can now take refinements into
account.

## How to refine a value

You can derive a more refined value from a less refined value by using the
`Value.Refine` method to obtain a _refinement builder_, which uses the
builder pattern to construct a new value with one or more extra refinements.

```go
val := cty.UnknownVal(cty.String).Refine().
    NotNull().
    StringPrefix("https://").
    NewValue()
```

The above snippet would produce a refined local value whose range is limited
only to non-null strings which start with the prefix `"https://"`. This
information can, in theory, allow `val.Equals(cty.NullVal(cty.String))` to
return `cty.False` rather than `cty.UnknownVal(cty.Bool)`, and allow a prefix
match against the string to return a known result.

In practice not all operations against unknown values can make full use of
unknown value refinements, but hopefully the coverage will increase over time.

Only unknown values can have refinements, because known values are already
refined by their concrete value: simple values like `cty.Zero` are constrained
to exactly one value, while some values like `cty.ListValEmpty(cty.DynamicPseudoType)`
represent a set of possible values -- all empty lists of any element type, in
this case.

However, the `Refine` operation _is_ also supported for known values and in that
case acts as a self-checking assertion that the known value does actually
meet the requirements. If you write your codepaths to unconditionally assign
refinements regardless of whether the value is known then your code will
self-check and raise a panic if the final known value doesn't match the
previously-promised refinements.

A similar rule applies to applying new refinements already-refined values: it's
fine to describe a less specific refinement, which will therefore be ignored
because it adds no new information. It's an application bug to describe a
contradictory refinement, such as a new string prefix that doesn't match one
previously assigned.

## Value ranges

The `Refine()` method described above constructs a value with refinements. To
access the information from those refinements, use the `Value.Range` method to
obtain a `cty.ValueRange` object, which describes a superset of all of the
values that a particular value could have.

For example, you can use `val.Range().DefinitelyNotNull()` to test whether a
particular value is guaranteed to be non-null once it is finally known. This
again works for both known and unknown values, so e.g.
`cty.StringVal("foo").Range().DefinitelyNotNull()` returns `true` because
a known, non-null string value is _definitely not null_.

When writing operations that depend only on information that can be determined
from refinements it's valid to depend exclusively on `Value.Range` and rely on
the fact that the range of an already-known value is just a very narrow range
that covers only what that specific value covers.

The model of value ranges is imprecise, though: it's limited only to information
we can track for unknown values through refinements. Many operations will still
need a special codepath to handle the unknown case vs. the known case so they
can take into account the additional detail from the exact value once known.

## Available Refinements

The set of possible refinement types might grow over time, but the initial set
is focused on a narrow set of possibilities that seems likely to allow a number
of other operations to either produce known results from unknown input or to
rule that particular input is invalid despite not yet being known.

The most notable restriction on refinements is that the available refinements
vary depending on the type constraint of the value being refined.

The least flexible case is `cty.DynamicVal` -- an unknown value of an unknown
type -- which is the one value that cannot be refined at all and will cause
a panic if you try. This is a pragmatic compromise for backward compatibility:
existing callers use patterns like `val == cty.DynamicVal` to test for this
specific special value, and any refinements of that value would make it no
longer equal.

Unknown values of built-in exact types, and also unknown values whose type
_kind_ is constrained even if the element/attribute types are not, can at
least be refined as being non-null, and because that is a common situation
there is a shorthand for it which avoids using the builder pattern:
`val.RefineNotNull()`.

All other possible refinements are type-constraint-specific:

* `cty.String`

    For strings we can refine a known prefix of the string, which is intended
    for situations where the string represents some microsyntax with a
    known prefix, such as a URL of a particular known scheme.

    * `.StringPrefix(string)` specifies a known prefix of the final string.

        By default an unknown string has no known prefix, which is the same
        as the prefix being the empty string.

        Because `cty`'s model of strings is a sequence of Unicode grapheme
        clusters, `.StringPrefix` will quietly disregard trailing Unicode
        code units of the given prefix that might combine with other code
        units to form a new combined grapheme. This is a good safe default
        behavior for situations where the remainder of the string is under
        end-user control and might begin with combining diacritics or
        emoji variation sequences. Applications should not rely on the
        details of this heuristic because it may become more precise in
        later releases.

    * `.StringPrefixFull(string)` is like `.StringPrefix` but does not trim
      possibly-combining code units from the end of the given string.

        Applications must use this with care, making sure that they control
        the final string enough to guarantee that the subsequent additional
        code units will never combine with any characters in the given prefix.

* `cty.Number`

    For numbers we can refine both the lower and upper bound of possible values,
    with each boundary being either inclusive or exclusive.

    * `.NumberRangeLowerBound(cty.Value, bool)` refines the lower bound of
      possible values for an unknown number. The boolean argument represents
      whether the bound is _inclusive_.

        The given value must be a non-null `cty.Number` value. An unrefined
        number effectively has a lower bound of `(cty.NegativeInfinity, true)`.

    * `.NumberRangeUpperBound(cty.Value, bool)` refines the upper bound of
      possible values for an unknown number. The boolean argument represents
      whether the bound is _inclusive_.

        The given value must be a non-null `cty.Number` value. An unrefined
        number effectively has an upper bound of `(cty.PositiveInfinity, true)`.

    * `.NumberRangeInclusive(min, max cty.Value)` is a helper wrapper around
      the previous two methods that declares both an upper and lower bound
      at the same time, while specifying that both are inclusive bounds.

* `cty.List`, `cty.Set`, and `cty.Map` types

    For all collection types we can refine the lower and upper bound of the
    length of the collection. The boundaries on length are always inclusive
    and are integers, because it isn't possible to have a fraction of an
    element.

    * `.CollectionLengthLowerBound(int)` refines the lower bound of possible
      lengths for an unknown collection.

        An unrefined collection effectively has a lower bound of zero, because
        it's not possible for a collection to have a negative length.

    * `.CollectionLengthUpperBound(int)` refines the upper bound of possible
      lengths for an unknown collection.

        An unrefined collection has an upper bound that matches the largest
        valid Go slice index on the current platform, because `cty`'s
        collections are implemented in terms of Go's collection types.
        However, applications should typically not expose that specific value
        to users (it's an implementation detail) and should instead present
        the maximum value as an unconstrained length.

    * `.CollectionLength(int)` is a shorthand that refines both the lower and
      upper bounds to the same value. This is a helpful requirement to make
      whenever possible because it will often allow the final value to be
      a known collection with unknown elements, as described in
      [Refinement Value Collapse](#refinement-value-collapse).

Some built-in operations will automatically take into account refinements from
their input operands and propagate them in a suitable way to the result.
However, that is not a guarantee for all operations and so should be treated
as a "best effort" behavior which will hopefully become more precise in future
versions.

Behaviors implemented in downstream applications, such as custom functions
using [the function system](functions.md), might also take into account
refinements. If they do their work using only _operation methods_ on `Value`
then the handling of refinements might come for free. If they do work using
_integration methods_ instead then they will need to explicitly handle
refinements if desired. If they don't then by default the result from an
unknown input will be a totally-unrefined unknown value, though will hopefully
still have a useful type constraint.

## Refinement Value Collapse

For some kinds of refinement it's possible to constrain the range so much that
only one possible value remains. In that case, the `.NewValue()` method of the
refinement builder might return a known value instead of an unknown value.

For example, if the lower bound and upper bound of a collection's length are
equal then the length of the collection is effectively known. For some lengths
of some collection kinds the refinement can collapse into a known collection
containing unknown values. For example, an unknown list that's known to have
exactly two values can be represented equivalently as a known list of length
two where both elements are unknown themselves.

The exact details of how refinement collapse is decided might change in future
versions, but only in ways that can make results "more known". It would be a
breaking change to weaken a rule to produce unknown values in more cases, so
that kind of change would be reserved only for fixing an important bug or
design error.

## Refinements are Dynamic Only

Refinements belong to unknown values rather than to type constraints, and so
refining an unknown value does not change its type constraint.

This design is a tradeoff: making the refinements dynamic and implicit means
that it's possible to add more detailed refinements over type without making
breaking changes to explicit type information, but the downside is that
it isn't possible to represent refinements in any situation that is only
aware of types.

For example, it isn't currently possible to represent the idea of an unknown
map whose elements each have a further refinement applied, because the
refinements apply to the map itself and there are not yet any specific element
values for the element refinements to attach to.

(It would be possible in theory to allow refining an unknown collection with
meta-refinements about its hypothetical elements, but that is not currently
supported because it would mean that refinements would need to be resolved
recursively and that would be considerably more complex and expensive than
the current single-value-only refinements structure.)

## Refinements Under Serialization

Refinements are intentionally designed so that they only constrain the range
of an unknown value, and never expand it. This means that it should typically
be safe to discard refinements in situations like serialization where there
may not be any way to represent the refinements. After decoding the unknown
value now has a wider range but it should still be a superset of the true
range of the value. This is an example of the general rule that no operation
on an unknown value is _guaranteed_ to fully preserve the input refinements
or to consider them when calculating the result.

The official MessagePack serialization in particular does have some support
for retaining approximations of refinements as part of its serialization of
unknown values, using a MessagePack extension value. Some detail may still
be lost under round-tripping but the output range should always be a superset
of the input range. As long as both the serializer and deserializer are using
the `cty/msgpack` sub-package unknown values will propagate automatically
without any additional caller effort.

## Relationship to "Marks"

The idea of annotating a value with additional information has some overlap
with the concept of [Marks](marks.md). However, the two have different purposes
and so different design details and tradeoffs.

Marks should typically be used for additional information that is independent
of the specific type and value, such as marking a value as having come from
a sensitive location. The marking then propagates to all results from operations
on that value, usually without changing the behavior of that operation. In a
sense the mark represents the _origin_ of the value rather than the value
itself.

Refinements are instead directly part of the value. By reducing the possible
range of an unknown value placeholder, other downstream operations can in turn
produce a more refined result, or possibly even a known result from unknown
inputs. Refinements do not naively propagate from one value to the next, but
some operations will use the refinements of their operands to calculate a new
set of refiments for their result, with the rules varying on a case-by-case
basis depending on what calculation the operation represents.
