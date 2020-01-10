# 1.2.2 (Unreleased)


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
