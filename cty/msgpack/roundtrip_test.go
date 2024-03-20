package msgpack

import (
	"fmt"
	"strings"
	"testing"

	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/convert"
)

func TestRoundTrip(t *testing.T) {
	bigNumberVal, err := cty.ParseNumberVal("9999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999")
	if err != nil {
		t.Fatal(err)
	}
	awkwardFractionVal, err := cty.ParseNumberVal("0.8") // awkward because it can't be represented exactly in binary
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		Value cty.Value
		Type  cty.Type
	}{

		{
			cty.StringVal("hello"),
			cty.String,
		},
		{
			cty.StringVal(""),
			cty.String,
		},
		{
			cty.NullVal(cty.String),
			cty.String,
		},
		{
			cty.UnknownVal(cty.String),
			cty.String,
		},
		{
			cty.UnknownVal(cty.String).RefineNotNull(),
			cty.String,
		},
		{
			cty.UnknownVal(cty.String).Refine().StringPrefix("foo-").NewValue(),
			cty.String,
		},
		{
			cty.UnknownVal(cty.String).Refine().NotNull().StringPrefix("foo-").NewValue(),
			cty.String,
		},
		{
			cty.True,
			cty.Bool,
		},
		{
			cty.False,
			cty.Bool,
		},
		{
			cty.NullVal(cty.Bool),
			cty.Bool,
		},
		{
			cty.UnknownVal(cty.Bool),
			cty.Bool,
		},
		{
			cty.UnknownVal(cty.Bool).RefineNotNull(),
			cty.Bool,
		},

		{
			cty.NumberIntVal(1),
			cty.Number,
		},
		{
			cty.NumberFloatVal(1.5),
			cty.Number,
		},
		{
			bigNumberVal,
			cty.Number,
		},
		{
			cty.MustParseNumberVal("9223372036854775807"),
			cty.Number,
		},
		{
			cty.MustParseNumberVal("9223372036854775808"),
			cty.Number,
		},
		{
			cty.MustParseNumberVal("9223372036854775809"),
			cty.Number,
		},
		{
			cty.MustParseNumberVal("18446744073709551616"),
			cty.Number,
		},
		{
			awkwardFractionVal,
			cty.Number,
		},
		{
			cty.PositiveInfinity,
			cty.Number,
		},
		{
			cty.NegativeInfinity,
			cty.Number,
		},
		{
			cty.UnknownVal(cty.Number),
			cty.Number,
		},
		{
			cty.UnknownVal(cty.Number).RefineNotNull(),
			cty.Number,
		},
		{
			cty.UnknownVal(cty.Number).Refine().NumberRangeLowerBound(cty.Zero, true).NewValue(),
			cty.Number,
		},
		{
			cty.UnknownVal(cty.Number).Refine().NumberRangeLowerBound(cty.Zero, false).NewValue(),
			cty.Number,
		},
		{
			cty.UnknownVal(cty.Number).Refine().NumberRangeUpperBound(cty.Zero, true).NewValue(),
			cty.Number,
		},
		{
			cty.UnknownVal(cty.Number).Refine().NumberRangeUpperBound(cty.Zero, false).NewValue(),
			cty.Number,
		},
		{
			cty.UnknownVal(cty.Number).Refine().NumberRangeInclusive(cty.Zero, cty.NumberIntVal(1)).NewValue(),
			cty.Number,
		},

		{
			cty.ListVal([]cty.Value{
				cty.StringVal("hello"),
			}),
			cty.List(cty.String),
		},
		{
			cty.ListVal([]cty.Value{
				cty.UnknownVal(cty.String),
			}),
			cty.List(cty.String),
		},
		{
			cty.ListVal([]cty.Value{
				cty.NullVal(cty.String),
			}),
			cty.List(cty.String),
		},
		{
			cty.NullVal(cty.List(cty.String)),
			cty.List(cty.String),
		},
		{
			cty.ListValEmpty(cty.String),
			cty.List(cty.String),
		},
		{
			cty.UnknownVal(cty.List(cty.String)),
			cty.List(cty.String),
		},
		{
			cty.UnknownVal(cty.List(cty.String)).RefineNotNull(),
			cty.List(cty.String),
		},
		{
			cty.UnknownVal(cty.List(cty.String)).Refine().CollectionLengthLowerBound(1).NewValue(),
			cty.List(cty.String),
		},
		{
			cty.UnknownVal(cty.List(cty.String)).Refine().CollectionLengthUpperBound(1).NewValue(),
			cty.List(cty.String),
		},
		{
			cty.UnknownVal(cty.List(cty.String)).Refine().CollectionLengthLowerBound(1).CollectionLengthUpperBound(2).NewValue(),
			cty.List(cty.String),
		},
		{
			// NOTE: This refinement should collapse to a known 2-element list with unknown elements
			cty.UnknownVal(cty.List(cty.String)).Refine().CollectionLengthLowerBound(2).CollectionLengthUpperBound(2).NewValue(),
			cty.List(cty.String),
		},
		{
			cty.UnknownVal(cty.List(cty.String)).Refine().CollectionLengthUpperBound(1).NotNull().NewValue(),
			cty.List(cty.String),
		},

		{
			cty.SetVal([]cty.Value{
				cty.StringVal("hello"),
			}),
			cty.Set(cty.String),
		},
		{
			cty.SetVal([]cty.Value{
				cty.UnknownVal(cty.String),
			}),
			cty.Set(cty.String),
		},
		{
			cty.SetVal([]cty.Value{
				cty.NullVal(cty.String),
			}),
			cty.Set(cty.String),
		},
		{
			cty.SetValEmpty(cty.String),
			cty.Set(cty.String),
		},

		{
			cty.MapVal(map[string]cty.Value{
				"greeting": cty.StringVal("hello"),
			}),
			cty.Map(cty.String),
		},
		{
			cty.MapVal(map[string]cty.Value{
				"greeting": cty.UnknownVal(cty.String),
			}),
			cty.Map(cty.String),
		},
		{
			cty.MapVal(map[string]cty.Value{
				"greeting": cty.NullVal(cty.String),
			}),
			cty.Map(cty.String),
		},
		{
			cty.MapValEmpty(cty.String),
			cty.Map(cty.String),
		},

		{
			cty.TupleVal([]cty.Value{
				cty.StringVal("hello"),
			}),
			cty.Tuple([]cty.Type{cty.String}),
		},
		{
			cty.TupleVal([]cty.Value{
				cty.UnknownVal(cty.String),
			}),
			cty.Tuple([]cty.Type{cty.String}),
		},
		{
			cty.TupleVal([]cty.Value{
				cty.NullVal(cty.String),
			}),
			cty.Tuple([]cty.Type{cty.String}),
		},
		{
			cty.EmptyTupleVal,
			cty.EmptyTuple,
		},

		{
			cty.ObjectVal(map[string]cty.Value{
				"greeting": cty.StringVal("hello"),
			}),
			cty.Object(map[string]cty.Type{
				"greeting": cty.String,
			}),
		},
		{
			cty.ObjectVal(map[string]cty.Value{
				"greeting": cty.UnknownVal(cty.String),
			}),
			cty.Object(map[string]cty.Type{
				"greeting": cty.String,
			}),
		},
		{
			cty.ObjectVal(map[string]cty.Value{
				"greeting": cty.NullVal(cty.String),
			}),
			cty.Object(map[string]cty.Type{
				"greeting": cty.String,
			}),
		},
		{
			cty.ObjectVal(map[string]cty.Value{
				"a": cty.NullVal(cty.String),
				"b": cty.NullVal(cty.String),
			}),
			cty.Object(map[string]cty.Type{
				"a": cty.String,
				"b": cty.String,
			}),
		},
		{
			cty.ObjectVal(map[string]cty.Value{
				"a": cty.UnknownVal(cty.String),
				"b": cty.UnknownVal(cty.String),
			}),
			cty.Object(map[string]cty.Type{
				"a": cty.String,
				"b": cty.String,
			}),
		},
		{
			cty.EmptyObjectVal,
			cty.EmptyObject,
		},

		{
			cty.NullVal(cty.String),
			cty.DynamicPseudoType,
		},
		{
			cty.DynamicVal,
			cty.DynamicPseudoType,
		},
		{
			cty.ListVal([]cty.Value{
				cty.StringVal("hello"),
			}),
			cty.List(cty.DynamicPseudoType),
		},
		{
			cty.ListVal([]cty.Value{
				cty.NullVal(cty.String),
			}),
			cty.List(cty.DynamicPseudoType),
		},
		{
			cty.ListVal([]cty.Value{
				cty.DynamicVal,
			}),
			cty.List(cty.DynamicPseudoType),
		},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("%#v as %#v", test.Value, test.Type), func(t *testing.T) {
			b, err := Marshal(test.Value, test.Type)
			if err != nil {
				t.Fatal(err)
			}

			t.Logf("encoded as %x", b)

			got, err := Unmarshal(b, test.Type)
			if err != nil {
				t.Fatal(err)
			}

			if !got.RawEquals(test.Value) {
				t.Errorf(
					"value did not round-trip\ninput:  %#v\nresult: %#v",
					test.Value, got,
				)
			}
		})
	}
}

func TestRoundTrip_fromString(t *testing.T) {
	tests := []struct {
		Value string
		Type  cty.Type
	}{
		{
			"0",
			cty.Number,
		},
		{
			"1",
			cty.Number,
		},
		{
			"-1",
			cty.Number,
		},
		{
			"9223372036854775807",
			cty.Number,
		},
		{
			"9223372036854775808",
			cty.Number,
		},
		{
			"9223372036854775809",
			cty.Number,
		},
		{
			"18446744073709551616",
			cty.Number,
		},
		{
			"-9223372036854775807",
			cty.Number,
		},
		{
			"-9223372036854775808",
			cty.Number,
		},
		{
			"-9223372036854775809",
			cty.Number,
		},
		{
			"-18446744073709551616",
			cty.Number,
		},
		{
			"true",
			cty.Bool,
		},
		{
			"false",
			cty.Bool,
		},
	}
	for _, test := range tests {
		t.Run(fmt.Sprintf("%#v as %#v", test.Value, test.Type), func(t *testing.T) {
			stringVal := cty.StringVal(test.Value)

			original, err := convert.Convert(stringVal, test.Type)
			if err != nil {
				t.Fatalf("input type must be convertible from string: %s", err)
			}

			{
				// We'll first make sure that the conversion works even without
				// MessagePack involved, since otherwise we might falsely blame
				// the MessagePack encoding for bugs in package convert.
				stringGot, err := convert.Convert(original, cty.String)
				if err != nil {
					t.Fatalf("result must be convertible to string: %s", err)
				}

				if !stringGot.RawEquals(stringVal) {
					t.Fatalf("value did not round-trip to string even without msgpack\ninput:  %#v\nresult: %#v", test.Value, stringGot)
				}
			}

			b, err := Marshal(original, test.Type)
			if err != nil {
				t.Fatal(err)
			}

			t.Logf("encoded as %x", b)

			got, err := Unmarshal(b, test.Type)
			if err != nil {
				t.Fatal(err)
			}

			if !got.RawEquals(original) {
				t.Errorf(
					"value did not round-trip\ninput:  %#v\nresult: %#v",
					test.Value, got,
				)
			}

			stringGot, err := convert.Convert(got, cty.String)
			if err != nil {
				t.Fatalf("result must be convertible to string: %s", err)
			}

			if !stringGot.RawEquals(stringVal) {
				t.Errorf("value did not round-trip to string\ninput:  %#v\nresult: %#v", test.Value, stringGot)
			}

		})
	}
}

// Unknown values with very long string prefix refinements do not round-trip
// losslessly. If the prefix is longer than 256 bytes it will be truncated to
// a maximum of 256 bytes.
func TestRoundTrip_truncatesStringPrefixRefinement(t *testing.T) {
	tests := []struct {
		Value          cty.Value
		Type           cty.Type
		RoundTripValue cty.Value
	}{
		{
			cty.UnknownVal(cty.String).Refine().StringPrefix(strings.Repeat("a", 1024)).NewValue(),
			cty.String,
			cty.UnknownVal(cty.String).Refine().StringPrefix(strings.Repeat("a", 255)).NewValue(),
		},
		{
			cty.UnknownVal(cty.String).Refine().NotNull().StringPrefix(strings.Repeat("b", 1024)).NewValue(),
			cty.String,
			cty.UnknownVal(cty.String).Refine().NotNull().StringPrefix(strings.Repeat("b", 255)).NewValue(),
		},
		{
			cty.UnknownVal(cty.String).Refine().StringPrefix(strings.Repeat("c", 255) + "-").NewValue(),
			cty.String,
			cty.UnknownVal(cty.String).Refine().StringPrefix(strings.Repeat("c", 255) + "-").NewValue(),
		},
		{
			cty.UnknownVal(cty.String).Refine().StringPrefix(strings.Repeat("d", 255) + "🤷🤷").NewValue(),

			cty.String,
			cty.UnknownVal(cty.String).Refine().StringPrefix(strings.Repeat("d", 255)).NewValue(),
		},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("%#v as %#v", test.Value, test.Type), func(t *testing.T) {
			b, err := Marshal(test.Value, test.Type)
			if err != nil {
				t.Fatal(err)
			}

			t.Logf("encoded as %x", b)

			got, err := Unmarshal(b, test.Type)
			if err != nil {
				t.Fatal(err)
			}

			if !got.RawEquals(test.RoundTripValue) {
				t.Errorf(
					"unexpected value after round-trip\ninput:  %#v\nexpect: %#v\nresult: %#v",
					test.Value, test.RoundTripValue, got)
			}
		})
	}
}
