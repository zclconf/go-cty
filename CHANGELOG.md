# 1.1.1 (Unreleased)

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
