package cty

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestValueEquals(t *testing.T) {
	capsuleA := CapsuleVal(capsuleTestType1, &capsuleTestType1Native{"capsuleA"})
	capsuleB := CapsuleVal(capsuleTestType1, &capsuleTestType1Native{"capsuleB"})
	capsuleC := CapsuleVal(capsuleTestType2, &capsuleTestType2Native{"capsuleC"})

	unknownResult := UnknownVal(Bool).RefineNotNull()

	tests := []struct {
		LHS      Value
		RHS      Value
		Expected Value
	}{
		// Booleans
		{
			BoolVal(true),
			BoolVal(true),
			BoolVal(true),
		},
		{
			BoolVal(false),
			BoolVal(false),
			BoolVal(true),
		},
		{
			BoolVal(true),
			BoolVal(false),
			BoolVal(false),
		},

		// Numbers
		{
			NumberIntVal(1),
			NumberIntVal(2),
			BoolVal(false),
		},
		{
			NumberIntVal(2),
			NumberIntVal(2),
			BoolVal(true),
		},
		{
			NumberIntVal(2),
			NumberFloatVal(2.2),
			BoolVal(false),
		},
		{
			NumberFloatVal(2.0),
			NumberFloatVal(2.2),
			BoolVal(false),
		},
		{
			MustParseNumberVal("0.0"),
			MustParseNumberVal("-0.0"), // a statically-generated negative zero
			BoolVal(true),
		},
		{
			NumberFloatVal(0.0),
			NumberFloatVal(0.0).Multiply(NumberIntVal(-1)), // a dynamically-generated negative zero
			BoolVal(true),
		},
		{
			MustParseNumberVal("3.14159265358979323846264338327950288419716939937510582097494459"),
			MustParseNumberVal("3.14159265358979323846264338327950288419716939937510582097494459"),
			BoolVal(true),
		},
		{
			MustParseNumberVal("-3.14159265358979323846264338327950288419716939937510582097494459"),
			MustParseNumberVal("-3.14159265358979323846264338327950288419716939937510582097494459"),
			BoolVal(true),
		},
		{
			MustParseNumberVal("3.14159265358979323846264338327950288419716939937510582097494459"),
			MustParseNumberVal("-3.14159265358979323846264338327950288419716939937510582097494459"),
			BoolVal(false),
		},
		{
			MustParseNumberVal("1.2"),
			NumberFloatVal(1.2),
			BoolVal(true),
		},
		{
			MustParseNumberVal("1.22222"),
			NumberFloatVal(1.22222),
			BoolVal(true),
		},
		{
			MustParseNumberVal("9223372036854775808"),
			NumberFloatVal(float64(9223372036854775808)),
			BoolVal(true),
		},

		// Strings
		{
			StringVal(""),
			StringVal(""),
			BoolVal(true),
		},
		{
			StringVal("hello"),
			StringVal("hello"),
			BoolVal(true),
		},
		{
			StringVal("hello"),
			StringVal("world"),
			BoolVal(false),
		},
		{
			StringVal("0"),
			StringVal(""),
			BoolVal(false),
		},
		{
			StringVal("años"),
			StringVal("años"),
			BoolVal(true),
		},
		{
			// Combining marks are normalized by StringVal
			StringVal("años"),  // (precomposed tilde-n)
			StringVal("años"), // (combining tilde followed by bare n)
			BoolVal(true),
		},
		{
			// tilde-n does not normalize with bare n
			StringVal("años"),
			StringVal("anos"),
			BoolVal(false),
		},

		// Objects
		{
			ObjectVal(map[string]Value{}),
			ObjectVal(map[string]Value{}),
			BoolVal(true),
		},
		{
			ObjectVal(map[string]Value{
				"num": NumberIntVal(1),
			}),
			ObjectVal(map[string]Value{
				"num": NumberIntVal(1),
			}),
			BoolVal(true),
		},
		{
			ObjectVal(map[string]Value{
				"h\u00e9llo": NumberIntVal(1), // precombined é
			}),
			ObjectVal(map[string]Value{
				"he\u0301llo": NumberIntVal(1), // e with combining acute accent
			}),
			BoolVal(true),
		},
		{
			ObjectVal(map[string]Value{
				"num": NumberIntVal(1),
			}),
			ObjectVal(map[string]Value{}),
			BoolVal(false),
		},
		{
			ObjectVal(map[string]Value{
				"num":  NumberIntVal(1),
				"flag": BoolVal(true),
			}),
			ObjectVal(map[string]Value{
				"num":  NumberIntVal(1),
				"flag": BoolVal(true),
			}),
			BoolVal(true),
		},
		{
			ObjectVal(map[string]Value{
				"num": NumberIntVal(1),
			}),
			ObjectVal(map[string]Value{
				"num": NumberIntVal(2),
			}),
			BoolVal(false),
		},
		{
			ObjectVal(map[string]Value{
				"num": NumberIntVal(1),
			}),
			ObjectVal(map[string]Value{
				"othernum": NumberIntVal(1),
			}),
			BoolVal(false),
		},
		{
			ObjectVal(map[string]Value{
				"num":  NumberIntVal(1),
				"flag": BoolVal(true),
			}),
			ObjectVal(map[string]Value{
				"num": NumberIntVal(1),
			}),
			BoolVal(false),
		},
		{
			ObjectVal(map[string]Value{
				"num":  NumberIntVal(1),
				"flag": BoolVal(true),
			}),
			ObjectVal(map[string]Value{
				"num":  NumberIntVal(1),
				"flag": BoolVal(false),
			}),
			BoolVal(false),
		},

		// Tuples
		{
			EmptyTupleVal,
			EmptyTupleVal,
			BoolVal(true),
		},
		{
			TupleVal([]Value{NumberIntVal(1)}),
			TupleVal([]Value{NumberIntVal(1)}),
			BoolVal(true),
		},
		{
			TupleVal([]Value{NumberIntVal(1)}),
			TupleVal([]Value{NumberIntVal(2)}),
			BoolVal(false),
		},
		{
			TupleVal([]Value{StringVal("hi")}),
			TupleVal([]Value{NumberIntVal(1)}),
			BoolVal(false),
		},
		{
			TupleVal([]Value{NumberIntVal(1)}),
			TupleVal([]Value{NumberIntVal(1), NumberIntVal(2)}),
			BoolVal(false),
		},
		{
			TupleVal([]Value{NumberIntVal(1), NumberIntVal(2)}),
			TupleVal([]Value{NumberIntVal(1)}),
			BoolVal(false),
		},
		{
			TupleVal([]Value{NumberIntVal(1), NumberIntVal(2)}),
			TupleVal([]Value{NumberIntVal(1), NumberIntVal(2)}),
			BoolVal(true),
		},
		{
			TupleVal([]Value{UnknownVal(Number)}),
			TupleVal([]Value{NumberIntVal(1)}),
			unknownResult,
		},
		{
			TupleVal([]Value{UnknownVal(Number)}),
			TupleVal([]Value{UnknownVal(Number)}),
			unknownResult,
		},
		{
			TupleVal([]Value{NumberIntVal(1)}),
			TupleVal([]Value{UnknownVal(Number)}),
			unknownResult,
		},
		{
			TupleVal([]Value{NumberIntVal(1)}),
			TupleVal([]Value{DynamicVal}),
			unknownResult,
		},
		{
			TupleVal([]Value{DynamicVal}),
			TupleVal([]Value{NumberIntVal(1)}),
			unknownResult,
		},
		{
			TupleVal([]Value{NumberIntVal(1)}),
			UnknownVal(Tuple([]Type{Number})),
			unknownResult,
		},
		{
			UnknownVal(Tuple([]Type{Number})),
			TupleVal([]Value{NumberIntVal(1)}),
			unknownResult,
		},
		{
			DynamicVal,
			TupleVal([]Value{NumberIntVal(1)}),
			unknownResult,
		},
		{
			TupleVal([]Value{NumberIntVal(1)}),
			DynamicVal,
			unknownResult,
		},

		// Lists
		{
			ListValEmpty(Number),
			ListValEmpty(Number),
			BoolVal(true),
		},
		{
			ListValEmpty(Number),
			ListValEmpty(Bool),
			BoolVal(false),
		},
		{
			ListVal([]Value{
				NumberIntVal(1),
			}),
			ListVal([]Value{
				NumberIntVal(1),
			}),
			BoolVal(true),
		},
		{
			ListVal([]Value{
				NumberIntVal(1),
			}),
			ListValEmpty(String),
			BoolVal(false),
		},
		{
			ListVal([]Value{
				NumberIntVal(1),
				NumberIntVal(2),
			}),
			ListVal([]Value{
				NumberIntVal(1),
				NumberIntVal(2),
			}),
			BoolVal(true),
		},
		{
			ListVal([]Value{
				NumberIntVal(1),
			}),
			ListVal([]Value{
				NumberIntVal(2),
			}),
			BoolVal(false),
		},
		{
			ListVal([]Value{
				NumberIntVal(1),
				NumberIntVal(2),
			}),
			ListVal([]Value{
				NumberIntVal(1),
			}),
			BoolVal(false),
		},
		{
			ListVal([]Value{
				NumberIntVal(1),
			}),
			ListVal([]Value{
				NumberIntVal(1),
				NumberIntVal(2),
			}),
			BoolVal(false),
		},

		// Maps
		{
			MapValEmpty(Number),
			MapValEmpty(Number),
			BoolVal(true),
		},
		{
			MapValEmpty(Number),
			MapValEmpty(Bool),
			BoolVal(false),
		},
		{
			MapVal(map[string]Value{
				"num": NumberIntVal(1),
			}),
			MapVal(map[string]Value{
				"num": NumberIntVal(1),
			}),
			BoolVal(true),
		},
		{
			MapVal(map[string]Value{
				"h\u00e9llo": NumberIntVal(1), // precombined é
			}),
			MapVal(map[string]Value{
				"he\u0301llo": NumberIntVal(1), // e with combining acute accent
			}),
			BoolVal(true),
		},
		{
			MapVal(map[string]Value{
				"num": NumberIntVal(1),
			}),
			MapValEmpty(String),
			BoolVal(false),
		},
		{
			MapVal(map[string]Value{
				"num1": NumberIntVal(1),
				"num2": NumberIntVal(2),
			}),
			MapVal(map[string]Value{
				"num1": NumberIntVal(1),
				"num2": NumberIntVal(2),
			}),
			BoolVal(true),
		},
		{
			MapVal(map[string]Value{
				"num": NumberIntVal(1),
			}),
			MapVal(map[string]Value{
				"num": NumberIntVal(2),
			}),
			BoolVal(false),
		},
		{
			MapVal(map[string]Value{
				"num": NumberIntVal(1),
			}),
			MapVal(map[string]Value{
				"othernum": NumberIntVal(1),
			}),
			BoolVal(false),
		},
		{
			MapVal(map[string]Value{
				"num1": NumberIntVal(1),
				"num2": NumberIntVal(2),
			}),
			MapVal(map[string]Value{
				"num1": NumberIntVal(1),
			}),
			BoolVal(false),
		},
		{
			MapVal(map[string]Value{
				"num1": NumberIntVal(1),
			}),
			MapVal(map[string]Value{
				"num1": NumberIntVal(1),
				"num2": NumberIntVal(2),
			}),
			BoolVal(false),
		},
		{
			MapVal(map[string]Value{
				"num1": NumberIntVal(1),
				"num2": NumberIntVal(2),
			}),
			MapVal(map[string]Value{
				"num1": NumberIntVal(1),
				"num2": NumberIntVal(3),
			}),
			BoolVal(false),
		},

		// Sets
		{
			SetValEmpty(Number),
			SetValEmpty(Number),
			BoolVal(true),
		},
		{
			SetValEmpty(Number),
			SetValEmpty(Bool),
			BoolVal(false),
		},
		{
			SetVal([]Value{
				NumberIntVal(1),
			}),
			SetVal([]Value{
				NumberIntVal(1),
			}),
			BoolVal(true),
		},
		{
			SetVal([]Value{
				NumberIntVal(1),
			}),
			SetValEmpty(String),
			BoolVal(false),
		},
		{
			SetVal([]Value{
				NumberIntVal(1),
				NumberIntVal(2),
			}),
			SetVal([]Value{
				NumberIntVal(2),
				NumberIntVal(1),
			}),
			BoolVal(true),
		},
		{
			SetVal([]Value{
				NumberIntVal(1),
			}),
			SetVal([]Value{
				NumberIntVal(2),
			}),
			BoolVal(false),
		},
		{
			SetVal([]Value{
				NumberIntVal(1),
				NumberIntVal(2),
			}),
			SetVal([]Value{
				NumberIntVal(1),
			}),
			BoolVal(false),
		},
		{
			SetVal([]Value{
				NumberIntVal(1),
			}),
			SetVal([]Value{
				NumberIntVal(1),
				NumberIntVal(2),
			}),
			BoolVal(false),
		},
		{
			SetVal([]Value{
				NumberIntVal(1),
			}),
			SetVal([]Value{
				UnknownVal(Number),
			}),
			unknownResult,
		},
		{
			SetVal([]Value{
				NumberIntVal(1),
			}),
			SetVal([]Value{
				NumberIntVal(1),
				UnknownVal(Number),
			}),
			unknownResult,
		},
		{
			SetVal([]Value{
				NumberIntVal(1),
				UnknownVal(Number),
			}),
			SetVal([]Value{
				NumberIntVal(1),
			}),
			unknownResult,
		},

		// Capsules
		{
			capsuleA,
			capsuleA,
			True,
		},
		{
			capsuleA,
			capsuleB,
			False,
		},
		{
			capsuleA,
			capsuleC,
			False,
		},
		{
			capsuleA,
			UnknownVal(capsuleTestType1), // same type
			unknownResult,
		},
		{
			capsuleA,
			UnknownVal(capsuleTestType2), // different type
			False,
		},

		// Unknowns and Dynamics
		{
			NumberIntVal(2),
			UnknownVal(Number),
			unknownResult,
		},
		{
			NumberIntVal(1),
			DynamicVal,
			unknownResult,
		},
		{
			NumberIntVal(2),
			UnknownVal(Number).Refine().NumberRangeLowerBound(Zero, true).NewValue(),
			unknownResult,
		},
		{
			NumberIntVal(2),
			UnknownVal(Number).Refine().NumberRangeLowerBound(NumberIntVal(4), true).NewValue(),
			False, // deduction from refinement
		},
		{
			DynamicVal,
			BoolVal(true),
			unknownResult,
		},
		{
			DynamicVal,
			DynamicVal,
			unknownResult,
		},
		{
			ListVal([]Value{
				StringVal("hi"),
				DynamicVal,
			}),
			ListVal([]Value{
				StringVal("hi"),
				DynamicVal,
			}),
			unknownResult,
		},
		{
			ListVal([]Value{
				StringVal("hi"),
				UnknownVal(String),
			}),
			ListVal([]Value{
				StringVal("hi"),
				UnknownVal(String),
			}),
			unknownResult,
		},
		{
			UnknownVal(List(String)).Refine().CollectionLengthLowerBound(1).NewValue(),
			ListValEmpty(String),
			False, // deduction from refinement
		},
		{
			MapVal(map[string]Value{
				"static":  StringVal("hi"),
				"dynamic": DynamicVal,
			}),
			MapVal(map[string]Value{
				"static":  StringVal("hi"),
				"dynamic": DynamicVal,
			}),
			unknownResult,
		},
		{
			MapVal(map[string]Value{
				"static":  StringVal("hi"),
				"dynamic": UnknownVal(String),
			}),
			MapVal(map[string]Value{
				"static":  StringVal("hi"),
				"dynamic": UnknownVal(String),
			}),
			unknownResult,
		},
		{
			NullVal(String),
			NullVal(DynamicPseudoType),
			True,
		},
		{
			NullVal(String),
			NullVal(String),
			True,
		},
		{
			UnknownVal(String),
			UnknownVal(Number),
			unknownResult,
		},
		{
			StringVal(""),
			NullVal(DynamicPseudoType),
			False,
		},
		{
			StringVal(""),
			NullVal(String),
			False,
		},
		{
			StringVal(""),
			UnknownVal(String),
			unknownResult,
		},
		{
			NullVal(DynamicPseudoType),
			NullVal(DynamicPseudoType),
			True,
		},
		{
			NullVal(String),
			UnknownVal(Number),
			unknownResult, // because second operand might eventually be null
		},
		{
			UnknownVal(String),
			NullVal(Number),
			unknownResult, // because first operand might eventually be null
		},
		{
			UnknownVal(String),
			UnknownVal(Number),
			unknownResult, // because both operands might eventually be null
		},
		{
			StringVal("hello"),
			UnknownVal(Number),
			False, // because no number value -- even null -- can be equal to a non-null string
		},
		{
			UnknownVal(String),
			NumberIntVal(1),
			False, // because no string value -- even null -- can be equal to a non-null number
		},
		{
			ObjectVal(map[string]Value{
				"a": StringVal("a"),
			}),
			// A null value is always known
			ObjectVal(map[string]Value{
				"a": NullVal(DynamicPseudoType),
			}),
			BoolVal(false),
		},
		{
			ObjectVal(map[string]Value{
				"a": NullVal(DynamicPseudoType),
			}),
			ObjectVal(map[string]Value{
				"a": NullVal(DynamicPseudoType),
			}),
			BoolVal(true),
		},
		{
			ObjectVal(map[string]Value{
				"a": StringVal("a"),
				"b": UnknownVal(Number),
			}),
			// While we have a dynamic type, the different object types should
			// still compare false
			ObjectVal(map[string]Value{
				"a": NullVal(DynamicPseudoType),
				"c": UnknownVal(Number),
			}),
			BoolVal(false),
		},
		{
			ObjectVal(map[string]Value{
				"a": StringVal("a"),
				"b": UnknownVal(Number),
			}),
			// While we have a dynamic type, the different object types should
			// still compare false
			ObjectVal(map[string]Value{
				"a": DynamicVal,
				"c": UnknownVal(Number),
			}),
			BoolVal(false),
		},
		{
			ObjectVal(map[string]Value{
				"a": NullVal(DynamicPseudoType),
			}),
			ObjectVal(map[string]Value{
				"a": DynamicVal,
			}),
			unknownResult,
		},
		{
			ObjectVal(map[string]Value{
				"a": NullVal(List(String)),
			}),
			// While the unknown val does contain dynamic types, the overall
			// container types can't conform.
			ObjectVal(map[string]Value{
				"a": UnknownVal(List(List(DynamicPseudoType))),
			}),
			BoolVal(false),
		},
		{
			ObjectVal(map[string]Value{
				"a": NullVal(List(List(String))),
			}),
			ObjectVal(map[string]Value{
				"a": UnknownVal(List(List(DynamicPseudoType))),
			}),
			unknownResult,
		},
		{
			NullVal(String),
			UnknownVal(String).Refine().NotNull().NewValue(),
			False,
		},
		{
			UnknownVal(String).Refine().NotNull().NewValue(),
			NullVal(String),
			False,
		},
		{
			UnknownVal(String).Refine().Null().NewValue(),
			NullVal(String),
			True, // NOTE: The refinement should collapse to NullVal(String)
		},
		{
			NullVal(String),
			UnknownVal(String).Refine().Null().NewValue(),
			True, // NOTE: The refinement should collapse to NullVal(String)
		},
		{
			UnknownVal(String).Refine().StringPrefix("foo-").NewValue(),
			StringVal("notfoo-bar"),
			False, // Deduction from refinement
		},
		{
			StringVal(""),
			UnknownVal(String).Refine().StringPrefix("foo-").NewValue(),
			False, // Deduction from refinement
		},
		{
			StringVal("").Mark("a"),
			UnknownVal(String).Mark("b").Refine().StringPrefix("foo-").NewValue(),
			False.Mark("a").Mark("b"), // Deduction from refinement
		},
		{
			UnknownVal(String).Refine().StringPrefix("foo-").NewValue(),
			StringVal("foo-bar"),
			unknownResult,
		},

		// Marks
		{
			StringVal("a").Mark(1),
			StringVal("b"),
			False.Mark(1),
		},
		{
			StringVal("a"),
			StringVal("b").Mark(2),
			False.Mark(2),
		},
		{
			StringVal("a").Mark(1),
			StringVal("b").Mark(2),
			False.WithMarks(NewValueMarks(1, 2)),
		},

		{
			MapVal(map[string]Value{
				"a": StringVal("a").Mark("boop"),
			}),
			MapVal(map[string]Value{
				"a": StringVal("a").Mark("blop"),
			}),
			True.WithMarks(NewValueMarks("boop", "blop")),
		},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("%#v.Equals(%#v)", test.LHS, test.RHS), func(t *testing.T) {
			got := test.LHS.Equals(test.RHS)
			if !got.RawEquals(test.Expected) {
				t.Fatalf("wrong Equals result\nLHS:  %#v\nRHS:  %#v\ngot:  %#v\nwant: %#v", test.LHS, test.RHS, got, test.Expected)
			}
		})
	}
}

func TestValueRawEquals(t *testing.T) {
	capsuleA := CapsuleVal(capsuleTestType1, &capsuleTestType1Native{"capsuleA"})
	capsuleB := CapsuleVal(capsuleTestType1, &capsuleTestType1Native{"capsuleB"})
	capsuleC := CapsuleVal(capsuleTestType2, &capsuleTestType2Native{"capsuleC"})

	tests := []struct {
		LHS      Value
		RHS      Value
		Expected bool
	}{
		// Booleans
		{
			BoolVal(true),
			BoolVal(true),
			true,
		},
		{
			BoolVal(false),
			BoolVal(false),
			true,
		},
		{
			BoolVal(true),
			BoolVal(false),
			false,
		},

		// Numbers
		{
			NumberIntVal(1),
			NumberIntVal(2),
			false,
		},
		{
			NumberIntVal(2),
			NumberIntVal(2),
			true,
		},

		// Strings
		{
			StringVal(""),
			StringVal(""),
			true,
		},
		{
			StringVal("hello"),
			StringVal("hello"),
			true,
		},
		{
			StringVal("hello"),
			StringVal("world"),
			false,
		},
		{
			StringVal("0"),
			StringVal(""),
			false,
		},
		{
			StringVal("años"),
			StringVal("años"),
			true,
		},
		{
			// Combining marks are normalized by StringVal
			StringVal("años"),  // (precomposed tilde-n)
			StringVal("años"), // (combining tilde followed by bare n)
			true,
		},
		{
			// tilde-n does not normalize with bare n
			StringVal("años"),
			StringVal("anos"),
			false,
		},

		// Objects
		{
			ObjectVal(map[string]Value{}),
			ObjectVal(map[string]Value{}),
			true,
		},
		{
			ObjectVal(map[string]Value{
				"num": NumberIntVal(1),
			}),
			ObjectVal(map[string]Value{
				"num": NumberIntVal(1),
			}),
			true,
		},
		{
			ObjectVal(map[string]Value{
				"h\u00e9llo": NumberIntVal(1), // precombined é
			}),
			ObjectVal(map[string]Value{
				"he\u0301llo": NumberIntVal(1), // e with combining acute accent
			}),
			true,
		},
		{
			ObjectVal(map[string]Value{
				"num": NumberIntVal(1),
			}),
			ObjectVal(map[string]Value{}),
			false,
		},
		{
			ObjectVal(map[string]Value{
				"num":  NumberIntVal(1),
				"flag": BoolVal(true),
			}),
			ObjectVal(map[string]Value{
				"num":  NumberIntVal(1),
				"flag": BoolVal(true),
			}),
			true,
		},
		{
			ObjectVal(map[string]Value{
				"num": NumberIntVal(1),
			}),
			ObjectVal(map[string]Value{
				"num": NumberIntVal(2),
			}),
			false,
		},
		{
			ObjectVal(map[string]Value{
				"num": NumberIntVal(1),
			}),
			ObjectVal(map[string]Value{
				"othernum": NumberIntVal(1),
			}),
			false,
		},
		{
			ObjectVal(map[string]Value{
				"num":  NumberIntVal(1),
				"flag": BoolVal(true),
			}),
			ObjectVal(map[string]Value{
				"num": NumberIntVal(1),
			}),
			false,
		},
		{
			ObjectVal(map[string]Value{
				"num":  NumberIntVal(1),
				"flag": BoolVal(true),
			}),
			ObjectVal(map[string]Value{
				"num":  NumberIntVal(1),
				"flag": BoolVal(false),
			}),
			false,
		},

		// Tuples
		{
			EmptyTupleVal,
			EmptyTupleVal,
			true,
		},
		{
			TupleVal([]Value{NumberIntVal(1)}),
			TupleVal([]Value{NumberIntVal(1)}),
			true,
		},
		{
			TupleVal([]Value{NumberIntVal(1)}),
			TupleVal([]Value{NumberIntVal(2)}),
			false,
		},
		{
			TupleVal([]Value{StringVal("hi")}),
			TupleVal([]Value{NumberIntVal(1)}),
			false,
		},
		{
			TupleVal([]Value{NumberIntVal(1)}),
			TupleVal([]Value{NumberIntVal(1), NumberIntVal(2)}),
			false,
		},
		{
			TupleVal([]Value{NumberIntVal(1), NumberIntVal(2)}),
			TupleVal([]Value{NumberIntVal(1)}),
			false,
		},
		{
			TupleVal([]Value{NumberIntVal(1), NumberIntVal(2)}),
			TupleVal([]Value{NumberIntVal(1), NumberIntVal(2)}),
			true,
		},
		{
			TupleVal([]Value{UnknownVal(Number)}),
			TupleVal([]Value{NumberIntVal(1)}),
			false,
		},
		{
			TupleVal([]Value{UnknownVal(Number)}),
			TupleVal([]Value{UnknownVal(Number)}),
			true,
		},
		{
			TupleVal([]Value{NumberIntVal(1)}),
			TupleVal([]Value{UnknownVal(Number)}),
			false,
		},
		{
			TupleVal([]Value{NumberIntVal(1)}),
			TupleVal([]Value{DynamicVal}),
			false,
		},
		{
			TupleVal([]Value{DynamicVal}),
			TupleVal([]Value{NumberIntVal(1)}),
			false,
		},
		{
			TupleVal([]Value{NumberIntVal(1)}),
			UnknownVal(Tuple([]Type{Number})),
			false,
		},
		{
			UnknownVal(Tuple([]Type{Number})),
			TupleVal([]Value{NumberIntVal(1)}),
			false,
		},
		{
			DynamicVal,
			TupleVal([]Value{NumberIntVal(1)}),
			false,
		},
		{
			TupleVal([]Value{NumberIntVal(1)}),
			DynamicVal,
			false,
		},

		// Lists
		{
			ListValEmpty(Number),
			ListValEmpty(Number),
			true,
		},
		{
			ListValEmpty(Number),
			ListValEmpty(Bool),
			false,
		},
		{
			ListVal([]Value{
				NumberIntVal(1),
			}),
			ListVal([]Value{
				NumberIntVal(1),
			}),
			true,
		},
		{
			ListVal([]Value{
				NumberIntVal(1),
			}),
			ListValEmpty(String),
			false,
		},
		{
			ListVal([]Value{
				NumberIntVal(1),
				NumberIntVal(2),
			}),
			ListVal([]Value{
				NumberIntVal(1),
				NumberIntVal(2),
			}),
			true,
		},
		{
			ListVal([]Value{
				NumberIntVal(1),
			}),
			ListVal([]Value{
				NumberIntVal(2),
			}),
			false,
		},
		{
			ListVal([]Value{
				NumberIntVal(1),
				NumberIntVal(2),
			}),
			ListVal([]Value{
				NumberIntVal(1),
			}),
			false,
		},
		{
			ListVal([]Value{
				NumberIntVal(1),
			}),
			ListVal([]Value{
				NumberIntVal(1),
				NumberIntVal(2),
			}),
			false,
		},

		// Maps
		{
			MapValEmpty(Number),
			MapValEmpty(Number),
			true,
		},
		{
			MapValEmpty(Number).Mark("a"),
			MapValEmpty(Number).Mark("a"),
			true,
		},
		{
			MapValEmpty(Number).Mark("a"),
			MapValEmpty(Number),
			false,
		},
		{
			MapValEmpty(Number),
			MapValEmpty(Number).Mark("a"),
			false,
		},
		{
			MapValEmpty(Number).Mark("a"),
			MapValEmpty(Number).Mark("a").Mark("b"),
			false,
		},
		{
			MapValEmpty(Number),
			MapValEmpty(Bool),
			false,
		},
		{
			MapVal(map[string]Value{
				"num": NumberIntVal(1),
			}),
			MapVal(map[string]Value{
				"num": NumberIntVal(1),
			}),
			true,
		},
		{
			MapVal(map[string]Value{
				"h\u00e9llo": NumberIntVal(1), // precombined é
			}),
			MapVal(map[string]Value{
				"he\u0301llo": NumberIntVal(1), // e with combining acute accent
			}),
			true,
		},
		{
			MapVal(map[string]Value{
				"num": NumberIntVal(1),
			}),
			MapValEmpty(String),
			false,
		},
		{
			MapVal(map[string]Value{
				"num1": NumberIntVal(1),
				"num2": NumberIntVal(2),
			}),
			MapVal(map[string]Value{
				"num1": NumberIntVal(1),
				"num2": NumberIntVal(2),
			}),
			true,
		},
		{
			MapVal(map[string]Value{
				"num": NumberIntVal(1),
			}),
			MapVal(map[string]Value{
				"num": NumberIntVal(2),
			}),
			false,
		},
		{
			MapVal(map[string]Value{
				"num": NumberIntVal(1),
			}),
			MapVal(map[string]Value{
				"othernum": NumberIntVal(1),
			}),
			false,
		},
		{
			MapVal(map[string]Value{
				"num1": NumberIntVal(1),
				"num2": NumberIntVal(2),
			}),
			MapVal(map[string]Value{
				"num1": NumberIntVal(1),
			}),
			false,
		},
		{
			MapVal(map[string]Value{
				"num1": NumberIntVal(1),
			}),
			MapVal(map[string]Value{
				"num1": NumberIntVal(1),
				"num2": NumberIntVal(2),
			}),
			false,
		},
		{
			MapVal(map[string]Value{
				"num1": NumberIntVal(1),
				"num2": NumberIntVal(2),
			}),
			MapVal(map[string]Value{
				"num1": NumberIntVal(1),
				"num2": NumberIntVal(3),
			}),
			false,
		},

		// Sets
		{
			SetValEmpty(Number),
			SetValEmpty(Number),
			true,
		},
		{
			SetValEmpty(Number),
			SetValEmpty(Bool),
			false,
		},
		{
			SetVal([]Value{
				NumberIntVal(1),
			}),
			SetVal([]Value{
				NumberIntVal(1),
			}),
			true,
		},
		{
			SetVal([]Value{
				NumberIntVal(1),
			}),
			SetValEmpty(String),
			false,
		},
		{
			SetVal([]Value{
				NumberIntVal(1),
				NumberIntVal(2),
			}),
			SetVal([]Value{
				NumberIntVal(2),
				NumberIntVal(1),
			}),
			true,
		},
		{
			SetVal([]Value{
				NumberIntVal(1),
			}),
			SetVal([]Value{
				NumberIntVal(2),
			}),
			false,
		},
		{
			SetVal([]Value{
				NumberIntVal(1),
				NumberIntVal(2),
			}),
			SetVal([]Value{
				NumberIntVal(1),
			}),
			false,
		},
		{
			SetVal([]Value{
				NumberIntVal(1),
			}),
			SetVal([]Value{
				NumberIntVal(1),
				NumberIntVal(2),
			}),
			false,
		},

		// Capsules
		{
			capsuleA,
			capsuleA,
			true,
		},
		{
			capsuleA,
			capsuleB,
			false,
		},
		{
			capsuleA,
			capsuleC,
			false,
		},
		{
			capsuleA,
			UnknownVal(capsuleTestType1), // same type
			false,
		},
		{
			capsuleA,
			UnknownVal(capsuleTestType2), // different type
			false,
		},

		// Unknowns and Dynamics
		{
			NumberIntVal(2),
			UnknownVal(Number),
			false,
		},
		{
			NumberIntVal(1),
			DynamicVal,
			false,
		},
		{
			DynamicVal,
			BoolVal(true),
			false,
		},
		{
			DynamicVal,
			DynamicVal,
			true, //?
		},
		{
			ListVal([]Value{
				StringVal("hi"),
				DynamicVal,
			}),
			ListVal([]Value{
				StringVal("hi"),
				DynamicVal,
			}),
			true,
		},
		{
			ListVal([]Value{
				StringVal("hi"),
				UnknownVal(String),
			}),
			ListVal([]Value{
				StringVal("hi"),
				UnknownVal(String),
			}),
			true,
		},
		{
			MapVal(map[string]Value{
				"static":  StringVal("hi"),
				"dynamic": DynamicVal,
			}),
			MapVal(map[string]Value{
				"static":  StringVal("hi"),
				"dynamic": DynamicVal,
			}),
			true,
		},
		{
			MapVal(map[string]Value{
				"static":  StringVal("hi"),
				"dynamic": UnknownVal(String),
			}),
			MapVal(map[string]Value{
				"static":  StringVal("hi"),
				"dynamic": UnknownVal(String),
			}),
			true,
		},

		{
			NullVal(String),
			NullVal(DynamicPseudoType),
			false,
		},
		{
			NullVal(String),
			NullVal(String),
			true,
		},
		{
			UnknownVal(String),
			UnknownVal(Number),
			false,
		},
		{
			StringVal(""),
			NullVal(DynamicPseudoType),
			false,
		},
		{
			StringVal(""),
			NullVal(String),
			false,
		},
		{
			StringVal(""),
			UnknownVal(String),
			false,
		},
		{
			NullVal(DynamicPseudoType),
			NullVal(DynamicPseudoType),
			true,
		},
		{
			NullVal(String),
			UnknownVal(Number),
			false, // because second operand might eventually be null
		},
		{
			UnknownVal(String),
			NullVal(Number),
			false, // because first operand might eventually be null
		},
		{
			UnknownVal(String),
			UnknownVal(Number),
			false, // because both operands might eventually be null
		},
		{
			StringVal("hello"),
			UnknownVal(Number),
			false, // because no number value -- even null -- can be equal to a non-null string
		},
		{
			UnknownVal(String),
			NumberIntVal(1),
			false, // because no string value -- even null -- can be equal to a non-null number
		},
		{
			ObjectVal(map[string]Value{
				"a": StringVal("a"),
			}),
			// A null value is always known
			ObjectVal(map[string]Value{
				"a": NullVal(DynamicPseudoType),
			}),
			false,
		},
		{
			ObjectVal(map[string]Value{
				"a": NullVal(DynamicPseudoType),
			}),
			ObjectVal(map[string]Value{
				"a": NullVal(DynamicPseudoType),
			}),
			true,
		},
		{
			ObjectVal(map[string]Value{
				"a": StringVal("a"),
				"b": UnknownVal(Number),
			}),
			// While we have a dynamic type, the different object types should
			// still compare false
			ObjectVal(map[string]Value{
				"a": NullVal(DynamicPseudoType),
				"c": UnknownVal(Number),
			}),
			false,
		},
		{
			ObjectVal(map[string]Value{
				"a": StringVal("a"),
				"b": UnknownVal(Number),
			}),
			// While we have a dynamic type, the different object types should
			// still compare false
			ObjectVal(map[string]Value{
				"a": DynamicVal,
				"c": UnknownVal(Number),
			}),
			false,
		},
		{
			ObjectVal(map[string]Value{
				"a": NullVal(DynamicPseudoType),
			}),
			ObjectVal(map[string]Value{
				"a": DynamicVal,
			}),
			false,
		},
		{
			ObjectVal(map[string]Value{
				"a": NullVal(List(String)),
			}),
			// While the unknown val does contain dynamic types, the overall
			// container types can't conform.
			ObjectVal(map[string]Value{
				"a": UnknownVal(List(List(DynamicPseudoType))),
			}),
			false,
		},
		{
			ObjectVal(map[string]Value{
				"a": NullVal(List(List(String))),
			}),
			ObjectVal(map[string]Value{
				"a": UnknownVal(List(List(DynamicPseudoType))),
			}),
			false,
		},
		{
			ObjectVal(map[string]Value{
				"a": SetVal([]Value{
					ObjectVal(map[string]Value{
						"b": UnknownVal(String),
					}),
				}),
			}),
			ObjectVal(map[string]Value{
				"a": SetVal([]Value{
					ObjectVal(map[string]Value{
						"b": UnknownVal(String),
					}),
				}),
			}),
			true,
		},
		{
			ObjectVal(map[string]Value{
				"a": SetVal([]Value{
					ObjectVal(map[string]Value{
						"b": UnknownVal(String),
						"c": StringVal("cee"),
					}),
				}),
			}),
			ObjectVal(map[string]Value{
				"a": SetVal([]Value{
					ObjectVal(map[string]Value{
						"b": UnknownVal(String),
						"c": StringVal("cee"),
					}),
				}),
			}),
			true,
		},
		{
			ObjectVal(map[string]Value{
				"a": SetVal([]Value{
					ObjectVal(map[string]Value{
						"b": UnknownVal(String),
					}),
				}),
			}),
			ObjectVal(map[string]Value{
				"a": SetVal([]Value{
					ObjectVal(map[string]Value{
						"c": UnknownVal(String),
					}),
				}),
			}),
			false,
		},
		{
			UnknownVal(Bool).Refine().NotNull().NewValue(),
			UnknownVal(Bool),
			false,
		},
		{
			UnknownVal(Bool),
			UnknownVal(Bool).Refine().NotNull().NewValue(),
			false,
		},
		{
			UnknownVal(Number).Refine().NumberRangeInclusive(Zero, NumberIntVal(1)).NewValue(),
			UnknownVal(Number).Refine().NumberRangeInclusive(Zero, NumberIntVal(2)).NewValue(),
			false,
		},
		{
			UnknownVal(Number).Refine().NumberRangeInclusive(Zero, NumberIntVal(1)).NewValue(),
			UnknownVal(Number).Refine().NumberRangeInclusive(Zero, NumberIntVal(1)).NewValue(),
			true,
		},
		{
			UnknownVal(String),
			UnknownVal(String).Refine().StringPrefix("foo").NewValue(),
			false,
		},
		{
			UnknownVal(String).Refine().StringPrefix("foo").NewValue(),
			UnknownVal(String).Refine().StringPrefix("foo").NewValue(),
			true,
		},
		{
			UnknownVal(String).Refine().NotNull().StringPrefix("foo").NewValue(),
			UnknownVal(String).Refine().StringPrefix("foo").NewValue(),
			false,
		},
		{
			UnknownVal(String).Refine().StringPrefix("foo").NewValue(),
			UnknownVal(String).Refine().StringPrefix("bar").NewValue(),
			false,
		},
		{
			UnknownVal(String).Refine().Null().NewValue(),
			NullVal(String),
			true, // The refinement expression collapses into a simple null
		},
		{
			UnknownVal(Number).Refine().NumberRangeInclusive(Zero, Zero).NewValue(),
			Zero,
			false, // Refinement can't collapse to zero because it might be null
		},
		{
			UnknownVal(Number).Refine().NotNull().NumberRangeInclusive(Zero, Zero).NewValue(),
			Zero,
			true, // Refinement collapses to zero because it's not null and the two bounds are equal
		},
		{
			UnknownVal(List(String)).Refine().NotNull().CollectionLengthUpperBound(0).NewValue(),
			ListValEmpty(String),
			true, // Colection length lower bound is always at least zero so this refinement collapses to an empty list
		},

		// Marks
		{
			StringVal("a").Mark(1),
			StringVal("b"),
			false,
		},
		{
			StringVal("a"),
			StringVal("b").Mark(2),
			false,
		},
		{
			StringVal("a").Mark(1),
			StringVal("b").Mark(2),
			false,
		},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("%#v.RawEquals(%#v)", test.LHS, test.RHS), func(t *testing.T) {
			got := test.LHS.RawEquals(test.RHS)
			if !got == test.Expected {
				t.Fatalf("wrong RawEquals result\ngot:  %#v\nwant: %#v", got, test.Expected)
			}
		})
	}
}

func TestValueAdd(t *testing.T) {
	tests := []struct {
		LHS      Value
		RHS      Value
		Expected Value
	}{
		{
			NumberIntVal(1),
			NumberIntVal(2),
			NumberIntVal(3),
		},
		{
			NumberIntVal(1),
			NumberIntVal(-2),
			NumberIntVal(-1),
		},
		{
			NumberIntVal(1),
			NumberFloatVal(0.5),
			NumberFloatVal(1.5),
		},
		{
			NumberIntVal(1),
			UnknownVal(Number),
			UnknownVal(Number).RefineNotNull(),
		},
		{
			NumberIntVal(1),
			UnknownVal(Number).Refine().
				NumberRangeLowerBound(NumberIntVal(2), false).
				NewValue(),
			UnknownVal(Number).Refine().
				NotNull().
				NumberRangeLowerBound(NumberIntVal(3), true).
				NewValue(),
		},
		{
			Zero,
			UnknownVal(Number).Refine().
				NumberRangeLowerBound(NumberIntVal(2), false).
				NewValue(),
			UnknownVal(Number).Refine().
				NotNull().
				NumberRangeLowerBound(NumberIntVal(2), true).
				NewValue(),
		},
		{
			UnknownVal(Number).Refine().
				NumberRangeLowerBound(NumberIntVal(2), false).
				NewValue(),
			UnknownVal(Number).Refine().
				NumberRangeLowerBound(NumberIntVal(2), false).
				NewValue(),
			UnknownVal(Number).Refine().
				NotNull().
				NumberRangeLowerBound(NumberIntVal(4), true).
				NewValue(),
		},
		{
			UnknownVal(Number).Refine().
				NumberRangeLowerBound(NumberIntVal(1), true).
				NumberRangeUpperBound(NumberIntVal(2), false).
				NewValue(),
			UnknownVal(Number).Refine().
				NumberRangeLowerBound(NumberIntVal(2), false).
				NewValue(),
			UnknownVal(Number).Refine().
				NotNull().
				NumberRangeLowerBound(NumberIntVal(3), true).
				NewValue(),
		},
		{
			UnknownVal(Number).Refine().
				NumberRangeLowerBound(NumberIntVal(1), true).
				NumberRangeUpperBound(NumberIntVal(2), false).
				NewValue(),
			UnknownVal(Number).Refine().
				NumberRangeLowerBound(NumberIntVal(2), false).
				NumberRangeUpperBound(NumberIntVal(3), false).
				NewValue(),
			UnknownVal(Number).Refine().
				NotNull().
				NumberRangeLowerBound(NumberIntVal(3), true).
				NumberRangeUpperBound(NumberIntVal(5), true).
				NewValue(),
		},
		{
			UnknownVal(Number),
			UnknownVal(Number),
			UnknownVal(Number).RefineNotNull(),
		},
		{
			NumberIntVal(1),
			DynamicVal,
			UnknownVal(Number).RefineNotNull(),
		},
		{
			DynamicVal,
			DynamicVal,
			UnknownVal(Number).RefineNotNull(),
		},
		{
			Zero.Mark(1),
			Zero,
			Zero.Mark(1),
		},
		{
			Zero,
			Zero.Mark(2),
			Zero.Mark(2),
		},
		{
			Zero.Mark(1),
			Zero.Mark(2),
			Zero.WithMarks(NewValueMarks(1, 2)),
		},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("%#v.Add(%#v)", test.LHS, test.RHS), func(t *testing.T) {
			got := test.LHS.Add(test.RHS)
			if !got.RawEquals(test.Expected) {
				t.Fatalf("Wrong result\ngot:  %#v\nwant: %#v", got, test.Expected)
			}
		})
	}
}

func TestValueSubtract(t *testing.T) {
	tests := []struct {
		LHS      Value
		RHS      Value
		Expected Value
	}{
		{
			NumberIntVal(1),
			NumberIntVal(2),
			NumberIntVal(-1),
		},
		{
			NumberIntVal(1),
			NumberIntVal(-2),
			NumberIntVal(3),
		},
		{
			NumberIntVal(1),
			NumberFloatVal(0.5),
			NumberFloatVal(0.5),
		},
		{
			NumberIntVal(1),
			UnknownVal(Number),
			UnknownVal(Number).RefineNotNull(),
		},
		{
			UnknownVal(Number),
			UnknownVal(Number),
			UnknownVal(Number).RefineNotNull(),
		},
		{
			NumberIntVal(1),
			UnknownVal(Number).Refine().
				NumberRangeLowerBound(NumberIntVal(2), true).
				NewValue(),
			UnknownVal(Number).Refine().
				NotNull().
				NumberRangeUpperBound(NumberIntVal(-1), true).
				NewValue(),
		},
		{
			Zero,
			UnknownVal(Number).Refine().
				NumberRangeLowerBound(NumberIntVal(2), true).
				NewValue(),
			UnknownVal(Number).Refine().
				NotNull().
				NumberRangeUpperBound(NumberIntVal(-2), true).
				NewValue(),
		},
		{
			UnknownVal(Number).Refine().
				NumberRangeLowerBound(NumberIntVal(2), true).
				NewValue(),
			UnknownVal(Number).Refine().
				NumberRangeLowerBound(NumberIntVal(2), true).
				NewValue(),
			UnknownVal(Number).RefineNotNull(), // We don't currently refine this case
		},
		{
			UnknownVal(Number).Refine().
				NumberRangeLowerBound(NumberIntVal(1), true).
				NumberRangeUpperBound(NumberIntVal(2), false).
				NewValue(),
			UnknownVal(Number).Refine().
				NumberRangeLowerBound(NumberIntVal(2), true).
				NewValue(),
			UnknownVal(Number).Refine().
				NotNull().
				NumberRangeUpperBound(NumberIntVal(0), true).
				NewValue(),
		},
		{
			UnknownVal(Number).Refine().
				NumberRangeLowerBound(NumberIntVal(1), true).
				NumberRangeUpperBound(NumberIntVal(2), false).
				NewValue(),
			UnknownVal(Number).Refine().
				NumberRangeLowerBound(NumberIntVal(2), false).
				NumberRangeUpperBound(NumberIntVal(3), false).
				NewValue(),
			UnknownVal(Number).Refine().
				NotNull().
				NumberRangeLowerBound(NumberIntVal(-2), true).
				NumberRangeUpperBound(NumberIntVal(0), true).
				NewValue(),
		},
		{
			NumberIntVal(1),
			DynamicVal,
			UnknownVal(Number).RefineNotNull(),
		},
		{
			DynamicVal,
			DynamicVal,
			UnknownVal(Number).RefineNotNull(),
		},
		{
			Zero.Mark(1),
			Zero,
			Zero.Mark(1),
		},
		{
			Zero,
			Zero.Mark(2),
			Zero.Mark(2),
		},
		{
			Zero.Mark(1),
			Zero.Mark(2),
			Zero.WithMarks(NewValueMarks(1, 2)),
		},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("%#v.Subtract(%#v)", test.LHS, test.RHS), func(t *testing.T) {
			got := test.LHS.Subtract(test.RHS)
			if !got.RawEquals(test.Expected) {
				t.Fatalf("wrong result\ngot:  %#v\nwant: %#v", got, test.Expected)
			}
		})
	}
}

func TestValueNegate(t *testing.T) {
	tests := []struct {
		Receiver Value
		Expected Value
	}{
		{
			NumberIntVal(1),
			NumberIntVal(-1),
		},
		{
			NumberFloatVal(0.5),
			NumberFloatVal(-0.5),
		},
		{
			UnknownVal(Number),
			UnknownVal(Number).RefineNotNull(),
		},
		{
			DynamicVal,
			UnknownVal(Number).RefineNotNull(),
		},
		{
			Zero.Mark(1),
			Zero.Mark(1),
		},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("%#v.Negate()", test.Receiver), func(t *testing.T) {
			got := test.Receiver.Negate()
			if !got.RawEquals(test.Expected) {
				t.Fatalf("wrong result\ngot:  %#v\nwant: %#v", got, test.Expected)
			}
		})
	}
}

func TestValueMultiply(t *testing.T) {
	tests := []struct {
		LHS      Value
		RHS      Value
		Expected Value
	}{
		{
			NumberIntVal(4),
			NumberIntVal(2),
			NumberIntVal(8),
		},
		{
			NumberIntVal(1),
			NumberIntVal(-2),
			NumberIntVal(-2),
		},
		{
			NumberIntVal(5),
			NumberFloatVal(0.5),
			NumberFloatVal(2.5),
		},
		{
			NumberIntVal(1),
			UnknownVal(Number),
			UnknownVal(Number).RefineNotNull(),
		},
		{
			UnknownVal(Number),
			UnknownVal(Number),
			UnknownVal(Number).RefineNotNull(),
		},
		{
			NumberIntVal(3),
			UnknownVal(Number).Refine().
				NumberRangeLowerBound(NumberIntVal(2), false).
				NewValue(),
			UnknownVal(Number).Refine().
				NotNull().
				NumberRangeLowerBound(NumberIntVal(6), true).
				NewValue(),
		},
		{
			Zero,
			UnknownVal(Number),
			Zero,
		},
		{
			UnknownVal(Number),
			Zero,
			Zero,
		},
		{
			Zero,
			UnknownVal(Number).Refine().
				NumberRangeLowerBound(NumberIntVal(2), false).
				NewValue(),
			Zero,
		},
		{
			UnknownVal(Number).Refine().
				NumberRangeLowerBound(NumberIntVal(2), false).
				NewValue(),
			UnknownVal(Number).Refine().
				NumberRangeLowerBound(NumberIntVal(4), false).
				NewValue(),
			UnknownVal(Number).Refine().
				NotNull().
				NumberRangeLowerBound(NumberIntVal(8), true).
				NewValue(),
		},
		{
			UnknownVal(Number).Refine().
				NumberRangeLowerBound(NumberIntVal(3), true).
				NumberRangeUpperBound(NumberIntVal(4), false).
				NewValue(),
			UnknownVal(Number).Refine().
				NumberRangeLowerBound(NumberIntVal(2), false).
				NewValue(),
			UnknownVal(Number).Refine().
				NotNull().
				NumberRangeLowerBound(NumberIntVal(6), true).
				NewValue(),
		},
		{
			UnknownVal(Number).Refine().
				NumberRangeLowerBound(NumberIntVal(1), true).
				NumberRangeUpperBound(NumberIntVal(2), false).
				NewValue(),
			UnknownVal(Number).Refine().
				NumberRangeLowerBound(NumberIntVal(2), false).
				NumberRangeUpperBound(NumberIntVal(3), false).
				NewValue(),
			UnknownVal(Number).Refine().
				NotNull().
				NumberRangeLowerBound(NumberIntVal(2), true).
				NumberRangeUpperBound(NumberIntVal(6), true).
				NewValue(),
		},
		{
			UnknownVal(Number).Refine().
				NumberRangeLowerBound(NumberIntVal(1), true).
				NumberRangeUpperBound(NumberIntVal(2), false).
				NewValue(),
			Zero,
			Zero, // deduced by refinement
		},
		{
			NumberIntVal(1),
			DynamicVal,
			UnknownVal(Number).RefineNotNull(),
		},
		{
			DynamicVal,
			DynamicVal,
			UnknownVal(Number).RefineNotNull(),
		},
		{
			Zero.Mark(1),
			Zero,
			Zero.Mark(1),
		},
		{
			Zero,
			Zero.Mark(2),
			Zero.Mark(2),
		},
		{
			Zero.Mark(1),
			Zero.Mark(2),
			Zero.WithMarks(NewValueMarks(1, 2)),
		},
		{
			MustParseNumberVal("967323432120515089486873574508975134568969931547"),
			NumberFloatVal(12345),
			MustParseNumberVal("11941607769527758779715454277313298036253933804947715"),
		},
		{
			NumberFloatVal(22337203685475.5),
			NumberFloatVal(22337203685475.5),
			MustParseNumberVal("498950668486420259929661100.2"),
		},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("%#v.Multiply(%#v)", test.LHS, test.RHS), func(t *testing.T) {
			got := test.LHS.Multiply(test.RHS)
			if !got.RawEquals(test.Expected) {
				t.Fatalf("wrong result\ngot:  %#v\nwant: %#v", got, test.Expected)
			}
		})
	}
}

func TestValueDivide(t *testing.T) {
	tests := []struct {
		LHS      Value
		RHS      Value
		Expected Value
	}{
		{
			NumberIntVal(10),
			NumberIntVal(2),
			NumberIntVal(5),
		},
		{
			NumberIntVal(1),
			NumberIntVal(-2),
			NumberFloatVal(-0.5),
		},
		{
			NumberIntVal(5),
			NumberFloatVal(0.5),
			NumberIntVal(10),
		},
		{
			NumberIntVal(5),
			NumberIntVal(0),
			PositiveInfinity,
		},
		{
			NumberIntVal(-5),
			NumberIntVal(0),
			NegativeInfinity,
		},
		{
			NumberIntVal(1),
			UnknownVal(Number),
			UnknownVal(Number).RefineNotNull(),
		},
		{
			UnknownVal(Number),
			UnknownVal(Number),
			UnknownVal(Number).RefineNotNull(),
		},
		{
			NumberIntVal(1),
			DynamicVal,
			UnknownVal(Number).RefineNotNull(),
		},
		{
			DynamicVal,
			DynamicVal,
			UnknownVal(Number).RefineNotNull(),
		},
		{
			Zero.Mark(1),
			NumberIntVal(1),
			Zero.Mark(1),
		},
		{
			Zero,
			NumberIntVal(1).Mark(2),
			Zero.Mark(2),
		},
		{
			Zero.Mark(1),
			NumberIntVal(1).Mark(2),
			Zero.WithMarks(NewValueMarks(1, 2)),
		},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("%#v.Divide(%#v)", test.LHS, test.RHS), func(t *testing.T) {
			got := test.LHS.Divide(test.RHS)
			if !got.RawEquals(test.Expected) {
				t.Fatalf("wrong result\ngot:  %#v\nwant: %#v", got, test.Expected)
			}
		})
	}
}

func TestValueModulo(t *testing.T) {
	tests := []struct {
		LHS      Value
		RHS      Value
		Expected Value
	}{
		{
			NumberIntVal(10),
			NumberIntVal(2),
			NumberIntVal(0),
		},
		{
			NumberIntVal(-10),
			NumberIntVal(2),
			NumberIntVal(0),
		},
		{
			NumberIntVal(11),
			NumberIntVal(2),
			NumberIntVal(1),
		},
		{
			NumberIntVal(-11),
			NumberIntVal(2),
			NumberIntVal(-1),
		},
		{
			NumberIntVal(1),
			NumberIntVal(-2),
			NumberFloatVal(1),
		},
		{
			NumberIntVal(5),
			NumberFloatVal(0.5),
			NumberIntVal(0),
		},
		{
			NumberIntVal(5),
			NumberFloatVal(1.5),
			NumberFloatVal(0.5),
		},
		{
			NumberIntVal(5),
			NumberIntVal(0),
			NumberIntVal(5),
		},
		{
			NumberIntVal(-5),
			NumberIntVal(0),
			NumberIntVal(-5),
		},
		{
			NumberIntVal(1),
			UnknownVal(Number),
			UnknownVal(Number).RefineNotNull(),
		},
		{
			UnknownVal(Number),
			UnknownVal(Number),
			UnknownVal(Number).RefineNotNull(),
		},
		{
			NumberIntVal(1),
			DynamicVal,
			UnknownVal(Number).RefineNotNull(),
		},
		{
			DynamicVal,
			DynamicVal,
			UnknownVal(Number).RefineNotNull(),
		},
		{
			NumberIntVal(10).Mark(1),
			NumberIntVal(10),
			Zero.Mark(1),
		},
		{
			NumberIntVal(10),
			NumberIntVal(10).Mark(2),
			Zero.Mark(2),
		},
		{
			NumberIntVal(10).Mark(1),
			NumberIntVal(10).Mark(2),
			Zero.WithMarks(NewValueMarks(1, 2)),
		},
		{
			MustParseNumberVal("967323432120515089486873574508975134568969931547"),
			NumberIntVal(10),
			NumberIntVal(7),
		},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("%#v.Modulo(%#v)", test.LHS, test.RHS), func(t *testing.T) {
			got := test.LHS.Modulo(test.RHS)
			if !got.RawEquals(test.Expected) {
				t.Fatalf("Modulo returned %#v; want %#v", got, test.Expected)
			}
		})
	}
}

func TestValueAbsolute(t *testing.T) {
	tests := []struct {
		Receiver Value
		Expected Value
	}{
		{
			NumberIntVal(1),
			NumberIntVal(1),
		},
		{
			NumberIntVal(-1),
			NumberIntVal(1),
		},
		{
			NumberFloatVal(0.5),
			NumberFloatVal(0.5),
		},
		{
			NumberFloatVal(-0.5),
			NumberFloatVal(0.5),
		},
		{
			PositiveInfinity,
			PositiveInfinity,
		},
		{
			NegativeInfinity,
			PositiveInfinity,
		},
		{
			UnknownVal(Number),
			UnknownVal(Number).Refine().NotNull().NumberRangeInclusive(Zero, UnknownVal(Number)).NewValue(),
		},
		{
			DynamicVal,
			UnknownVal(Number).Refine().NotNull().NumberRangeInclusive(Zero, UnknownVal(Number)).NewValue(),
		},
		{
			NumberIntVal(-1).Mark(1),
			NumberIntVal(1).Mark(1),
		},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("%#v.Absolute()", test.Receiver), func(t *testing.T) {
			got := test.Receiver.Absolute()
			if !got.RawEquals(test.Expected) {
				t.Fatalf("wrong result\ngot:  %#v\nwant: %#v", got, test.Expected)
			}
		})
	}
}

func TestValueGetAttr(t *testing.T) {
	tests := []struct {
		Object   Value
		AttrName string
		Expected Value
	}{
		{
			ObjectVal(map[string]Value{
				"greeting": StringVal("hello"),
			}),
			"greeting",
			StringVal("hello"),
		},
		{
			ObjectVal(map[string]Value{
				"greeting": StringVal("hello"),
			}),
			"greeting",
			StringVal("hello"),
		},
		{
			UnknownVal(Object(map[string]Type{
				"gr\u00e9eting": String, // precombined é
			})),
			"gre\u0301eting", // e with combining acute accent
			UnknownVal(String),
		},
		{
			DynamicVal,
			"hello",
			DynamicVal,
		},
		{
			ObjectVal(map[string]Value{
				"greeting": StringVal("hello"),
			}).Mark(1),
			"greeting",
			StringVal("hello").Mark(1),
		},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("%#v.GetAttr(%q)", test.Object, test.AttrName), func(t *testing.T) {
			got := test.Object.GetAttr(test.AttrName)
			if !got.RawEquals(test.Expected) {
				t.Fatalf("GetAttr returned %#v; want %#v", got, test.Expected)
			}
		})
	}
}

func TestValueIndex(t *testing.T) {
	tests := []struct {
		Collection Value
		Key        Value
		Expected   Value
	}{
		{
			ListVal([]Value{StringVal("hello")}),
			NumberIntVal(0),
			StringVal("hello"),
		},
		{
			ListVal([]Value{StringVal("hello"), StringVal("world")}),
			NumberIntVal(1),
			StringVal("world"),
		},
		{
			ListVal([]Value{StringVal("hello")}),
			UnknownVal(Number),
			UnknownVal(String),
		},
		{
			ListVal([]Value{StringVal("hello")}),
			DynamicVal,
			UnknownVal(String),
		},
		{
			UnknownVal(List(String)),
			NumberIntVal(0),
			UnknownVal(String),
		},
		{
			MapVal(map[string]Value{"greeting": StringVal("hello")}),
			StringVal("greeting"),
			StringVal("hello"),
		},
		{
			MapVal(map[string]Value{"gr\u00e9eting": StringVal("hello")}), // precombined é
			StringVal("gre\u0301eting"),                                   // e with combining acute accent
			StringVal("hello"),
		},
		{
			MapVal(map[string]Value{"greeting": True}),
			UnknownVal(String),
			UnknownVal(Bool),
		},
		{
			MapVal(map[string]Value{"greeting": True}),
			DynamicVal,
			UnknownVal(Bool),
		},
		{
			UnknownVal(Map(String)),
			StringVal("greeting"),
			UnknownVal(String),
		},
		{
			DynamicVal,
			StringVal("hello"),
			DynamicVal,
		},
		{
			DynamicVal,
			NumberIntVal(0),
			DynamicVal,
		},
		{
			TupleVal([]Value{StringVal("hello")}),
			NumberIntVal(0),
			StringVal("hello"),
		},
		{
			TupleVal([]Value{StringVal("hello"), NumberIntVal(5)}),
			NumberIntVal(0),
			StringVal("hello"),
		},
		{
			TupleVal([]Value{StringVal("hello"), NumberIntVal(5)}),
			NumberIntVal(1),
			NumberIntVal(5),
		},
		{
			TupleVal([]Value{StringVal("hello"), DynamicVal}),
			NumberIntVal(0),
			StringVal("hello"),
		},
		{
			TupleVal([]Value{StringVal("hello"), DynamicVal}),
			NumberIntVal(1),
			DynamicVal,
		},
		{
			TupleVal([]Value{StringVal("hello"), UnknownVal(Number)}),
			NumberIntVal(0),
			StringVal("hello"),
		},
		{
			TupleVal([]Value{StringVal("hello"), UnknownVal(Number)}),
			NumberIntVal(1),
			UnknownVal(Number),
		},
		{
			TupleVal([]Value{StringVal("hello"), UnknownVal(Number)}),
			UnknownVal(Number),
			DynamicVal,
		},
		{
			UnknownVal(Tuple([]Type{String})),
			NumberIntVal(0),
			UnknownVal(String),
		},
		{
			ListVal([]Value{StringVal("hello")}).Mark(1),
			NumberIntVal(0),
			StringVal("hello").Mark(1),
		},
		{
			ListVal([]Value{StringVal("hello")}),
			NumberIntVal(0).Mark(1),
			StringVal("hello").Mark(1),
		},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("%#v.Index(%q)", test.Collection, test.Key), func(t *testing.T) {
			got := test.Collection.Index(test.Key)
			if !got.RawEquals(test.Expected) {
				t.Fatalf("Index returned %#v; want %#v", got, test.Expected)
			}
		})
	}
}

func TestValueHasIndex(t *testing.T) {
	tests := []struct {
		Collection Value
		Key        Value
		Expected   Value
	}{
		{
			ListVal([]Value{StringVal("hello")}),
			NumberIntVal(0),
			True,
		},
		{
			ListVal([]Value{StringVal("hello"), StringVal("world")}),
			NumberIntVal(1),
			True,
		},
		{
			ListVal([]Value{StringVal("hello"), StringVal("world")}),
			NumberIntVal(2),
			False,
		},
		{
			ListVal([]Value{StringVal("hello"), StringVal("world")}),
			NumberIntVal(-1),
			False,
		},
		{
			ListVal([]Value{StringVal("hello"), StringVal("world")}),
			NumberFloatVal(0.5),
			False,
		},
		{
			ListVal([]Value{StringVal("hello"), StringVal("world")}),
			StringVal("greeting"),
			False,
		},
		{
			ListVal([]Value{StringVal("hello"), StringVal("world")}),
			True,
			False,
		},
		{
			ListVal([]Value{StringVal("hello")}),
			UnknownVal(Number),
			UnknownVal(Bool).RefineNotNull(),
		},
		{
			ListVal([]Value{StringVal("hello")}),
			DynamicVal,
			UnknownVal(Bool).RefineNotNull(),
		},
		{
			UnknownVal(List(String)),
			NumberIntVal(0),
			UnknownVal(Bool).RefineNotNull(),
		},
		{
			UnknownVal(List(String)),
			StringVal("hello"),
			False,
		},
		{
			MapVal(map[string]Value{"greeting": StringVal("hello")}),
			StringVal("greeting"),
			True,
		},
		{
			MapVal(map[string]Value{"gre\u0301eting": StringVal("hello")}), // e with combining acute accent
			StringVal("gr\u00e9eting"),                                     // precombined é
			True,
		},
		{
			MapVal(map[string]Value{"greeting": StringVal("hello")}),
			StringVal("grouting"),
			False,
		},
		{
			MapVal(map[string]Value{"greeting": StringVal("hello")}),
			StringVal(""),
			False,
		},
		{
			MapVal(map[string]Value{"greeting": StringVal("hello")}),
			Zero,
			False,
		},
		{
			MapVal(map[string]Value{"greeting": StringVal("hello")}),
			True,
			False,
		},
		{
			MapVal(map[string]Value{"greeting": StringVal("hello")}),
			UnknownVal(String),
			UnknownVal(Bool).RefineNotNull(),
		},
		{
			MapVal(map[string]Value{"greeting": StringVal("hello")}),
			DynamicVal,
			UnknownVal(Bool).RefineNotNull(),
		},
		{
			UnknownVal(Map(String)),
			StringVal("hello"),
			UnknownVal(Bool).RefineNotNull(),
		},
		{
			UnknownVal(Map(String)),
			NumberIntVal(0),
			False,
		},
		{
			TupleVal([]Value{StringVal("hello")}),
			NumberIntVal(0),
			True,
		},
		{
			TupleVal([]Value{StringVal("hello"), StringVal("world")}),
			NumberIntVal(1),
			True,
		},
		{
			TupleVal([]Value{StringVal("hello"), StringVal("world")}),
			NumberIntVal(2),
			False,
		},
		{
			TupleVal([]Value{StringVal("hello"), StringVal("world")}),
			NumberIntVal(-1),
			False,
		},
		{
			TupleVal([]Value{StringVal("hello"), StringVal("world")}),
			NumberFloatVal(0.5),
			False,
		},
		{
			TupleVal([]Value{StringVal("hello"), StringVal("world")}),
			StringVal("greeting"),
			False,
		},
		{
			TupleVal([]Value{StringVal("hello"), StringVal("world")}),
			True,
			False,
		},
		{
			TupleVal([]Value{StringVal("hello")}),
			UnknownVal(Number),
			UnknownVal(Bool).RefineNotNull(),
		},
		{
			UnknownVal(Tuple([]Type{String})),
			NumberIntVal(0),
			True,
		},
		{
			TupleVal([]Value{StringVal("hello")}),
			DynamicVal,
			UnknownVal(Bool).RefineNotNull(),
		},
		{
			DynamicVal,
			StringVal("hello"),
			UnknownVal(Bool).RefineNotNull(),
		},
		{
			DynamicVal,
			NumberIntVal(0),
			UnknownVal(Bool).RefineNotNull(),
		},
		{
			ListVal([]Value{StringVal("hello")}).Mark(1),
			NumberIntVal(0),
			True.Mark(1),
		},
		{
			ListVal([]Value{StringVal("hello")}),
			NumberIntVal(0).Mark(1),
			True.Mark(1),
		},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("%#v.HasIndex(%q)", test.Collection, test.Key), func(t *testing.T) {
			got := test.Collection.HasIndex(test.Key)
			if !got.RawEquals(test.Expected) {
				t.Fatalf("HasIndex returned %#v; want %#v", got, test.Expected)
			}
		})
	}
}

func TestValueForEachElement(t *testing.T) {
	type call struct {
		Key     Value
		Element Value
	}
	tests := []struct {
		Receiver Value
		Expected []call
		Stopped  bool
	}{
		{
			ListValEmpty(String),
			[]call{},
			false,
		},
		{
			ListVal([]Value{
				NumberIntVal(1),
				NumberIntVal(2),
			}),
			[]call{
				{NumberIntVal(0), NumberIntVal(1)},
				{NumberIntVal(1), NumberIntVal(2)},
			},
			false,
		},
		{
			ListVal([]Value{
				StringVal("hey"),
				StringVal("stop"),
				StringVal("hey"),
			}),
			[]call{
				{NumberIntVal(0), StringVal("hey")},
				{NumberIntVal(1), StringVal("stop")},
			},
			true,
		},
		{
			SetValEmpty(String),
			[]call{},
			false,
		},
		{
			SetVal([]Value{
				NumberIntVal(1),
				NumberIntVal(10),
				NumberIntVal(2),
			}),
			[]call{
				// Numbers in sets are always iterated in numerical order.
				{NumberIntVal(1), NumberIntVal(1)},
				{NumberIntVal(2), NumberIntVal(2)},
				{NumberIntVal(10), NumberIntVal(10)},
			},
			false,
		},
		{
			SetVal([]Value{
				StringVal("hi"),
				StringVal("stop"),
				StringVal("zzz"),
			}),
			[]call{
				// Strings in sets are always iterated in lexicographical order.
				{StringVal("hi"), StringVal("hi")},
				{StringVal("stop"), StringVal("stop")},
			},
			true,
		},
		{
			MapVal(map[string]Value{
				"second": NumberIntVal(2),
				"first":  NumberIntVal(1),
			}),
			[]call{
				{StringVal("first"), NumberIntVal(1)},
				{StringVal("second"), NumberIntVal(2)},
			},
			false,
		},
		{
			MapVal(map[string]Value{
				"item2": StringVal("value2"),
				"item1": StringVal("stop"),
				"item0": StringVal("value0"),
			}),
			[]call{
				{StringVal("item0"), StringVal("value0")},
				{StringVal("item1"), StringVal("stop")},
			},
			true,
		},
		{
			EmptyTupleVal,
			[]call{},
			false,
		},
		{
			TupleVal([]Value{
				StringVal("hello"),
				NumberIntVal(2),
			}),
			[]call{
				{NumberIntVal(0), StringVal("hello")},
				{NumberIntVal(1), NumberIntVal(2)},
			},
			false,
		},
		{
			TupleVal([]Value{
				NumberIntVal(5),
				StringVal("stop"),
				True,
			}),
			[]call{
				{NumberIntVal(0), NumberIntVal(5)},
				{NumberIntVal(1), StringVal("stop")},
			},
			true,
		},
		{
			EmptyObjectVal,
			[]call{},
			false,
		},
		{
			ObjectVal(map[string]Value{
				"bool":   True,
				"string": StringVal("hello"),
			}),
			[]call{
				{StringVal("bool"), True},
				{StringVal("string"), StringVal("hello")},
			},
			false,
		},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("%#v.ForEachElement()", test.Receiver), func(t *testing.T) {
			calls := make([]call, 0)
			stopped := test.Receiver.ForEachElement(
				func(key Value, elem Value) (stop bool) {
					calls = append(calls, call{
						Key:     key,
						Element: elem,
					})
					if elem.v == "stop" {
						stop = true
					}
					return
				},
			)
			if !reflect.DeepEqual(calls, test.Expected) {
				t.Errorf(
					"wrong calls from ForEachElement\ngot:  %#v\nwant: %#v",
					calls, test.Expected,
				)
			}
			if stopped != test.Stopped {
				t.Errorf(
					"ForEachElement returned %#v; want %#v",
					stopped, test.Stopped,
				)
			}
		})
	}
}

func TestValueNot(t *testing.T) {
	tests := []struct {
		Receiver Value
		Expected Value
	}{
		{
			True,
			False,
		},
		{
			False,
			True,
		},
		{
			UnknownVal(Bool),
			UnknownVal(Bool).RefineNotNull(),
		},
		{
			DynamicVal,
			UnknownVal(Bool).RefineNotNull(),
		},
		{
			True.Mark(1),
			False.Mark(1),
		},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("%#v.Not()", test.Receiver), func(t *testing.T) {
			got := test.Receiver.Not()
			if !got.RawEquals(test.Expected) {
				t.Fatalf("Not returned %#v; want %#v", got, test.Expected)
			}
		})
	}
}

func TestValueAnd(t *testing.T) {
	tests := []struct {
		Receiver Value
		Other    Value
		Expected Value
	}{
		{
			False,
			False,
			False,
		},
		{
			False,
			True,
			False,
		},
		{
			True,
			False,
			False,
		},
		{
			True,
			True,
			True,
		},
		{
			UnknownVal(Bool),
			UnknownVal(Bool),
			UnknownVal(Bool).RefineNotNull(),
		},
		{
			True,
			UnknownVal(Bool),
			UnknownVal(Bool).RefineNotNull(),
		},
		{
			UnknownVal(Bool),
			True,
			UnknownVal(Bool).RefineNotNull(),
		},
		{
			False,
			UnknownVal(Bool),
			False,
		},
		{
			UnknownVal(Bool),
			False,
			False,
		},
		{
			DynamicVal,
			DynamicVal,
			UnknownVal(Bool).RefineNotNull(),
		},
		{
			True,
			DynamicVal,
			UnknownVal(Bool).RefineNotNull(),
		},
		{
			DynamicVal,
			True,
			UnknownVal(Bool).RefineNotNull(),
		},
		{
			False,
			DynamicVal,
			False,
		},
		{
			DynamicVal,
			False,
			False,
		},
		{
			True.Mark(1),
			True,
			True.Mark(1),
		},
		{
			True,
			True.Mark(1),
			True.Mark(1),
		},
		{
			True.Mark(1),
			True.Mark(1),
			True.Mark(1),
		},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("%#v.And(%#v)", test.Receiver, test.Other), func(t *testing.T) {
			got := test.Receiver.And(test.Other)
			if !got.RawEquals(test.Expected) {
				t.Fatalf("And returned %#v; want %#v", got, test.Expected)
			}
		})
	}
}

func TestValueOr(t *testing.T) {
	tests := []struct {
		Receiver Value
		Other    Value
		Expected Value
	}{
		{
			False,
			False,
			False,
		},
		{
			False,
			True,
			True,
		},
		{
			True,
			False,
			True,
		},
		{
			True,
			True,
			True,
		},
		{
			UnknownVal(Bool),
			UnknownVal(Bool),
			UnknownVal(Bool).RefineNotNull(),
		},
		{
			True,
			UnknownVal(Bool),
			True,
		},
		{
			UnknownVal(Bool),
			True,
			True,
		},
		{
			False,
			UnknownVal(Bool),
			UnknownVal(Bool).RefineNotNull(),
		},
		{
			UnknownVal(Bool),
			False,
			UnknownVal(Bool).RefineNotNull(),
		},
		{
			DynamicVal,
			DynamicVal,
			UnknownVal(Bool).RefineNotNull(),
		},
		{
			True,
			DynamicVal,
			True,
		},
		{
			DynamicVal,
			True,
			True,
		},
		{
			False,
			DynamicVal,
			UnknownVal(Bool).RefineNotNull(),
		},
		{
			DynamicVal,
			False,
			UnknownVal(Bool).RefineNotNull(),
		},
		{
			True.Mark(1),
			False,
			True.Mark(1),
		},
		{
			True,
			False.Mark(1),
			True.Mark(1),
		},
		{
			True.Mark(1),
			False.Mark(1),
			True.Mark(1),
		},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("%#v.Or(%#v)", test.Receiver, test.Other), func(t *testing.T) {
			got := test.Receiver.Or(test.Other)
			if !got.RawEquals(test.Expected) {
				t.Fatalf("Or returned %#v; want %#v", got, test.Expected)
			}
		})
	}
}

func TestLessThan(t *testing.T) {
	tests := []struct {
		Receiver Value
		Other    Value
		Expected Value
	}{
		{
			NumberIntVal(0),
			NumberIntVal(1),
			True,
		},
		{
			NumberIntVal(1),
			NumberIntVal(0),
			False,
		},
		{
			NumberIntVal(0),
			NumberIntVal(0),
			False,
		},
		{
			NumberFloatVal(0.1),
			NumberFloatVal(0.2),
			True,
		},
		{
			NumberFloatVal(0.2),
			NumberFloatVal(0.1),
			False,
		},
		{
			NumberIntVal(0),
			NumberFloatVal(0.2),
			True,
		},
		{
			NumberFloatVal(0.2),
			NumberIntVal(0),
			False,
		},
		{
			NumberFloatVal(0.2),
			NumberFloatVal(0.2),
			False,
		},
		{
			UnknownVal(Number),
			UnknownVal(Number),
			UnknownVal(Bool).RefineNotNull(),
		},
		{
			NumberIntVal(1),
			UnknownVal(Number),
			UnknownVal(Bool).RefineNotNull(),
		},
		{
			UnknownVal(Number),
			NumberIntVal(1),
			UnknownVal(Bool).RefineNotNull(),
		},
		{
			UnknownVal(Number).Refine().NumberRangeUpperBound(Zero, true).NewValue(),
			NumberIntVal(1),
			True, // Deduction from the refinement
		},
		{
			UnknownVal(Number).Refine().NumberRangeLowerBound(NumberIntVal(2), true).NewValue(),
			NumberIntVal(1),
			False, // Deduction from the refinement
		},
		{
			DynamicVal,
			DynamicVal,
			UnknownVal(Bool).RefineNotNull(),
		},
		{
			NumberIntVal(1),
			DynamicVal,
			UnknownVal(Bool).RefineNotNull(),
		},
		{
			DynamicVal,
			NumberIntVal(1),
			UnknownVal(Bool).RefineNotNull(),
		},
		{
			NumberIntVal(0).Mark(1),
			NumberIntVal(1),
			True.Mark(1),
		},
		{
			NumberIntVal(0),
			NumberIntVal(1).Mark(1),
			True.Mark(1),
		},
		{
			NumberIntVal(0).Mark(1),
			NumberIntVal(1).Mark(1),
			True.Mark(1),
		},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("%#v.LessThan(%#v)", test.Receiver, test.Other), func(t *testing.T) {
			got := test.Receiver.LessThan(test.Other)
			if !got.RawEquals(test.Expected) {
				t.Fatalf("LessThan returned %#v; want %#v", got, test.Expected)
			}
		})
	}
}

func TestGreaterThan(t *testing.T) {
	tests := []struct {
		Receiver Value
		Other    Value
		Expected Value
	}{
		{
			NumberIntVal(0),
			NumberIntVal(1),
			False,
		},
		{
			NumberIntVal(1),
			NumberIntVal(0),
			True,
		},
		{
			NumberIntVal(0),
			NumberIntVal(0),
			False,
		},
		{
			NumberFloatVal(0.1),
			NumberFloatVal(0.2),
			False,
		},
		{
			NumberFloatVal(0.2),
			NumberFloatVal(0.1),
			True,
		},
		{
			NumberIntVal(0),
			NumberFloatVal(0.2),
			False,
		},
		{
			NumberFloatVal(0.2),
			NumberIntVal(0),
			True,
		},
		{
			NumberFloatVal(0.2),
			NumberFloatVal(0.2),
			False,
		},
		{
			UnknownVal(Number),
			UnknownVal(Number),
			UnknownVal(Bool).RefineNotNull(),
		},
		{
			NumberIntVal(1),
			UnknownVal(Number),
			UnknownVal(Bool).RefineNotNull(),
		},
		{
			UnknownVal(Number),
			NumberIntVal(1),
			UnknownVal(Bool).RefineNotNull(),
		},
		{
			UnknownVal(Number).Refine().NumberRangeLowerBound(NumberIntVal(2), true).NewValue(),
			NumberIntVal(1),
			True, // Deduction based on the refinements
		},
		{
			UnknownVal(Number).Refine().NumberRangeUpperBound(NumberIntVal(0), true).NewValue(),
			NumberIntVal(1),
			False, // Deduction based on the refinements
		},
		{
			DynamicVal,
			DynamicVal,
			UnknownVal(Bool).RefineNotNull(),
		},
		{
			NumberIntVal(1),
			DynamicVal,
			UnknownVal(Bool).RefineNotNull(),
		},
		{
			DynamicVal,
			NumberIntVal(1),
			UnknownVal(Bool).RefineNotNull(),
		},
		{
			NumberIntVal(1).Mark(1),
			NumberIntVal(0),
			True.Mark(1),
		},
		{
			NumberIntVal(1),
			NumberIntVal(0).Mark(1),
			True.Mark(1),
		},
		{
			NumberIntVal(1).Mark(1),
			NumberIntVal(0).Mark(1),
			True.Mark(1),
		},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("%#v.GreaterThan(%#v)", test.Receiver, test.Other), func(t *testing.T) {
			got := test.Receiver.GreaterThan(test.Other)
			if !got.RawEquals(test.Expected) {
				t.Fatalf("GreaterThan returned %#v; want %#v", got, test.Expected)
			}
		})
	}
}

func TestLessThanOrEqualTo(t *testing.T) {
	tests := []struct {
		Receiver Value
		Other    Value
		Expected Value
	}{
		{
			NumberIntVal(0),
			NumberIntVal(1),
			True,
		},
		{
			NumberIntVal(1),
			NumberIntVal(0),
			False,
		},
		{
			NumberIntVal(0),
			NumberIntVal(0),
			True,
		},
		{
			NumberFloatVal(0.1),
			NumberFloatVal(0.2),
			True,
		},
		{
			NumberFloatVal(0.2),
			NumberFloatVal(0.1),
			False,
		},
		{
			NumberIntVal(0),
			NumberFloatVal(0.2),
			True,
		},
		{
			NumberFloatVal(0.2),
			NumberIntVal(0),
			False,
		},
		{
			NumberFloatVal(0.2),
			NumberFloatVal(0.2),
			True,
		},
		{
			UnknownVal(Number),
			UnknownVal(Number),
			UnknownVal(Bool).RefineNotNull(),
		},
		{
			NumberIntVal(1),
			UnknownVal(Number),
			UnknownVal(Bool).RefineNotNull(),
		},
		{
			UnknownVal(Number),
			NumberIntVal(1),
			UnknownVal(Bool).RefineNotNull(),
		},
		{
			DynamicVal,
			DynamicVal,
			UnknownVal(Bool).RefineNotNull(),
		},
		{
			NumberIntVal(1),
			DynamicVal,
			UnknownVal(Bool).RefineNotNull(),
		},
		{
			DynamicVal,
			NumberIntVal(1),
			UnknownVal(Bool).RefineNotNull(),
		},
		{
			NumberIntVal(0).Mark(1),
			NumberIntVal(1),
			True.Mark(1),
		},
		{
			NumberIntVal(0),
			NumberIntVal(1).Mark(1),
			True.Mark(1),
		},
		{
			NumberIntVal(0).Mark(1),
			NumberIntVal(1).Mark(1),
			True.Mark(1),
		},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("%#v.LessThanOrEqualTo(%#v)", test.Receiver, test.Other), func(t *testing.T) {
			got := test.Receiver.LessThanOrEqualTo(test.Other)
			if !got.RawEquals(test.Expected) {
				t.Fatalf("LessThanOrEqualTo returned %#v; want %#v", got, test.Expected)
			}
		})
	}
}

func TestGreaterThanOrEqualTo(t *testing.T) {
	tests := []struct {
		Receiver Value
		Other    Value
		Expected Value
	}{
		{
			NumberIntVal(0),
			NumberIntVal(1),
			False,
		},
		{
			NumberIntVal(1),
			NumberIntVal(0),
			True,
		},
		{
			NumberIntVal(0),
			NumberIntVal(0),
			True,
		},
		{
			NumberFloatVal(0.1),
			NumberFloatVal(0.2),
			False,
		},
		{
			NumberFloatVal(0.2),
			NumberFloatVal(0.1),
			True,
		},
		{
			NumberIntVal(0),
			NumberFloatVal(0.2),
			False,
		},
		{
			NumberFloatVal(0.2),
			NumberIntVal(0),
			True,
		},
		{
			NumberFloatVal(0.2),
			NumberFloatVal(0.2),
			True,
		},
		{
			UnknownVal(Number),
			UnknownVal(Number),
			UnknownVal(Bool).RefineNotNull(),
		},
		{
			NumberIntVal(1),
			UnknownVal(Number),
			UnknownVal(Bool).RefineNotNull(),
		},
		{
			UnknownVal(Number),
			NumberIntVal(1),
			UnknownVal(Bool).RefineNotNull(),
		},
		{
			DynamicVal,
			DynamicVal,
			UnknownVal(Bool).RefineNotNull(),
		},
		{
			NumberIntVal(1),
			DynamicVal,
			UnknownVal(Bool).RefineNotNull(),
		},
		{
			DynamicVal,
			NumberIntVal(1),
			UnknownVal(Bool).RefineNotNull(),
		},
		{
			NumberIntVal(0).Mark(1),
			NumberIntVal(1),
			False.Mark(1),
		},
		{
			NumberIntVal(0),
			NumberIntVal(1).Mark(1),
			False.Mark(1),
		},
		{
			NumberIntVal(0).Mark(1),
			NumberIntVal(1).Mark(1),
			False.Mark(1),
		},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("%#v.GreaterThanOrEqualTo(%#v)", test.Receiver, test.Other), func(t *testing.T) {
			got := test.Receiver.GreaterThanOrEqualTo(test.Other)
			if !got.RawEquals(test.Expected) {
				t.Fatalf("GreaterThanOrEqualTo returned %#v; want %#v", got, test.Expected)
			}
		})
	}
}

func TestValueGoString(t *testing.T) {
	tests := []struct {
		Value Value
		Want  string
	}{
		{
			NullVal(DynamicPseudoType),
			`cty.NullVal(cty.DynamicPseudoType)`,
		},
		{
			NullVal(String),
			`cty.NullVal(cty.String)`,
		},
		{
			NullVal(Tuple([]Type{String, Bool})),
			`cty.NullVal(cty.Tuple([]cty.Type{cty.String, cty.Bool}))`,
		},
		{
			UnknownVal(DynamicPseudoType),
			`cty.DynamicVal`,
		},
		{
			UnknownVal(String),
			`cty.UnknownVal(cty.String)`,
		},
		{
			UnknownVal(Tuple([]Type{String, Bool})),
			`cty.UnknownVal(cty.Tuple([]cty.Type{cty.String, cty.Bool}))`,
		},
		{
			UnknownVal(String).Refine().NotNull().NewValue(),
			`cty.UnknownVal(cty.String).RefineNotNull()`,
		},
		{
			UnknownVal(String).Refine().NotNull().StringPrefix("a-").NewValue(),
			`cty.UnknownVal(cty.String).Refine().NotNull().StringPrefixFull("a-").NewValue()`,
		},
		{
			UnknownVal(String).Refine().NotNull().StringPrefix("foo").NewValue(), // The last character of the prefix gets discarded in case the next character is a combining diacritic
			`cty.UnknownVal(cty.String).Refine().NotNull().StringPrefixFull("fo").NewValue()`,
		},
		{
			UnknownVal(Bool).Refine().NotNull().NewValue(),
			`cty.UnknownVal(cty.Bool).RefineNotNull()`,
		},
		{
			UnknownVal(Number).Refine().NumberRangeInclusive(Zero, UnknownVal(Number)).NewValue(),
			`cty.UnknownVal(cty.Number).Refine().NumberLowerBound(cty.NumberIntVal(0), true).NewValue()`,
		},
		{
			UnknownVal(Number).Refine().NumberRangeInclusive(Zero, NumberIntVal(1)).NewValue(),
			`cty.UnknownVal(cty.Number).Refine().NumberLowerBound(cty.NumberIntVal(0), true).NumberUpperBound(cty.NumberIntVal(1), true).NewValue()`,
		},

		{
			StringVal(""),
			`cty.StringVal("")`,
		},
		{
			StringVal("hello"),
			`cty.StringVal("hello")`,
		},

		{
			Zero,
			`cty.NumberIntVal(0)`,
		},
		{
			NumberFloatVal(1.2),
			`cty.NumberFloatVal(1.2)`,
		},
		{
			NumberFloatVal(1.0),
			`cty.NumberIntVal(1)`, // the "float-ness" of the input is lost because its value is a whole number
		},
		{
			MustParseNumberVal("3.14159265358979323846264338327950288419716939937510582097494459"),
			`cty.MustParseNumberVal("3.14159265358979323846264338327950288419716939937510582097494459")`,
		},

		{
			True,
			`cty.True`,
		},
		{
			False,
			`cty.False`,
		},

		{
			ListValEmpty(String),
			`cty.ListValEmpty(cty.String)`,
		},
		{
			ListValEmpty(List(String)),
			`cty.ListValEmpty(cty.List(cty.String))`,
		},
		{
			ListVal([]Value{True}),
			`cty.ListVal([]cty.Value{cty.True})`,
		},

		{
			SetValEmpty(String),
			`cty.SetValEmpty(cty.String)`,
		},
		{
			SetValEmpty(Map(String)),
			`cty.SetValEmpty(cty.Map(cty.String))`,
		},
		{
			SetVal([]Value{True}),
			`cty.SetVal([]cty.Value{cty.True})`,
		},

		{
			EmptyTupleVal,
			`cty.EmptyTupleVal`,
		},
		{
			TupleVal(nil),
			`cty.EmptyTupleVal`,
		},
		{
			TupleVal([]Value{True}),
			`cty.TupleVal([]cty.Value{cty.True})`,
		},

		{
			MapValEmpty(String),
			`cty.MapValEmpty(cty.String)`,
		},
		{
			MapValEmpty(Set(String)),
			`cty.MapValEmpty(cty.Set(cty.String))`,
		},
		{
			MapVal(map[string]Value{"boop": True}),
			`cty.MapVal(map[string]cty.Value{"boop":cty.True})`,
		},

		{
			EmptyObjectVal,
			`cty.EmptyObjectVal`,
		},
		{
			ObjectVal(nil),
			`cty.EmptyObjectVal`,
		},
		{
			ObjectVal(map[string]Value{"foo": True}),
			`cty.ObjectVal(map[string]cty.Value{"foo":cty.True})`,
		},
	}

	for _, test := range tests {
		t.Run(test.Value.GoString(), func(t *testing.T) {
			got := test.Value.GoString()
			want := test.Want
			if got != want {
				t.Errorf("wrong result\ngot:  %s\nwant: %s", got, want)
			}
		})
	}
}

func TestHasWhollyKnownType(t *testing.T) {
	tests := []struct {
		Value Value
		Want  bool
	}{
		{
			Value: DynamicVal,
			Want:  false,
		},
		{
			Value: ObjectVal(map[string]Value{
				"dyn": DynamicVal,
			}),
			Want: false,
		},
		{
			Value: NullVal(Object(map[string]Type{
				"dyn": DynamicPseudoType,
			})),
			Want: true,
		},
		{
			Value: TupleVal([]Value{
				StringVal("a"),
				NullVal(DynamicPseudoType),
			}),
			Want: true,
		},
		{
			Value: ListVal([]Value{
				ObjectVal(map[string]Value{
					"null": NullVal(DynamicPseudoType),
				}),
			}),
			Want: true,
		},
		{
			Value: ListVal([]Value{
				NullVal(Object(map[string]Type{
					"dyn": DynamicPseudoType,
				})),
			}),
			Want: true,
		},
		{
			Value: ObjectVal(map[string]Value{
				"tuple": TupleVal([]Value{
					StringVal("a"),
					NullVal(DynamicPseudoType),
				}),
			}),
			Want: true,
		},
		{
			Value: ObjectVal(map[string]Value{
				"tuple": TupleVal([]Value{
					ObjectVal(map[string]Value{
						"dyn": DynamicVal,
					}),
				}),
			}),
			Want: false,
		},
	}
	for _, test := range tests {
		t.Run(test.Value.GoString(), func(t *testing.T) {
			got := test.Value.HasWhollyKnownType()
			want := test.Want
			if got != want {
				t.Errorf("wrong result\ngot:  %v\nwant: %v", got, want)
			}
		})
	}
}

func TestHasElement(t *testing.T) {
	tests := []struct {
		Set  Value
		Elem Value
		Want Value
	}{
		{
			SetValEmpty(String),
			StringVal("hello"),
			False,
		},
		{
			SetVal([]Value{
				StringVal("hello"),
			}),
			StringVal("hello"),
			True,
		},
		{
			SetVal([]Value{
				StringVal("hello"),
				StringVal("world"),
			}),
			StringVal("hello"),
			True,
		},
		{
			SetVal([]Value{
				StringVal("hello"),
				StringVal("world"),
			}),
			StringVal("hi"),
			False,
		},
		{
			SetVal([]Value{
				StringVal("hello"),
				UnknownVal(String),
			}),
			StringVal("hello"),
			True, // "hello" is definitely present regardless of what the unknown value is
		},
		{
			SetVal([]Value{
				StringVal("hello"),
				UnknownVal(String),
			}),
			StringVal("world"),
			UnknownVal(Bool).RefineNotNull(), // The unknown value might turn out to be "world"
		},
		{
			SetVal([]Value{
				UnknownVal(String),
			}),
			StringVal("world"),
			UnknownVal(Bool).RefineNotNull(), // The unknown value might turn out to be "world"
		},
		{
			SetVal([]Value{
				UnknownVal(String),
				UnknownVal(String),
			}),
			StringVal("world"),
			UnknownVal(Bool).RefineNotNull(), // One of the unknown values might turn out to be "world"
		},
		{
			SetVal([]Value{
				StringVal("hello"),
				UnknownVal(String),
			}),
			True,
			False, // A set of string cannot possibly contain a bool
		},
		{
			SetVal([]Value{
				StringVal("hello"),
				UnknownVal(String),
			}),
			UnknownVal(String),
			// The unknowns are placeholders for values, not values themselves, so the presence of an unknown
			// in the set doesn't cause this to return true: there's no guarantee that both of the unknowns
			// above will be equal once finalized.
			UnknownVal(Bool).RefineNotNull(),
		},
		{
			SetVal([]Value{
				StringVal("hello"),
				StringVal("world"),
			}),
			UnknownVal(String),
			UnknownVal(Bool).RefineNotNull(),
		},
		{
			SetVal([]Value{
				StringVal("hello"),
				StringVal("world"),
			}),
			DynamicVal,
			UnknownVal(Bool).RefineNotNull(),
		},
		{
			DynamicVal,
			StringVal("hello"),
			UnknownVal(Bool).RefineNotNull(),
		},
		{
			SetVal([]Value{
				NullVal(DynamicPseudoType),
			}),
			NullVal(DynamicPseudoType),
			True,
		},
		{
			SetVal([]Value{
				DynamicVal,
			}),
			NullVal(DynamicPseudoType),
			UnknownVal(Bool).RefineNotNull(),
		},
		{
			SetVal([]Value{
				DynamicVal,
			}),
			DynamicVal,
			UnknownVal(Bool).RefineNotNull(),
		},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("%#v has %#v", test.Set, test.Elem), func(t *testing.T) {
			t.Logf("%#v.HasElement(%#v)", test.Set, test.Elem)

			got := test.Set.HasElement(test.Elem)
			if !test.Want.RawEquals(got) {
				t.Errorf("wrong result\ngot:  %#v\nwant: %#v", got, test.Want)
			}
		})
	}
}

func TestElements(t *testing.T) {
	tests := []struct {
		Input Value
		Want  [][2]Value
	}{
		{
			ListValEmpty(String),
			[][2]Value{},
		},
		{
			ListVal([]Value{
				StringVal("hello"),
				StringVal("world"),
			}),
			[][2]Value{
				{NumberIntVal(0), StringVal("hello")},
				{NumberIntVal(1), StringVal("world")},
			},
		},
		{
			TupleVal([]Value{
				StringVal("hello"),
				StringVal("world"),
			}),
			[][2]Value{
				{NumberIntVal(0), StringVal("hello")},
				{NumberIntVal(1), StringVal("world")},
			},
		},
		{
			SetVal([]Value{
				StringVal("hello"),
				StringVal("world"),
			}),
			[][2]Value{
				// When the element type is string, the results are returned
				// in lexicographical order. Otherwise the order is unspecified.
				{StringVal("hello"), StringVal("hello")},
				{StringVal("world"), StringVal("world")},
			},
		},
		{
			MapVal(map[string]Value{
				"greeting": StringVal("hello"),
				"greetee":  StringVal("world"),
			}),
			[][2]Value{
				{StringVal("greetee"), StringVal("world")},
				{StringVal("greeting"), StringVal("hello")},
			},
		},
		{
			ObjectVal(map[string]Value{
				"greeting": StringVal("hello"),
				"greetee":  StringVal("world"),
			}),
			[][2]Value{
				{StringVal("greetee"), StringVal("world")},
				{StringVal("greeting"), StringVal("hello")},
			},
		},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("%#v", test.Input), func(t *testing.T) {
			t.Logf("%#v.Elements()", test.Input)

			got := make([][2]Value, 0)
			for k, v := range test.Input.Elements() {
				got = append(got, [2]Value{k, v})
			}

			compare := cmp.Comparer(func(a, b Value) bool { return a.RawEquals(b) })
			if diff := cmp.Diff(test.Want, got, compare); diff != "" {
				t.Error("wrong result\n" + diff)
			}
		})
	}
}

func TestFloatCopy(t *testing.T) {
	// ensure manipulating floats does not modify the cty.Value
	v := NumberFloatVal(1.9)
	vString := v.GoString()

	// do something that will modify the internal big.Float mantissa
	v.AsBigFloat().SetInt64(1)

	if vString != v.GoString() {
		t.Fatalf("original value changed from %s to %#v", vString, v)
	}
}
