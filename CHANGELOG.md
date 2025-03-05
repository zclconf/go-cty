# 1.4.1 (Unreleased)


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
