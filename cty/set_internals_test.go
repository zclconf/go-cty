package cty

import (
	"fmt"
	"math/big"
	"testing"

	"github.com/zclconf/go-cty/cty/set"
)

func TestSetHashBytes(t *testing.T) {
	tests := []struct {
		value     Value
		want      string
		wantMarks ValueMarks
	}{
		{
			UnknownVal(Number),
			"?",
			nil,
		},
		{
			UnknownVal(String),
			"?",
			nil,
		},
		{
			NullVal(Number),
			"~",
			nil,
		},
		{
			NullVal(String),
			"~",
			nil,
		},
		{
			DynamicVal,
			"?",
			nil,
		},
		{
			NumberVal(big.NewFloat(12)),
			"12",
			nil,
		},
		{
			// This weird case is an intentionally-invalid number value that
			// mimics the incorrect result of a gob round-trip of a cty.Number
			// value. For more information, see the function
			// gobDecodeFixNumberPtr. Unfortunately the set internals need to
			// be tolerant of this situation because gob-decoding a set
			// causes this situation to arise before we have had an opportunity
			// to run gobDecodeFixNumberPtr yet.
			Value{
				ty: Number,
				v:  *big.NewFloat(13),
			},
			"13",
			nil,
		},
		{
			StringVal(""),
			`""`,
			nil,
		},
		{
			StringVal("pizza"),
			`"pizza"`,
			nil,
		},
		{
			True,
			"T",
			nil,
		},
		{
			False,
			"F",
			nil,
		},
		{
			ListValEmpty(Bool),
			"[]",
			nil,
		},
		{
			ListValEmpty(DynamicPseudoType),
			"[]",
			nil,
		},
		{
			ListVal([]Value{True, False}),
			"[T;F;]",
			nil,
		},
		{
			ListVal([]Value{UnknownVal(Bool)}),
			"[?;]",
			nil,
		},
		{
			ListVal([]Value{ListValEmpty(Bool)}),
			"[[];]",
			nil,
		},
		{
			MapValEmpty(Bool),
			"{}",
			nil,
		},
		{
			MapVal(map[string]Value{"true": True, "false": False}),
			`{"false":F;"true":T;}`,
			nil,
		},
		{
			MapVal(map[string]Value{"true": True, "unknown": UnknownVal(Bool), "dynamic": DynamicVal}),
			`{"dynamic":?;"true":T;"unknown":?;}`,
			nil,
		},
		{
			SetValEmpty(Bool),
			"[]",
			nil,
		},
		{
			SetVal([]Value{True, True, False}),
			"[F;T;]",
			nil,
		},
		{
			SetVal([]Value{UnknownVal(Bool), UnknownVal(Bool)}),
			"[?;?;]", // unknowns are never equal, so we can have multiple of them
			nil,
		},
		{
			EmptyObjectVal,
			"<>",
			nil,
		},
		{
			ObjectVal(map[string]Value{
				"name": StringVal("ermintrude"),
				"age":  NumberVal(big.NewFloat(54)),
			}),
			`<54;"ermintrude";>`,
			nil,
		},
		{
			EmptyTupleVal,
			"<>",
			nil,
		},
		{
			TupleVal([]Value{
				StringVal("ermintrude"),
				NumberVal(big.NewFloat(54)),
			}),
			`<"ermintrude";54;>`,
			nil,
		},

		// Marked values
		{
			StringVal("pizza").Mark(1),
			`"pizza"`,
			NewValueMarks(1),
		},
		{
			ObjectVal(map[string]Value{
				"name": StringVal("ermintrude").Mark(1),
				"age":  NumberVal(big.NewFloat(54)).Mark(2),
			}),
			`<54;"ermintrude";>`,
			NewValueMarks(1, 2),
		},
	}

	for _, test := range tests {
		t.Run(gobDecodeFixNumberPtrVal(test.value).GoString(), func(t *testing.T) {
			gotRaw, gotMarks := makeSetHashBytes(test.value)
			got := string(gotRaw)
			if got != test.want {
				t.Errorf("wrong result\ngot:  %s\nwant: %s", got, test.want)
			}
			if !test.wantMarks.Equal(gotMarks) {
				t.Errorf("wrong result marks\ngot:  %#v\nwant: %#v", gotMarks, test.wantMarks)
			}
		})
	}
}

func TestSetOrder(t *testing.T) {
	tests := []struct {
		a, b Value
		want bool
	}{
		// Strings sort lexicographically (this is a compatibility constraint)
		{
			StringVal("a"),
			StringVal("b"),
			true,
		},
		{
			StringVal("b"),
			StringVal("a"),
			false,
		},
		{
			UnknownVal(String),
			StringVal("a"),
			false,
		},
		{
			StringVal("a"),
			UnknownVal(String),
			true,
		},

		// Numbers sort numerically (this is a compatibility constraint)
		{
			Zero,
			NumberIntVal(1),
			true,
		},
		{
			NumberIntVal(1),
			Zero,
			false,
		},

		// Booleans sort false before true (this is a compatibility constraint)
		{
			False,
			True,
			true,
		},
		{
			True,
			False,
			false,
		},

		// Unknown and Null values push to the end of a sort (this is a compatibility constraint)
		{
			UnknownVal(String),
			UnknownVal(String),
			false, // no defined ordering
		},
		{
			NullVal(String),
			StringVal("a"),
			false,
		},
		{
			StringVal("a"),
			NullVal(String),
			true,
		},
		{
			UnknownVal(String),
			NullVal(String),
			true,
		},
		{
			NullVal(String),
			UnknownVal(String),
			false,
		},

		// All other types just use an arbitrary fallback sort. These results
		// are _not_ compatibility constraints but we are testing them here
		// to verify that the result is consistent between runs for a
		// specific version of cty.
		{
			ListValEmpty(String),
			ListVal([]Value{StringVal("boop")}),
			false,
		},
		{
			ListVal([]Value{StringVal("boop")}),
			ListValEmpty(String),
			true,
		},
		{
			SetValEmpty(String),
			SetVal([]Value{StringVal("boop")}),
			false,
		},
		{
			SetVal([]Value{StringVal("boop")}),
			SetValEmpty(String),
			true,
		},
		{
			MapValEmpty(String),
			MapVal(map[string]Value{"blah": StringVal("boop")}),
			false,
		},
		{
			MapVal(map[string]Value{"blah": StringVal("boop")}),
			MapValEmpty(String),
			true,
		},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("%#v < %#v", test.a, test.b), func(t *testing.T) {
			rules := setRules{test.a.Type()} // both values are assumed to have the same type
			got := rules.Less(test.a.v, test.b.v)
			if got != test.want {
				t.Errorf("wrong result\na: %#v\nb: %#v\ngot:  %#v\nwant: %#v", test.a, test.b, got, test.want)
			}
		})
	}
}

func TestSetRulesSameRules(t *testing.T) {
	tests := []struct {
		a    set.Rules
		b    set.Rules
		want bool
	}{
		{
			setRules{EmptyObject},
			setRules{DynamicPseudoType},
			false,
		},
		{
			setRules{EmptyObject},
			setRules{EmptyObject},
			true,
		},
		{
			setRules{String},
			setRules{String},
			true,
		},
		{
			setRules{Object(map[string]Type{"a": String})},
			setRules{Object(map[string]Type{"a": String})},
			true,
		},
		{
			setRules{Object(map[string]Type{"a": String})},
			setRules{Object(map[string]Type{"a": Bool})},
			false,
		},
		{
			pathSetRules{},
			pathSetRules{},
			true,
		},
		{
			setRules{DynamicPseudoType},
			pathSetRules{},
			false,
		},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("%#v.SameRules(%#v)", test.a, test.b), func(t *testing.T) {
			got := test.a.SameRules(test.b)
			if got != test.want {
				t.Errorf("wrong result\na: %#v\nb: %#v\ngot %#v, want %#v", test.a, test.b, got, test.want)
			}
		})
	}
}
