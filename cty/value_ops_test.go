package cty

import (
	"fmt"
	"reflect"
	"testing"
)

func TestValueEquals(t *testing.T) {
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

		// Unknowns and Dynamics
		{
			NumberIntVal(2),
			UnknownVal(Number),
			UnknownVal(Bool),
		},
		{
			NumberIntVal(1),
			DynamicVal,
			UnknownVal(Bool),
		},
		{
			DynamicVal,
			BoolVal(true),
			UnknownVal(Bool),
		},
		{
			DynamicVal,
			DynamicVal,
			UnknownVal(Bool),
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
			UnknownVal(Bool),
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
			UnknownVal(Bool),
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
			UnknownVal(Bool),
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
			UnknownVal(Bool),
		},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("%#v.Equals(%#v)", test.LHS, test.RHS), func(t *testing.T) {
			got := test.LHS.Equals(test.RHS)
			if !got.RawEquals(test.Expected) {
				t.Fatalf("Equals returned %#v; want %#v", got, test.Expected)
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
			UnknownVal(Number),
		},
		{
			UnknownVal(Number),
			UnknownVal(Number),
			UnknownVal(Number),
		},
		{
			NumberIntVal(1),
			DynamicVal,
			DynamicVal,
		},
		{
			DynamicVal,
			DynamicVal,
			DynamicVal,
		},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("%#v.Add(%#v)", test.LHS, test.RHS), func(t *testing.T) {
			got := test.LHS.Add(test.RHS)
			if !got.RawEquals(test.Expected) {
				t.Fatalf("Add returned %#v; want %#v", got, test.Expected)
			}
		})
	}
}

func TestValueSub(t *testing.T) {
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
			UnknownVal(Number),
		},
		{
			UnknownVal(Number),
			UnknownVal(Number),
			UnknownVal(Number),
		},
		{
			NumberIntVal(1),
			DynamicVal,
			DynamicVal,
		},
		{
			DynamicVal,
			DynamicVal,
			DynamicVal,
		},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("%#v.Sub(%#v)", test.LHS, test.RHS), func(t *testing.T) {
			got := test.LHS.Sub(test.RHS)
			if !got.RawEquals(test.Expected) {
				t.Fatalf("Sub returned %#v; want %#v", got, test.Expected)
			}
		})
	}
}

func TestValueNeg(t *testing.T) {
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
			UnknownVal(Number),
		},
		{
			DynamicVal,
			DynamicVal,
		},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("%#v.Neg()", test.Receiver), func(t *testing.T) {
			got := test.Receiver.Neg()
			if !got.RawEquals(test.Expected) {
				t.Fatalf("Neg returned %#v; want %#v", got, test.Expected)
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
			UnknownVal(Object(map[string]Type{
				"greeting": String,
			})),
			"greeting",
			UnknownVal(String),
		},
		{
			DynamicVal,
			"hello",
			DynamicVal,
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
				NumberIntVal(2),
			}),
			[]call{
				// Ordering is arbitrary but consistent, so future changes
				// to the set implementation may reorder these.
				{NilVal, NumberIntVal(2)},
				{NilVal, NumberIntVal(1)},
			},
			false,
		},
		{
			SetVal([]Value{
				StringVal("hi"),
				StringVal("stop"),
				StringVal("hey"),
			}),
			[]call{
				// Ordering is arbitrary but consistent, so future changes
				// to the set implementation may reorder these.
				{NilVal, StringVal("hi")},
				{NilVal, StringVal("stop")},
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
