package stdlib

import (
	"fmt"
	"testing"

	"github.com/zclconf/go-cty/cty"
)

func TestJSONEncode(t *testing.T) {
	tests := []struct {
		Input cty.Value
		Want  cty.Value
	}{
		// This does not comprehensively test all possible inputs because
		// the underlying functions in package json already have tests of
		// their own. Here we are mainly concerned with seeing that the
		// function's definition accepts all reasonable values.
		{
			cty.NumberIntVal(15),
			cty.StringVal(`15`),
		},
		{
			cty.StringVal("hello"),
			cty.StringVal(`"hello"`),
		},
		{
			cty.True,
			cty.StringVal(`true`),
		},
		{
			cty.ListValEmpty(cty.Number),
			cty.StringVal(`[]`),
		},
		{
			cty.ListVal([]cty.Value{cty.True, cty.False}),
			cty.StringVal(`[true,false]`),
		},
		{
			cty.ObjectVal(map[string]cty.Value{"true": cty.True, "false": cty.False}),
			cty.StringVal(`{"false":false,"true":true}`),
		},
		{
			cty.UnknownVal(cty.Number),
			cty.UnknownVal(cty.String).RefineNotNull(),
		},
		{
			cty.ObjectVal(map[string]cty.Value{"dunno": cty.UnknownVal(cty.Bool), "false": cty.False}),
			cty.UnknownVal(cty.String).Refine().NotNull().StringPrefixFull("{").NewValue(),
		},
		{
			cty.ListVal([]cty.Value{cty.UnknownVal(cty.String)}),
			cty.UnknownVal(cty.String).Refine().NotNull().StringPrefixFull("[").NewValue(),
		},
		{
			cty.UnknownVal(cty.String),
			cty.UnknownVal(cty.String).RefineNotNull(), // Can't refine the prefix because the input might be null
		},
		{
			cty.UnknownVal(cty.String).RefineNotNull(),
			cty.UnknownVal(cty.String).Refine().NotNull().StringPrefixFull(`"`).NewValue(),
		},
		{
			cty.UnknownVal(cty.Number),
			cty.UnknownVal(cty.String).RefineNotNull(),
		},
		{
			cty.UnknownVal(cty.Bool),
			cty.UnknownVal(cty.String).RefineNotNull(),
		},
		{
			cty.DynamicVal,
			cty.UnknownVal(cty.String).RefineNotNull(),
		},
		{
			cty.NullVal(cty.String),
			cty.StringVal("null"),
		},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("JSONEncode(%#v)", test.Input), func(t *testing.T) {
			got, err := JSONEncode(test.Input)

			if err != nil {
				t.Fatalf("unexpected error: %s", err)
			}

			if !got.RawEquals(test.Want) {
				t.Errorf("wrong result\ngot:  %#v\nwant: %#v", got, test.Want)
			}
		})
	}
}

func TestJSONDecode(t *testing.T) {
	tests := []struct {
		Input cty.Value
		Want  cty.Value
	}{
		{
			cty.StringVal(`15`),
			cty.NumberIntVal(15),
		},
		{
			cty.StringVal(`"hello"`),
			cty.StringVal("hello"),
		},
		{
			cty.StringVal(`true`),
			cty.True,
		},
		{
			cty.StringVal(`[]`),
			cty.EmptyTupleVal,
		},
		{
			cty.StringVal(`[true,false]`),
			cty.TupleVal([]cty.Value{cty.True, cty.False}),
		},
		{
			cty.StringVal(`{"false":false,"true":true}`),
			cty.ObjectVal(map[string]cty.Value{"true": cty.True, "false": cty.False}),
		},
		{
			cty.UnknownVal(cty.String),
			cty.DynamicVal, // need to know the value to determine the type
		},
		{
			cty.UnknownVal(cty.String).Refine().StringPrefixFull("1").NewValue(),
			cty.UnknownVal(cty.Number), // deduced from refinement
		},
		{
			cty.UnknownVal(cty.String).Refine().StringPrefixFull("-").NewValue(),
			cty.UnknownVal(cty.Number), // deduced from refinement
		},
		{
			cty.UnknownVal(cty.String).Refine().StringPrefixFull(".").NewValue(),
			cty.UnknownVal(cty.Number), // deduced from refinement
		},
		{
			cty.UnknownVal(cty.String).Refine().StringPrefixFull("t").NewValue(),
			cty.UnknownVal(cty.Bool), // deduced from refinement
		},
		{
			cty.UnknownVal(cty.String).Refine().StringPrefixFull("f").NewValue(),
			cty.UnknownVal(cty.Bool), // deduced from refinement
		},
		{
			cty.UnknownVal(cty.String).Refine().StringPrefixFull(`"blurt`).NewValue(),
			cty.UnknownVal(cty.String), // deduced from refinement
		},
		{
			cty.UnknownVal(cty.String).Refine().StringPrefixFull(`{`).NewValue(),
			cty.DynamicVal, // can't deduce the result type, but potentially valid syntax
		},
		{
			cty.UnknownVal(cty.String).Refine().StringPrefixFull(`[`).NewValue(),
			cty.DynamicVal, // can't deduce the result type, but potentially valid syntax
		},
		{
			cty.DynamicVal,
			cty.DynamicVal,
		},
		{
			cty.StringVal(`true`).Mark(1),
			cty.True.Mark(1),
		},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("JSONDecode(%#v)", test.Input), func(t *testing.T) {
			got, err := JSONDecode(test.Input)

			if err != nil {
				t.Fatalf("unexpected error: %s", err)
			}

			if !got.RawEquals(test.Want) {
				t.Errorf("wrong result\ngot:  %#v\nwant: %#v", got, test.Want)
			}
		})
	}

	errorTests := []struct {
		Input     cty.Value
		WantError string
	}{
		{
			cty.StringVal("aaaa"),
			`invalid character 'a' looking for beginning of value`,
		},
		{
			cty.StringVal("nope"),
			`invalid character 'o' in literal null (expecting 'u')`, // (the 'n' looked like the beginning of 'null')
		},
		{
			cty.UnknownVal(cty.String).Refine().StringPrefixFull(`a`).NewValue(),
			`a JSON document cannot begin with the character 'a'`, // error deduced from refinement, despite full value being unknown
		},
	}
	for _, test := range errorTests {
		t.Run(fmt.Sprintf("JSONDecode(%#v)", test.Input), func(t *testing.T) {
			_, err := JSONDecode(test.Input)
			if err == nil {
				t.Fatal("unexpected success")
			}

			if got, want := err.Error(), test.WantError; got != want {
				t.Errorf("wrong error\ngot:  %s\nwant: %s", got, want)
			}
		})
	}
}
