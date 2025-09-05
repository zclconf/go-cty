# 1.17.0 (September 5, 2025)

`cty` now requires Go 1.23 or later.

- `cty.Value.Elements` offers a modern `iter.Seq2`-based equivalent of `cty.Value.ElementIterator`.
- `cty.DeepValues` offers a modern `iter.Seq2`-based equivalent of `cty.Walk`.
- `cty.Value.WrangleMarksDeep` allows inspecting and modifying individual marks throughout a possibly-nested data structure.

  Having now got some experience using marks more extensively in some callers, it's become clear that it's often necessary for different subsystems to be able to collaborate using independent marks without upsetting each other's assumptions. Today that tends to be achieved using hand-written transforms either with `cty.Transform` or `cty.Value.UnmarkDeepWithPaths`/`cty.Value.MarkWithPaths`, both of which can be pretty expensive even in the common case where there are no marks present at all.

  This new function allows inspecting and transforming marks with far less overhead, by creating new values only for parts of a structure that actually need to change and by reusing (rather than recreating) the "payloads" of the values being modified when we know that only the marks have changed.

- `cty.ValueMarksOfType` and `cty.ValueMarksOfTypeDeep` make it easier to use type-based rather than value-based mark schemes, where different values of a common type are used to track a specific kind of relationship with multiple external values.
- `cty.Value.HasMarkDeep` provides a "deep" version of the existing `cty.Value.HasMark`, searching throughout a possibly-nested structure for any values that have the given mark.
- `cty.Value.UnmarkDeep` and `cty.Value.UnmarkDeepWithPaths` are now implemented in terms of `cty.Value.WrangleMarksDeep`, so they benefit from its reduced overhead. In particular they avoid reconstructing a data structure that contains no marked values at all.
- `cty.Value.MarkWithPaths` now has a fast path when it's given a zero-length `PathValueMarks`, in which case it just returns the value it was given with no modifications.

# 1.16.4 (August 20, 2025)

* `cty.UnknownAsNull` now accepts marked values and preserves the given marks in its result. Previously it had no direct support for marks and so would either panic or return incorrect results when given marked values.

# 1.16.3 (May 16, 2025)

* `convert`: Now generates more specific error messages in various cases of type conversion failure, giving additional information about the type that was given as compared to the type that was wanted by the caller.

# 1.16.2 (January 21, 2025)

* `json`: `ImpliedType` now returns an error if a JSON object contains two properties of the same name. As a compatibility concession it allows duplicates whose values have the same implied type, since it was unintentionally possible to combine `ImpliedType` and `Unmarshal` successfully in that case before, but this is not an endorsement of using duplicate property names since that makes the input ambiguous in any case. ([#199](https://github.com/zclconf/go-cty/issues/199))
* `function/stdlib`: `ElementFunc` no longer crashes when asked for a negative index into a tuple. This fixes a miss in the negative index support added back in v1.15.0. ([#200](https://github.com/zclconf/go-cty/pull/200))

# 1.16.1 (January 13, 2025)

* `cty`: `Value.HasElement` now treats unknown set elements consistently with how much of the rest of `cty` treats them.
* `function/stdlib`: `FormatFunc` and `FormatListFunc` now handle unknown and null values of unknown type as arguments, rather than letting the function system's short-circuit behavior take care of it. This allows `cty.DynamicVal` and `cty.NullVal(cty.DynamicPseudoType)` to be treated consistently with other values, returning results consistent with the documented behavior, rather than forcing the function to immediately return `cty.DynamicVal`.

# 1.16.0 (January 3, 2025)

* `convert`: When converting between two different capsule types, will now try to use the "conversion _from_" implementation from the target type if the source type doesn't have a suitable "conversion _to_" implementation. ([#194](https://github.com/zclconf/go-cty/pull/194))
* `convert`: When converting to a map whose element type is an object type with optional attributes, will no longer construct a broken result when a final map is empty. ([#198](https://github.com/zclconf/go-cty/pull/198))

# 1.15.1 (November 26, 2024)

* `function`: Function calls can now return more mark-related information when called with unknown values when neither `AllowMarks` nor `AllowUnknown` are set for a particular parameter. ([#196](https://github.com/zclconf/go-cty/pull/196))

# 1.15.0 (July 15, 2024)

* `function/stdlib`: The `element` function now accepts negative indices, extending the illusion of an infinitely-long list into the negative direction too.
* `cty`: The various "transform" functions were previously incorrectly propagating marks from a parent object down to attribute values when calling the caller-provided transform functions. The marks will now no longer be propagated downwards, which is consistent with the treatment of collection and tuple elements. If your transform function needs to take into account context about marks of upstream containers then you can maintain a stack of active marks in your `Transformer` implementation, pushing in `Enter` and popping in `Exit`. ([#185](https://github.com/zclconf/go-cty/pull/185))

# 1.14.4 (March 20, 2024)

* `msgpack`: Now uses string encoding instead of float encoding for a whole number that is too large to fit in any of MessagePack's integer types.
* `function/stdlib`: Type conversion functions (constructed with `MakeToFunc`) can now convert null values of unknown type into null values of the target type, rather than returning an unknown value in that case.
* `json`: Will now correctly reject attempts to encode `cty.DynamicVal`, whereas before it would just produce an invalid JSON document without any error. (This is invalid because JSON encoding cannot support unknown values at all; `cty.DynamicVal` is a special case of unknown value where even the _type_ isn't known.)

# 1.14.3 (February 29, 2024)

* `msgpack`: Fixed edge-case bug that could cause loss of floating point precision when round-tripping due to incorrectly using a MessagePack integer to represent a large non-integral number. [#176](https://github.com/zclconf/go-cty/pull/176)
* `cty`: Fixed some false-negative numeric equality test results by comparing numbers as integers when possible. [#176](https://github.com/zclconf/go-cty/pull/176)

# 1.14.2 (January 23, 2024)

* `convert`: Converting from an unknown map value to an object type now correctly handles the situation where the map element type disagrees with an _optional_ attribute of the target type, since when a map value is unknown we don't yet know which keys it has and thus cannot predict what subset of the elements will get converted as attributes in the resulting object. ([#175](https://github.com/zclconf/go-cty/pull/175))

# 1.14.1 (October 5, 2023)

* `cty`: It's now valid to use the `Refine` method on `cty.DynamicVal`, although all refinements will be silently discarded. This replaces the original behavior of panicking when trying to refine `cty.DynamicVal`.
* `cty`: `Value.Range` will now return a clearer panic message if called on a marked value. The "value range" concept is only applicable to unmarked values because not all of the `ValueRange` functions are able to propagate marks into their return values, due to returning Go primitive types instead of new `cty.Value` results.

    Callers that use marks must, as usual, take care to unmark them before exporting values into "normal" Go types, and then explicitly re-apply the marks to their result as appropriate. Applications that make no use of value marks, and library callers that exclude marked values from what they support, can safely ignore this requirement.

# 1.14.0 (August 30, 2023)

This release updates the supported version of Unicode from Unicode 13 to Unicode 15. This is a backwards-compatible change that means that cty supports normalization and segmentation of strings containing new Unicode characters. The algorithms for normalization and segmentation themselves are unchanged.

If you use `cty` in an application that cares about consistent Unicode support, you should upgrade to Go 1.21 at the same time as updating to `cty` v1.14, because that will then also update the Unicode tables embedded in the Go standard library (used for case folding, etc).

* `cty`: The `cty.String` type will now normalize incoming string values using the Unicode 15 normalization rules.
* `function/stdlib`: The various string functions which split strings into individual characters as part of their work will now use the Unicode 15 version of the text segmentation algorithm to do so.

# 1.13.3 (August 24, 2023)

* `msgpack`: As a compromise to avoid unbounded memory usage for a situation that some callers won't take advantage of anyway, the MessagePack decoder has a maximum length limit on encoded unknown value refinements. For consistency, the encoder will now truncate string prefix refinements if necessary to avoid making the encoded refinements too long. ([#167](https://github.com/zclconf/go-cty/pull/167))

    This is consistent with the documented conventions for serializing refinements -- that we can potentially lose detail through serialization -- but in this case we are still able to preserve shorter string prefixes, whereas other serializations tend to just discard refinement information altogether.

# 1.13.2 (May 22, 2023)

* `cty`: `IndexStep.Apply` will no longer panic if given a marked collection to traverse through. ([#160](https://github.com/zclconf/go-cty/pull/160)).

# 1.13.1 (March 16, 2023)

* `function`: If a function parameter that doesn't declare `AllowDynamicType: true` recieves a `cty.DynamicVal`, the function system would previously just skip calling the function's `Type` callback and treat the result type as unknown. However, the `Call` method was then still calling a function's `Impl` callback anyway, which violated the usual contract that `Type` acts as a guard for `Impl` so `Impl` doesn't have to repeat type-checking already done in `Type`: it's only valid to call `Impl` if `Type` was previosly called _and_ it succeeded.

    The function system will now skip calling `Impl` if it skips calling `Type`, immediately returning `cty.DynamicVal` in that case. Individual functions can opt out of this behavior by marking one or more of their parameters as `AllowDynamicType: true` and then handling that situation manually inside the `Type` and `Impl` callbacks.

    As a result of this problem, some of the `function/stdlib` functions were not correctly handling `cty.DynamicVal` arguments after being extended to support refinements in the v1.13.0 release, causing unexpected errors or panics when calling them. Those functions are fixed indirectly by this change, since their callbacks will no longer run at all in those cases, as was true before they were extended to support refinements.

# 1.13.0 (February 23, 2023)

## Upgrade Notes

This release introduces a new concept called [Refinements](./docs/refinements.md), which allow `cty` to constrain the range of an unknown value beyond just a type constraint and then make deductions about validity or result range based on those refinements.

These changes are consistent with [the backward-compatibility policy](COMPATIBILITY.md) but you may see some changed results in your unit tests of operations involving unknown values. If the new results don't seem like valid refinements of what was previously being returned in the v1.12 series, please open an issue to discuss that.

If the new results have a range that is a valid subset of the old results then that is expected behavior and you should update your tests as part of upgrading.

## Other changes in this release

* Refinements: `cty` will can track a refined range for some unknown values and will take those into account when evaluating certain operations, thereby allowing a "more known" result than before. ([#153](https://github.com/zclconf/go-cty/pull/153))
* `function/stdlib`: The `FormatDate` and `TimeAdd` functions in previous releases were accidentally more liberal than intended in their interpretation of timestamp strings documented as requiring RFC3339. ([#152](https://github.com/zclconf/go-cty/pull/152))

    Those functions are now corrected to use a stricter RFC3339 parser, meaning that they will now reject some inputs that were previously accepted but were not valid per the RFC3339 syntax rules. The documentation for these functions already specified that RFC3339 syntax was required and so this is a fix to a defect rather than a breaking change, but calling applications which embed these functions may wish to pass on an upgrade note about this behavior difference in their own releaase notes after upgrading.

# 1.12.1 (November 8, 2022)

* `convert`: Will now produce correct type constraints when the input value is an empty collection and the target element type has optional attributes. In this case the conversion process must remove the optional attribute annotations because those are only for type conversion purposes and have no meaning when used in the type constraint for an empty collection. ([#143](https://github.com/zclconf/go-cty/pull/143))
* `convert`: Will now prefer to retain a concrete type in the input value when the input is either null or unknown and the target type is `cty.DynamicPseudoType`, which represents "any type". ([#144](https://github.com/zclconf/go-cty/pull/144))

# 1.12.0 (October 27, 2022)

* `function`: Each function can now have an English-language description summarizing its behavior. This is intended as a default string to use when an application wants to provide code hover tips or similar development aids. However, these descriptions are basic and only available in English, so applications may still prefer to provide their own descriptions and ignore those encoded in this module. ([#137](https://github.com/zclconf/go-cty/pull/137))
* `convert`: When running in "unsafe mode" (which allows additional conversions that can potentially fail with certain input values), we'll now allow converting from a map type to an object type with optional attributes as long as all of the _present_ map elements are compatible with their corresponding optional attributes.

    It's still a dynamic error to convert a map whose element type is incompatible with any of the attributes that _do_ have corresponding keys in the given map. ([#139](https://github.com/zclconf/go-cty/pull/139))
* `convert`: Will now produce correct type constraints when the input value is null and the target type has optional attributes. In this case the conversion process must remove the optional attribute annotations because those are only for type conversion purposes and have no meaning when used in the type constraint for a null or unknown value. ([#140](https://github.com/zclconf/go-cty/pull/140), [#141](https://github.com/zclconf/go-cty/pull/141))

# 1.11.1 (October 17, 2022)

* `convert`: Fix for error when converting empty sets and lists with nested optional attributes by explicitly removing optional attribute information from collections.

# 1.11.0 (August 22, 2022)

## Upgrade Notes

This release contains some changes to some aspects of the API that are either legacy or de-facto internal (from before the Go toolchain had an explicit idea of that). Any external module using these will experience these as breaking changes, but we know of no such caller and so are admitting these without a major release in the interests of not creating churn for users of the main API.

* **`encoding/gob` support utilities removed**: we added these as a concession to HashiCorp who wanted to try to send `cty` values over some legacy protocols/formats used by legacy versions of HashiCorp Terraform. In the end those efforts were not successful for other reasons and so no Terraform release ever relied on this functionality.

    `encoding/gob` support has been burdensome due to how its unmarshaler interface is defined and so `cty` values and types are no longer automatically compatible with `encoding/gob`. Callers should instead use explicitly-implemented encodings, such as the built-in JSON and msgpack encodings or external libraries which use the public `cty` API to encode and decode.
* **cty now requires Go 1.18**: although the main API is not yet making any use of type parameters, we've begun to adopt it in the hope of improving the maintainability of some internal details, starting with the backing implementation of set types.

    Since type parameters are not supported by earlier versions of the Go compiler, callers must upgrade to Go 1.18 before using cty v1.11.0 or later.

## Other changes in this release

* `cty`: Improved performance when comparing nonzero numbers to zero, by performing a relatively-cheap sign check on both numbers before falling back on the more expensive general equality implementation. ([#125](https://github.com/zclconf/go-cty/pull/125))
* `cty`: It's now possible to use capsule types in the elements of sets. Previously `cty` would panic if asked to construct a value of a set type whose element type either is or contains a capsule type, but there is now explicit support for storing encapsulated values in sets and optional (but recommended) support for a custom hashing function per type in order to improve performance for sets with a large number of elements.
* `convert`: Unify will no longer panic when asked to find a common base type for a tuple type and a list of unknown element type, and will instead just signal that such a unification is not possible. ([#126](https://github.com/zclconf/go-cty/pull/126))
* `stdlib`: `FlattenFunc` will no longer panic if it encounters a null value of a type that would normally be subject to flattening. Instead, it will treat it in the same way as a null value of any non-flattenable type. ([#129](https://github.com/zclconf/go-cty/pull/129))

# 1.10.0 (November 2, 2021)

* `cty`: The documented definition and comparison logic of `cty.Number` is now refined to acknowledge that its true range is limited only to values that have both a binary floating point and decimal representation, because `cty` values are primarily designed to traverse JSON serialization where numbers are always defined as decimal strings.

    In particular, that means that two `cty.Number` values now always compare as equal if their representation in JSON (under `cty`'s own JSON encoder) would be equal, even though the decimal approximation we use for that conversion is slightly lossy. This pragmatic compromise avoids confusing situations where a round-trip through JSON serialization (or other serializations that use the same number format) may produce a value that doesn't compare equal to the original.
    
    This new definition of equals should not cause any significant behavior change for any integer in our in-memory storage range, but may cause some fractional values to compare equal where they didn't before if they differ only by a small fraction.
* `cty`: Don't panic in `Value.Equals` if comparing complex data structures with nested marked values. Instead, `Equals` will aggregate all of the marks on the resulting boolean value as we typically expect for operations that derived from marked values. ([#112](https://github.com/zclconf/go-cty/pull/112))
* `cty`: `Value.AsBigFloat` now properly isolates its result from the internal state of the associated value. It previously _attempted_ to do this (so that modifying the result would not affect the supposedly-immutable `cty.Number` value) but ended up creating an object which still had some shared buffers. The result is now entirely separate from the internal state of the recieving value. ([#114](https://github.com/zclconf/go-cty/pull/114))
* `function/stdlib`: The `FormatList` function will now return an unknown value if any of the arguments have an unknown type, because in that case it can't tell whether that value will ultimately become a string or a list of strings, and thus it can't predict how many elements the result will have. ([#115](https://github.com/zclconf/go-cty/pull/115))

# 1.9.1 (August 17, 2021)

* `cty`: Don't panic in `Value.Equals` if comparing complex data structures with nested marked values. Instead, `Equals` will aggregate all of the marks on the resulting boolean value as we typically expect for operations that derived from marked values. ([#112](https://github.com/zclconf/go-cty/pull/112))
* `cty`: `Value.AsBigFloat` now properly isolates its result from the internal state of the associated value. It previously _attempted_ to do this (so that modifying the result would not affect the supposedly-immutable `cty.Number` value) but ended up creating an object which still had some shared buffers. The result is now entirely separate from the internal state of the recieving value. ([#114](https://github.com/zclconf/go-cty/pull/114))
* `function/stdlib`: The `FormatList` function will now return an unknown value if any of the arguments have an unknown type, because in that case it can't tell whether that value will ultimately become a string or a list of strings, and thus it can't predict how many elements the result will have. ([#115](https://github.com/zclconf/go-cty/pull/115))

# 1.9.0 (July 6, 2021)

* `cty`: `cty.Walk`, `cty.Transform`, and `cty.TransformWithTransformer` now all correctly support marked values. Previously they would panic when encountering marked collections, because they would try to recurse into them without handling the markings.
* `function/stdlib`: The `floor` and `ceil` functions no longer lower the precision of arguments to what would fit inside a 64-bit float, instead preserving precision in a similar way as most other arithmetic functions. ([#111](https://github.com/zclconf/go-cty/pull/111))
* `function/stdlib`: The `flatten` function was incorrectly treating null values of an unknown type as if they were unknown values. Now it will treat them the same as any other non-list/non-tuple value, flattening them down into the result as-is. ([#110](https://github.com/zclconf/go-cty/pull/110))

# 1.8.4 (June 22, 2021)

* `function/stdlib`: The `flatten` function will now correctly return `cty.DynamicVal` if it encounters `cty.DynamicVal` anywhere in the given data structure, because it can't predict how many elements the result will have in that situation. ([#106](https://github.com/zclconf/go-cty/pull/106), [#107](https://github.com/zclconf/go-cty/pull/107))
* `function/stdlib`: The `setproduct` function will no longer panic when given a set containing unknown values, which would therefore be a set with an unknown length. ([#109](https://github.com/zclconf/go-cty/pull/109))

# 1.8.3 (May 4, 2021)

* `function/stdlib`: Fix a panic in `SetproductFunc` in situations where one of the input collections is empty. ([#103](https://github.com/zclconf/go-cty/pull/103))
* `function/stdlib`: Improvements to `ElementFunc`, `ReverseListFunc`, and `SliceFunc` to handle marked values more precisely (individual element vs. whole-collection marks). ([#101](https://github.com/zclconf/go-cty/pull/101))

# 1.8.2 (April 20, 2021)

* `cty`: `Value.Mark` will no longer incorrectly create nested markings when applied to a value that is already marked. Instead, it will unpack the reciever and use its underlying value directly, merging all of the marks into a new mark set. ([#96](https://github.com/zclconf/go-cty/pull/96))
* `cty:` `Value.RawEquals` will no longer panic if asked to compare two maps where at least one of them is marked. ([#96](https://github.com/zclconf/go-cty/pull/96))
* `function/stdlib`: Improvements to `ChunklistFunc`, `ConcatFunc`, `FlattenFunc`, `KeysFunc`, `LengthFunc`, `LookupFunc`, `MergeFunc`, `SetproductFunc`, `ValuesFunc`, and `ZipmapFunc` to handle marked values more precisely (individual element vs. whole-collection marks). ([#94](https://github.com/zclconf/go-cty/pull/94), [#95](https://github.com/zclconf/go-cty/pull/95), [#96](https://github.com/zclconf/go-cty/pull/96), [#97](https://github.com/zclconf/go-cty/pull/97), [#98](https://github.com/zclconf/go-cty/pull/98), [#99](https://github.com/zclconf/go-cty/pull/99), [#100](https://github.com/zclconf/go-cty/pull/100))

# 1.8.1 (March 16, 2021)

* `convert`: Fix for panics and some general misbehavior when converting null values to type constraints containing objects with optional attributes. ([#88](https://github.com/zclconf/go-cty/pull/88))
* `convert`: Type unification of a mixture of list and tuple types and for a mixture of map and object types will now do the same recursive unification that we previously did for unification of just list types and just map types respectively, to avoid producing a very different and confusing result in situations where callers try to construct collections from a mixture of nested collections and nested structural types. ([#89](https://github.com/zclconf/go-cty/pull/89))
* `convert`: Conversion will no longer panic if we can't find a suitable single element type to use when converting to a collection type with a dynamically-selected element type. ([#91](https://github.com/zclconf/go-cty/pull/91))
* `function`: The `ReturnTypeForValues` and `Call` methods on `Function` will now protect functions from having to deal with nested marked values for arguments that don't specifically declare `AllowMarks: true`, as a concession for the fact that many functions were written prior to the introduction of marks as a concept. ([#92](https://github.com/zclconf/go-cty/pull/92))

# 1.8.0 (February 22, 2021)

* `cty`: When running on Go 1.16 or later, the `cty.String` type will now normalize incoming string values using the Unicode 13 normalization rules.
* `function/stdlib`: The various string functions which split strings into individual characters as part of their work will now use the Unicode 13 version of the text segmentation algorithm to do so.

# 1.7.2 (February 22, 2021)

* `cty`: The `Type.GoString` implementation for object types with optional attributes was previously producing incorrect results due to an implementation bug. ([#86](https://github.com/zclconf/go-cty/pull/86))

# 1.7.1 (December 15, 2020)

* `cty`: The `Value.Multiply` and `Value.Modulo` functions now correctly propagate the floating point precision of the arguments, which avoids generating incorrect results for large integer operands. ([#75](https://github.com/zclconf/go-cty/pull/75))
* `convert`: The `convert.MismatchMessage` function will now correctly identify mismatching attributes in objects, rather than misreporting attributes that are actually present and correct. ([#78](https://github.com/zclconf/go-cty/pull/78))
* `function/stdlib`: The `merge` function now returns an empty object if all of its arguments are `null`, rather than returning `null` as before. That's more consistent with its usual behavior of ignoring `null` arguments when there is at least one non-null argument. ([#82](https://github.com/zclconf/go-cty/pull/82))
* `function/stdlib`: The `coalescelist` function now ignores any arguments that are null, rather than panicking as before.. ([#81](https://github.com/zclconf/go-cty/pull/81))

# 1.7.0 (October 20, 2020)

* `cty`: `Value.UnmarkDeepWithPaths` and `Value.MarkWithPaths` are like `Value.UnmarkDeep` and `Value.Mark` but they retain path information for each marked value, so that marks can be re-applied later without all the loss of detail that results from `Value.UnmarkDeep` aggregating together all of the nested marks.
* `function`: Unless a parameter has `AllowMarks: true` explicitly set, the functions infrastructure will now guarantee that it never sees a marked value even if the mark is deep inside a data structure. Previously that guarantee was only shallow for the top-level value, similar to `AllowUnknown`, but because marks are a relatively new addition to `cty` and numerous existing functions are not written to deal with them this is the more conservative and robust default. ([#72](https://github.com/zclconf/go-cty/pull/72))
* `function/stdlib`: The `formatdate` function was not correctly handling literal sequences at the end of the format string. It will now handle those as intended. ([#69](https://github.com/zclconf/go-cty/pull/69))

# 1.6.1 (September 2, 2020)

* `cty`:: Fix a regression from 1.6.0 where `Value.RawEqual` no longer returned the correct result given a pair of sets containing partially-unknown values. ([#64](https://github.com/zclconf/go-cty/pull/64))

# 1.6.0 (August 30, 2020)

* Fixed various defects in the handling of sets containing unknown values. This will cause unknown values to now be returned in more situations, whereas before `cty` would often return incorrect results when working with sets containing unknown values. The list of defects fixed in this release includes:
    - `cty`: The length of a set containing unknown values, as defined by `Value.Length`, is itself unknown, reflecting the fact that unknown values may be placeholders for values that are equal to other values in the set, which would thus coalesce into a single value.
    - `cty:` Converting a set with unknown values to a list produces an unknown value, because type conversion can't predict which indices each element of the set should take (the unknown elements could appear anywhere in the sort order) or the length of the resulting list.
    - `function/stdlib`: the `LengthFunc` and `ToList` functions wrap the behaviors described in the previous two items and are therefore also fixed in the same way.
    - `function/stclib`: `FormatListFunc` depends on knowing the length of all of its sequence arguments (which includes support for sets), so it will return an unknown result if given a set with an unknown length.
    - `function/stdlib`: The various set operation functions were previously producing incorrect results if one of their given sets contained unknown values, because they didn't consider that unknown values on one set may be placeholders for values that are equal to elements of the other set. For example, `SetSubtractFunc` now produces a wholly-unknown result if either of its arguments contains an unknown element, because it can't predict whether that unknown element represents a value equal to an element in the other set.
    - `cty`: The `Value.Equal` function would previously incorrectly return a known `cty.False` if one of the given sets contained an unknown value. It will now return `cty.UnknownVal(cty.Bool)` in that case, reflecting that the result could be either `cty.True` or `cty.False` were the unknown values to be replaced with known values.
    - `cty`: The `Value.LengthInt` function was also returning incorrect results for sets containing unknown elements. However, given that it is commonly used in conjunction with `ElementIterator` to determine the capacity for a slice to append elements to, it is not fixed and is instead redefined to return the _maximum possible length_, which would result if all of the unknown values represent values that are not equal to any other set element. Applications that use `Value.LengthInt` to determine lengths to return to users who are working in the space of `cty` values should switch to using `Value.Length` instead and handle the possibility of the length being unknown, to avoid returning incorrect results for sets with unknown values.

    These are not classified as breaking changes because the previous behavior was defective per the design goals for unknown values. However, callers may notice their application behavior changes along with these fixes when upgrading. The new behaviors should all be more correct than the old; if you observe a change in behavior where there is now an _incorrect_ result for sets containing unknown values (that is, where `cty` claims it knows an answer that it should not actually know), please report that in a GitHub issue.

    We advise callers which work with sets that may potentially contain unknown values to review their own set-handling functions to check if they too might be handling sets with unknown values incorrectly, particularly if they work with sets using [integration methods rather than operation methods](./docs/types.md#common-operations-and-integration-methods) (for example, using `Value.ValueList` or `Value.ValueSet` to extract elements directly). It seems that incorrect handling of sets with unknown values has been a common hazard, particularly in codepaths that aim to treat lists and sets as being interchangable.
* `function/stdlib`: The `element` function will no longer panic if given a negative index. Instead, it will return a proper error. ([#62](https://github.com/zclconf/go-cty/pull/62))
* `convert`: **Experimental** support for annotating one or more attributes of an object type as "optional", which the `convert` package can then use to suppress the error that would normally be returned if the source type has no corresponding attribute, and can substitute a correctly-typed null value instead. This new behavior is subject to change even in minor release of `cty`, until it has been tested in experimental releases of downstream applications and potentially modified in response.

# 1.5.1 (June 25, 2020)

* `function/stdlib`: The `merge` function will no longer panic if all given maps are empty. ([#58](https://github.com/zclconf/go-cty/pull/58))
* `function/stdlib`: The various set-manipulation functions, like `setunion`, will no longer panic if given an unknown set value. ([#59](https://github.com/zclconf/go-cty/pull/59))

# 1.5.0 (June 11, 2020)

* `cty`: New `Value.HasWhollyKnownType` method, for testing whether a value's type could potentially change if any unknown values it was constructed from were to become known. ([#55](https://github.com/zclconf/go-cty/pull/55))
* `convert`: Fix incorrect panic when converting a tuple with a dynamic-typed null member into a list or set, due to overly-liberal type unification. ([#56](https://github.com/zclconf/go-cty/pull/56))

# 1.4.2 (May 29, 2020)

* `function/stdlib`: The `jsonencode` function will now correctly accept a null as its argument, and produce the JSON representation `"null"` rather than returning an error. ([#54](https://github.com/zclconf/go-cty/pull/54))

# 1.4.1 (May 18, 2020)

* `function/stdlib`: Fix various panics related to sets with unknown element types in the set-manipulation functions. ([#52](https://github.com/zclconf/go-cty/pull/52))
* `convert`: Don't panic when asked to convert a tuple of objects to a list type constraint containing a nested `cty.DynamicPseudoType`. ([#53](https://github.com/zclconf/go-cty/pull/53))

# 1.4.0 (April 7, 2020)

* `function/stdlib`: The string functions that partition strings into individual characters (grapheme clusters) now use the appropriate segmentation rules from Unicode 12.0.0, while previous versions used Unicode 9.0.0.
* `function/stdlib`: New functions `Replace` and `RegexReplace` for matching and replacing sequences of characters in a given string with another given string. ([#45](https://github.com/zclconf/go-cty/pull/45))
* `function/stdlib`: The function `Substr` will now produce a zero-length string when given a length of zero. Previously it was incorrectly returning the remainder of the string after the given offset. ([#48](https://github.com/zclconf/go-cty/pull/48))
* `function/stdlib`: The `Floor` and `Ceil` functions will now return an infinity if given an infinity, rather than returning the maximum/minimum integer value. ([#51](https://github.com/zclconf/go-cty/pull/51))
* `cty`: Convenience methods for constructing path index steps from normal Go int and string values. ([#50](https://github.com/zclconf/go-cty/pull/50))

# 1.3.1 (March 3, 2020)

* `convert`: Fix incorrect conversion rules for maps of maps that were leading to panics. This will now succeed in some more cases that ought to have been valid, and produce a proper error if there is no valid outcome. ([#47](https://github.com/zclconf/go-cty/pull/47))
* `function/stdlib`: Fix an implementation error in the `Contains` function that was introduced in 1.3.0, so it will now produce a correct result rather than failing with a confusing error message. ([#46](https://github.com/zclconf/go-cty/pull/46))

# 1.3.0 (February 19, 2020)

* `convert`: There are now conversions from map types to object types, as long as the given map type's element type is convertible to all of the object type's attribute types. ([#42](https://github.com/zclconf/go-cty/pull/42))
* `function/stdlib`: HashiCorp has contributed a number of additional functions to the standard library that were originally implemented directly inside their Terraform codebase: ([#37](https://github.com/zclconf/go-cty/pull/37))
  * `Element`: take an element from a list or tuple by index, using modulo wrap-around.
  * `CoalesceList`: return the first non-empty list argument.
  * `Compact`: take a list of strings and return a new list of strings with all empty strings removed.
  * `Contains`: returns true if a given value appears as an element in a list, tuple, or set.
  * `Distinct`: filters duplicate elements from a list while retaining the order of remaining items.
  * `ChunkList`: turn a list into a list-of-lists where each top-level list is a "chunk" of a particular size of elements from the input.
  * `Flatten`: given a sequence that might contain other sequences, eliminate any intermediate sequences to produce a flat sequence.
  * `Keys`: return a list of keys from a map or object value in lexical order.
  * `Values`: return a list of values from a map in the same order as `Keys`.
  * `Lookup`: conditional lookup of an element from a map if it's present, or a fallback value if not. (This one differs from its Terraform equivalent in that the default value argument is _required_.)
  * `Merge`: given one or more maps or objects, merge them together into a single collection.
  * `ReverseList`: given a list, return a new list with the same items in the opposite order.
  * `SetProduct`: compute the cartesian product of one or more sets.
  * `Slice`: extract a consecutive sub-list from a list.
  * `Zipmap`: given a pair of lists of the same length, interpret the first as keys and the second as corresponding values to produce a map.
  * A factory `MakeToFunc` to build functions that each convert to a particular type constraint.
  * `TimeAdd`: add a duration to a timestamp to produce a new timestamp.
  * `Ceil` and `Floor`: round a fractional value to the nearest integer, away from or towards zero respectively.
  * `Log`: computes a logarithm in a given base.
  * `Pow`: implements exponentiation.
  * `ParseInt`: parses a string containing digits in a particular base to produce a whole number value.
  * `Join`: concatenates the elements of a list of strings with a given separator to produce a string.
  * `Split`: partitions a string by a given separator, returning a list of strings.
  * `Sort`: sorts a list of strings into lexical order.
  * `Chomp`: removes one or more newline characters from the end of a given string, producing a new string.
  * `Indent`: prepends a number of spaces to all lines except the first in a given string, producing a new string.
  * `Title`: converts a string to "title case".
  * `TrimSpace`: trims spaces from the start and end of a given string.
  * `Trim`: generalization of `TrimSpace` that allows user-specified trimming characters.
  * `TrimPrefix`: like `Trim` but only at the start of the string.
  * `TrimSuffix`: like `Trim` but only at the end of the string.

# 1.2.1 (January 10, 2020)

* `cty`: Fixed an infinite recursion bug when working with sets containing nested data structures. ([#35](https://github.com/zclconf/go-cty/pull/35))

# 1.2.0 (December 14, 2019)

* `cty`: Applications can now implement a general subset of the `cty` operations when creating a capsule type. For more information, see [Capsule Type Operation Definitions](./docs/capsule-type-operations.md).
* `cty`: Values now support a new mechanism called [Value Marks](./docs/marks.md) which can be used to transit additional metadata through expressions by marking the input values and then observing which marks propagated to the result value. This could be used, for example, to detect whether a value was derived from a particular other value in case that is useful for giving extra feedback in an error message.

# 1.1.1 (November 26, 2019)

* `cty`: Fixed a panic situation when trying to round-trip `cty.Number` values
  through `encoding/gob`. ([#32](https://github.com/zclconf/go-cty/pull/32))
* `convert`: Invalid string conversions to bool that use incorrect case will now give more actionable feedback. ([#29](https://github.com/zclconf/go-cty/pull/29))
* `function/stdlib`: The `formatlist` function will no longer panic if given
  an unknown tuple as one of its arguments.

# 1.1.0 (July 25, 2019)

* New method `Path.Equals` for robustly comparing `cty.Path` values. Previously
  callers might've used `reflect.DeepEqual` or similar, but that is not
  correct when a path contains a `cty.Number` index because `reflect.DeepEqual`
  does not correctly represent equality for number values.
  ([#25](https://github.com/zclconf/go-cty/pull/25))

# 1.0.0 (June 6, 2019)

Initial stable release.
