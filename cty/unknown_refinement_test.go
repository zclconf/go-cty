package cty

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestValueRefine(t *testing.T) {
	tests := map[string]struct {
		Build     func() Value
		Want      Value
		WantPanic any
	}{
		"DynamicVal silently ignores all refinements": {
			Build: func() Value {
				// This particular value, unlike any other value, will just
				// accept whatever refinements that are thrown at it and
				// completely ignore all of them and just continue being
				// itself.
				// This is a compromise for backward-compatiblity because
				// existing codebases expect cty.DynamicVal itself to be
				// the only value that is an unknown value of an unknown
				// type, aside from the possibility of marks.
				return DynamicVal.Refine().
					NotNull().
					StringPrefix("beep").
					NumberRangeInclusive(Zero, NumberIntVal(10)).
					CollectionLength(5).
					NewValue()
			},
			Want: DynamicVal,
		},
		"untyped null can be refined as being null": {
			Build: func() Value {
				return NullVal(DynamicPseudoType).Refine().
					Null().
					NewValue()
			},
			Want: NullVal(DynamicPseudoType),
		},
		"untyped null cannot be refined as being non-null": {
			Build: func() Value {
				return NullVal(DynamicPseudoType).RefineNotNull()
			},
			WantPanic: `refining null value as non-null`,
		},
		"unknown object can be refined non-null": {
			Build: func() Value {
				return UnknownVal(EmptyObject).RefineNotNull()
			},
			Want: UnknownVal(EmptyObject).RefineNotNull(),
		},
		"unknown tuple can be refined non-null": {
			Build: func() Value {
				return UnknownVal(EmptyTuple).RefineNotNull()
			},
			Want: UnknownVal(EmptyTuple).RefineNotNull(),
		},
		"unknown list can be refined non-null": {
			Build: func() Value {
				return UnknownVal(List(String)).RefineNotNull()
			},
			Want: UnknownVal(List(String)).RefineNotNull(),
		},
		"unknown map can be refined non-null": {
			Build: func() Value {
				return UnknownVal(Map(String)).RefineNotNull()
			},
			Want: UnknownVal(Map(String)).RefineNotNull(),
		},
		"unknown set can be refined non-null": {
			Build: func() Value {
				return UnknownVal(Set(String)).RefineNotNull()
			},
			Want: UnknownVal(Set(String)).RefineNotNull(),
		},
		"unknown string can be refined non-null": {
			Build: func() Value {
				return UnknownVal(String).RefineNotNull()
			},
			Want: UnknownVal(String).RefineNotNull(),
		},
		"unknown number can be refined non-null": {
			Build: func() Value {
				return UnknownVal(Number).RefineNotNull()
			},
			Want: UnknownVal(Number).RefineNotNull(),
		},
		"unknown bool can be refined non-null": {
			Build: func() Value {
				return UnknownVal(Bool).RefineNotNull()
			},
			Want: UnknownVal(Bool).RefineNotNull(),
		},
		"known null value can have its nullness confirmed": {
			Build: func() Value {
				return NullVal(Bool).Refine().
					Null().
					NewValue()
			},
			Want: NullVal(Bool),
		},
		"known null value cannot be refined as not null": {
			Build: func() Value {
				return NullVal(Bool).RefineNotNull()
			},
			WantPanic: `refining null value as non-null`,
		},

		// String refinements
		"unknown string can be refined with a prefix": {
			Build: func() Value {
				return UnknownVal(String).Refine().
					StringPrefix("foo-").
					NewValue()
			},
			Want: UnknownVal(String).Refine().
				StringPrefixFull("foo-").
				NewValue(),
		},
		"string prefix gets truncated if it might combine (latin diacritics)": {
			Build: func() Value {
				return UnknownVal(String).Refine().
					StringPrefix("foo").
					NewValue()
			},
			Want: UnknownVal(String).Refine().
				StringPrefixFull("fo").
				NewValue(),
		},
		"string prefix gets truncated if it might combine (emoji sequences)": {
			Build: func() Value {
				return UnknownVal(String).Refine().
					StringPrefix("a😶"). // Can combine with "clouds" to produce "face in clouds"
					NewValue()
			},
			Want: UnknownVal(String).Refine().
				StringPrefixFull("a").
				NewValue(),
		},
		"string prefix forced despite possibility of combining": {
			Build: func() Value {
				return UnknownVal(String).Refine().
					StringPrefixFull("foo").
					NewValue()
			},
			Want: UnknownVal(String).Refine().
				StringPrefixFull("foo").
				NewValue(),
		},
		"a string prefix can be extended": {
			Build: func() Value {
				return UnknownVal(String).Refine().
					StringPrefixFull("foo-").
					StringPrefixFull("foo-bar-").
					NewValue()
			},
			Want: UnknownVal(String).Refine().
				StringPrefixFull("foo-bar-").
				NewValue(),
		},
		"cannot provide a string prefix that conflicts with existing refinement": {
			Build: func() Value {
				return UnknownVal(String).Refine().
					StringPrefixFull("foo-").
					StringPrefixFull("bar-").
					NewValue()
			},
			WantPanic: `refined prefix is inconsistent with previous refined prefix`,
		},
		"a known string can have its prefix confirmed": {
			Build: func() Value {
				return StringVal("foo-baz").Refine().
					StringPrefixFull("foo-").
					NewValue()
			},
			Want: StringVal("foo-baz"),
		},
		"a known string does not accept a conflicting prefix": {
			Build: func() Value {
				return StringVal("foo-baz").Refine().
					StringPrefixFull("bar-").
					NewValue()
			},
			WantPanic: `refined prefix is inconsistent with known value`,
		},
		"non-string values cannot be refined with string prefix": {
			Build: func() Value {
				return UnknownVal(Number).Refine().
					StringPrefixFull("foo").
					NewValue()
			},
			WantPanic: `cannot refine string prefix for a cty.Number value`,
		},

		// Number refinements
		"unknown number can have refined lower bound": {
			Build: func() Value {
				return UnknownVal(Number).Refine().
					NumberRangeLowerBound(NumberIntVal(1), true).
					NewValue()
			},
			Want: UnknownVal(Number).Refine().
				NumberRangeLowerBound(NumberIntVal(1), true).
				NewValue(),
		},
		"unknown number can have refined upper bound": {
			Build: func() Value {
				return UnknownVal(Number).Refine().
					NumberRangeUpperBound(NumberIntVal(1), true).
					NewValue()
			},
			Want: UnknownVal(Number).Refine().
				NumberRangeUpperBound(NumberIntVal(1), true).
				NewValue(),
		},
		"unknown number can have refined both bounds": {
			Build: func() Value {
				return UnknownVal(Number).Refine().
					NumberRangeLowerBound(NumberIntVal(1), true).
					NumberRangeUpperBound(NumberIntVal(2), false).
					NewValue()
			},
			Want: UnknownVal(Number).Refine().
				NumberRangeLowerBound(NumberIntVal(1), true).
				NumberRangeUpperBound(NumberIntVal(2), false).
				NewValue(),
		},
		"refining unknown non-null number with equal upper and lower bound produces known number": {
			Build: func() Value {
				return UnknownVal(Number).Refine().
					NumberRangeLowerBound(NumberIntVal(1), true).
					NumberRangeUpperBound(NumberIntVal(1), true).
					NotNull().
					NewValue()
			},
			Want: NumberIntVal(1),
		},
		"unknown number cannot have conflicting bounds": {
			Build: func() Value {
				return UnknownVal(Number).Refine().
					NumberRangeLowerBound(NumberIntVal(2), true).
					NumberRangeUpperBound(NumberIntVal(1), false).
					NewValue()
			},
			WantPanic: `number lower bound cty.NumberIntVal(2) is greater than upper bound cty.NumberIntVal(1)`,
		},
		"known number can have its bounds confirmed": {
			Build: func() Value {
				return NumberIntVal(1).Refine().
					NumberRangeLowerBound(NumberIntVal(0), true).
					NumberRangeUpperBound(NumberIntVal(2), true).
					NotNull().
					NewValue()
			},
			Want: NumberIntVal(1),
		},
		"can't refine a known number with non-matching bounds": {
			Build: func() Value {
				return NumberIntVal(10).Refine().
					NumberRangeLowerBound(NumberIntVal(0), true).
					NumberRangeUpperBound(NumberIntVal(2), true).
					NotNull().
					NewValue()
			},
			WantPanic: `refining cty.NumberIntVal(10) to be <= cty.NumberIntVal(2)`,
		},

		// List length refinements
		"unknown list can be refined with length lower bound": {
			Build: func() Value {
				return UnknownVal(List(String)).Refine().
					CollectionLengthLowerBound(1).
					NewValue()
			},
			Want: UnknownVal(List(String)).Refine().
				CollectionLengthLowerBound(1).
				NewValue(),
		},
		"unknown list can be refined with length upper bound": {
			Build: func() Value {
				return UnknownVal(List(String)).Refine().
					CollectionLengthUpperBound(1).
					NewValue()
			},
			Want: UnknownVal(List(String)).Refine().
				CollectionLengthUpperBound(1).
				NewValue(),
		},
		"unknown list can be refined with length bounds": {
			Build: func() Value {
				return UnknownVal(List(String)).Refine().
					CollectionLengthLowerBound(1).
					CollectionLengthUpperBound(3).
					NewValue()
			},
			Want: UnknownVal(List(String)).Refine().
				CollectionLengthLowerBound(1).
				CollectionLengthUpperBound(3).
				NewValue(),
		},
		"unknown non-null list with known length becomes known list of unknowns": {
			Build: func() Value {
				return UnknownVal(List(String)).Refine().
					NotNull().
					CollectionLength(2).
					NewValue()
			},
			Want: ListVal([]Value{UnknownVal(String), UnknownVal(String)}),
		},
		"unknown non-null list with known zero length becomes known list": {
			Build: func() Value {
				return UnknownVal(List(String)).Refine().
					NotNull().
					CollectionLength(0).
					NewValue()
			},
			Want: ListValEmpty(String),
		},
		"known list can have its length confirmed with a refinement": {
			Build: func() Value {
				return ListValEmpty(String).Refine().
					CollectionLength(0).
					NewValue()
			},
			Want: ListValEmpty(String),
		},
		"cannot refine known list with conflicting length bounds": {
			Build: func() Value {
				return ListValEmpty(String).Refine().
					CollectionLength(1).
					NewValue()
			},
			WantPanic: `refining collection of length cty.NumberIntVal(0) with lower bound 1`,
		},

		// Map length refinements
		"unknown map can be refined with length lower bound": {
			Build: func() Value {
				return UnknownVal(Map(String)).Refine().
					CollectionLengthLowerBound(1).
					NewValue()
			},
			Want: UnknownVal(Map(String)).Refine().
				CollectionLengthLowerBound(1).
				NewValue(),
		},
		"unknown map can be refined with length upper bound": {
			Build: func() Value {
				return UnknownVal(Map(String)).Refine().
					CollectionLengthUpperBound(1).
					NewValue()
			},
			Want: UnknownVal(Map(String)).Refine().
				CollectionLengthUpperBound(1).
				NewValue(),
		},
		"unknown map can be refined with length bounds": {
			Build: func() Value {
				return UnknownVal(Map(String)).Refine().
					CollectionLengthLowerBound(1).
					CollectionLengthUpperBound(3).
					NewValue()
			},
			Want: UnknownVal(Map(String)).Refine().
				CollectionLengthLowerBound(1).
				CollectionLengthUpperBound(3).
				NewValue(),
		},
		"unknown map can be refined with known length": {
			Build: func() Value {
				return UnknownVal(Map(String)).Refine().
					NotNull().
					CollectionLength(2).
					NewValue()
			},
			Want: UnknownVal(Map(String)).Refine().
				NotNull().
				CollectionLength(2).
				NewValue(),
		},
		"unknown non-null map with known zero length becomes known map": {
			Build: func() Value {
				return UnknownVal(Map(String)).Refine().
					NotNull().
					CollectionLength(0).
					NewValue()
			},
			Want: MapValEmpty(String),
		},
		"known map can have its length confirmed with a refinement": {
			Build: func() Value {
				return MapValEmpty(String).Refine().
					CollectionLength(0).
					NewValue()
			},
			Want: MapValEmpty(String),
		},
		"cannot refine known map with conflicting length bounds": {
			Build: func() Value {
				return MapValEmpty(String).Refine().
					CollectionLength(1).
					NewValue()
			},
			WantPanic: `refining collection of length cty.NumberIntVal(0) with lower bound 1`,
		},

		// Set length refinements
		"unknown set can be refined with length lower bound": {
			Build: func() Value {
				return UnknownVal(Set(String)).Refine().
					CollectionLengthLowerBound(1).
					NewValue()
			},
			Want: UnknownVal(Set(String)).Refine().
				CollectionLengthLowerBound(1).
				NewValue(),
		},
		"unknown set can be refined with length upper bound": {
			Build: func() Value {
				return UnknownVal(Set(String)).Refine().
					CollectionLengthUpperBound(1).
					NewValue()
			},
			Want: UnknownVal(Set(String)).Refine().
				CollectionLengthUpperBound(1).
				NewValue(),
		},
		"unknown set can be refined with length bounds": {
			Build: func() Value {
				return UnknownVal(Set(String)).Refine().
					CollectionLengthLowerBound(1).
					CollectionLengthUpperBound(3).
					NewValue()
			},
			Want: UnknownVal(Set(String)).Refine().
				CollectionLengthLowerBound(1).
				CollectionLengthUpperBound(3).
				NewValue(),
		},
		"unknown set can be refined with known length": {
			Build: func() Value {
				return UnknownVal(Set(String)).Refine().
					NotNull().
					CollectionLength(2).
					NewValue()
			},
			Want: UnknownVal(Set(String)).Refine().
				NotNull().
				CollectionLength(2).
				NewValue(),
		},
		"unknown non-null set with known zero length becomes known empty set": {
			Build: func() Value {
				return UnknownVal(Set(String)).Refine().
					NotNull().
					CollectionLength(0).
					NewValue()
			},
			Want: SetValEmpty(String),
		},
		"known set can have its length confirmed with a refinement": {
			Build: func() Value {
				return SetValEmpty(String).Refine().
					CollectionLength(0).
					NewValue()
			},
			Want: SetValEmpty(String),
		},
		"cannot refine known set with conflicting length bounds": {
			Build: func() Value {
				return SetValEmpty(String).Refine().
					CollectionLength(1).
					NewValue()
			},
			WantPanic: `refining collection of length cty.NumberIntVal(0) with lower bound 1`,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			try := func(f func() Value) (ret Value, panicVal any) {
				defer func() {
					panicVal = recover()
				}()
				ret = f()
				return
			}
			got, panicVal := try(test.Build)

			if test.WantPanic == nil {
				if panicVal != nil {
					t.Fatalf("unexpected panic: %s", panicVal)
				}
				if !test.Want.RawEquals(got) {
					t.Errorf("wrong result\ngot:  %#v\nwant: %#v", got, test.Want)
				}
			} else {
				if panicVal == nil {
					t.Fatalf("unexpected success\nresult: %#v\nwant panic: %#v", got, test.WantPanic)
				}

				if diff := cmp.Diff(test.WantPanic, panicVal); diff != "" {
					t.Errorf("wrong panic value\n%s", diff)
				}
			}

		})
	}
}
