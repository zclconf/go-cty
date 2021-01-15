package stdlib

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"testing"

	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/function/functest"
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
			cty.UnknownVal(cty.String),
		},
		{
			cty.ObjectVal(map[string]cty.Value{"dunno": cty.UnknownVal(cty.Bool), "false": cty.False}),
			cty.UnknownVal(cty.String),
		},
		{
			cty.DynamicVal,
			cty.UnknownVal(cty.String),
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

	// Property-based tests against randomly-selected inputs
	// These can be quite time-consuming when our random generator produces
	// larger data structures, so we use property-based testing only sparingly
	// for this function.
	t.Run(
		"produces parseable JSON for all known, unmarked values",
		functest.Test(
			functest.GenFixedArgs(
				functest.GenAnySerializableValues(),
			),
			func(args []cty.Value) bool {
				v, err := JSONEncodeFunc.Call(args)
				if err != nil || v.Type() != cty.String {
					return false
				}
				src := v.AsString()
				r := strings.NewReader(src)
				dec := json.NewDecoder(r)
				for {
					_, err := dec.Token()
					if err == io.EOF {
						return true // tokenized the whole thing
					}
					if err != nil {
						return false
					}
				}
			},
		).Run,
	)
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

	// Property-based tests against randomly-selected inputs
	// These can be quite time-consuming when our random generator produces
	// larger data structures, so we use property-based testing only sparingly
	// for this function.

	// We'll use our sibling JSONEncode function to help us construct valid
	// JSON input to test with. That's only as good as the correctness of that
	// function of course, but it has its own tests.
	genJSONStrings := functest.GenAnySerializableValues().Map(func(v cty.Value) cty.Value {
		v, err := JSONEncode(v)
		if err != nil {
			panic(fmt.Sprintf("JSONEncode failed: %s", err))
		}
		return v
	})
	t.Run(
		"JSONEncode is the inverse of JSONDecode for all serializable values",
		functest.TestInverse(
			genJSONStrings,
			JSONDecode,
			JSONEncode,
		).Run,
	)
}
