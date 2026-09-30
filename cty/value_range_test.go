package cty

import (
	"fmt"
	"testing"
)

func TestValueRangeAsValue(t *testing.T) {
	tests := []struct {
		Input ValueRange
		Want  Value
	}{
		{
			ValueRange{
				ty: String,
				raw: &refinementNullable{
					isNull: tristateTrue,
				},
			},
			NullVal(String),
		},
		{
			ValueRange{
				ty: EmptyObject,
				raw: &refinementNullable{
					isNull: tristateFalse,
				},
			},
			UnknownVal(EmptyObject).RefineNotNull(),
		},
		{
			ValueRange{
				ty: String,
				raw: &refinementString{
					prefix: "foo:",
					refinementNullable: refinementNullable{
						isNull: tristateFalse,
					},
				},
			},
			UnknownVal(String).Refine().
				NotNull().
				StringPrefixFull("foo:").
				NewValue(),
		},
		{
			ValueRange{
				ty: List(String),
				raw: &refinementCollection{
					minLen: 2,
					maxLen: 2,
					refinementNullable: refinementNullable{
						isNull: tristateFalse,
					},
				},
			},
			ListVal([]Value{UnknownVal(String), UnknownVal(String)}),
		},
		{
			ValueRange{
				ty: Set(String),
				raw: &refinementCollection{
					minLen: 1,
					maxLen: 1,
					refinementNullable: refinementNullable{
						isNull: tristateFalse,
					},
				},
			},
			SetVal([]Value{UnknownVal(String)}),
		},
		{
			ValueRange{
				ty: String,
				raw: &refinementString{
					prefix: "boop:",
				},
			},
			UnknownVal(String).Refine().
				StringPrefixFull("boop:").
				NewValue(),
		},
		{
			DynamicVal.Range(),
			DynamicVal,
		},
		{
			NullVal(Number).Range(),
			NullVal(Number),
		},
		{
			StringVal("hello").Range(),
			UnknownVal(String).Refine().
				NotNull().
				StringPrefixFull("hello").
				NewValue(),
		},
		{
			True.Range(),
			UnknownVal(Bool).RefineNotNull(),
		},
		{
			NumberIntVal(32).Range(),
			NumberIntVal(32),
		},
		{
			ListValEmpty(String).Range(),
			ListValEmpty(String),
		},
		{
			SetValEmpty(String).Range(),
			SetValEmpty(String),
		},
		{
			MapValEmpty(String).Range(),
			MapValEmpty(String),
		},
		{
			ListVal([]Value{True}).Range(),
			ListVal([]Value{UnknownVal(Bool)}),
		},
		{
			ListVal([]Value{True, False}).Range(),
			ListVal([]Value{UnknownVal(Bool), UnknownVal(Bool)}),
		},
		{
			SetVal([]Value{True}).Range(),
			SetVal([]Value{UnknownVal(Bool)}),
		},
		{
			SetVal([]Value{True, False}).Range(),
			UnknownVal(Set(Bool)).Refine().
				NotNull().
				CollectionLength(2).
				NewValue(),
		},
		{
			SetVal([]Value{True}).Range(),
			SetVal([]Value{UnknownVal(Bool)}),
		},
		{
			MapVal(map[string]Value{"a": True}).Range(),
			UnknownVal(Map(Bool)).Refine().
				NotNull().
				CollectionLength(1).
				NewValue(),
		},
		{
			MapVal(map[string]Value{"a": True, "b": False}).Range(),
			UnknownVal(Map(Bool)).Refine().
				NotNull().
				CollectionLength(2).
				NewValue(),
		},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("%#v", test.Input), func(t *testing.T) {
			got := test.Input.AsValue()
			if !got.RawEquals(test.Want) {
				t.Errorf("wrong result\ninput: %#v\ngot:   %#v\nwant:  %#v", test.Input, got, test.Want)

				// Some of the input values here rely on manually-constructed
				// unknown values with refinements, and so it's possible to
				// accidentally construct a shape that RefinementBuilder could
				// never actually produce and so would fail RawEquals above
				// despite being semantically equivalent. Therefore we also show
				// the internals here to help identify those problems, which
				// are typically bugs in this test rather than bugs in the
				// code being tested (but not necessarily!).
				t.Logf("innards of values:\ngot v:   %#v\ngot ty:  %#v\nwant v:  %#v\nwant ty: %#v", got.v, got.ty, test.Want.v, test.Want.ty)
			}
		})
	}
}
