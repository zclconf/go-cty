# `cty` backward-compatibility policy

This library includes a number of behaviors that aim to support "best effort"
partial evaluation in the presence of wholly- or partially-unknown inputs.
Over time we've improved the accuracy of those analyses, but doing so changes
the specific results returned by certain operations.

This document aims to describe what sorts of changes are allowed in new minor
releases and how those changes might affect the behavior of dependents after
upgrading.

Where possible we'll avoid making changes like these in _patch_ releases, which
focus instead only on correcting incorrect behavior. An exception would be if
a minor release introduced an incorrect behavior and then a patch release
repaired it to either restore the previous correct behavior or implement a new
compromise correct behavior.

## Unknown Values can become "more known"

The most significant policy is that any operation that was previously returning
an unknown value may return either a known value or a _more refined_ unknown
value in later releases, as long as the new result is a subset of the range
of the previous result.

When using only the _operation methods_ and functionality derived from them,
`cty` will typically handle these deductions automatically and return the most
specific result it is able to. In those cases we expect that these changes will
be seen as an improvement for end-users, and not require significant changes
to calling applications to pass on those benefits.

When working with _integration methods_ (those which return results using
"normal" Go types rather than `cty.Value`) these changes can be more sigificant,
because applications can therefore observe the differences more readily.
For example, if an unknown value is replaced with a known value of the same
type then `Value.IsKnown` will begin returning `true` where it previously
returned `false`. Applications should be designed to avoid depending on
specific implementation details like these and instead aim to be more general
to handle both known and unknown values.

A specific sensitive area for compatibility is the `Value.RawEquals` method,
which is sensitive to all of the possible variations in values. Applications
should not use this method for normal application code to avoid exposing
implementation details to end-users, but might use it to assert exact expected
results in unit tests. Such test cases may begin failing after upgrading, and
application developers should carefully consider whether the new results conform
to these rules and update the tests to match as part of their upgrade if so. If
the changed result seems _not_ to conform to these rules then that might be a
bug; please report it!

## Error situations may begin succeeding

Over time the valid inputs or other constraints on functionality might be
loosened to support new capabilities. Any operation or function that returned
an error in a previous release can begin succeeding with any valid result in
a new release.

## Error message text might change

This library aims to generate good, actionable error messages for user-facing
problems and to give sufficient information to a calling application to generate
its own high-quality error messages in situations where `cty` is not directly
"talking to" an end-user.

This means that in later releases the exact text of error messages in certain
situations may change, typically to add additional context or increase
precision.

If a function is documented as returning a particular error type in a certain
situation then that should be preserved in future releases, but if there is
no explicit documentation then calling applications should not depend on the
dynamic type of any `error` result, or should at least do so cautiously with
a fallback to a general error handler.

## Passing on changes to Go standard library

Some parts of `cty` are wrappers around functionality implemented in the Go
standard library. If the underlying packages change in newer versions of Go
then we may or may not pass on the change through the `cty` API, depending on
the circumstances.

A specific notable example is Unicode support: this library depends on various
Unicode algorithms and data tables indirectly through its dependencies,
including some in the Go standard library, and so its exact treatment of strings
is likely to vary between releases as the Unicode standard grows. We aim to
follow the version of Unicode supported in the latest version of the Go standard
library, although we may lag behind slightly after new Go releases due to the
need to update other libraries that implement other parts of the Unicode
specifications.
