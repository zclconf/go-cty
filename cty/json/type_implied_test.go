package json

import (
	"testing"

	"github.com/zclconf/go-cty/cty"
)

func TestImpliedType(t *testing.T) {
	tests := []struct {
		Input string
		Want  cty.Type
	}{
		{
			"null",
			cty.DynamicPseudoType,
		},
		{
			"1",
			cty.Number,
		},
		{
			"1.2222222222222222222222222222222222",
			cty.Number,
		},
		{
			"999999999999999999999999999999999999999999999999999999999999",
			cty.Number,
		},
		{
			`""`,
			cty.String,
		},
		{
			`"hello"`,
			cty.String,
		},
		{
			"true",
			cty.Bool,
		},
		{
			"false",
			cty.Bool,
		},
		{
			"{}",
			cty.EmptyObject,
		},
		{
			`{"true": true}`,
			cty.Object(map[string]cty.Type{
				"true": cty.Bool,
			}),
		},
		{
			`{"true": true, "name": "Ermintrude", "null": null}`,
			cty.Object(map[string]cty.Type{
				"true": cty.Bool,
				"name": cty.String,
				"null": cty.DynamicPseudoType,
			}),
		},
		{
			"[]",
			cty.EmptyTuple,
		},
		{
			"[true, 1.2, null]",
			cty.Tuple([]cty.Type{cty.Bool, cty.Number, cty.DynamicPseudoType}),
		},
		{
			`[[true], [1.2], [null]]`,
			cty.Tuple([]cty.Type{
				cty.Tuple([]cty.Type{cty.Bool}),
				cty.Tuple([]cty.Type{cty.Number}),
				cty.Tuple([]cty.Type{cty.DynamicPseudoType}),
			}),
		},
		{
			`[{"true": true}, {"name": "Ermintrude"}, {"null": null}]`,
			cty.Tuple([]cty.Type{
				cty.Object(map[string]cty.Type{
					"true": cty.Bool,
				}),
				cty.Object(map[string]cty.Type{
					"name": cty.String,
				}),
				cty.Object(map[string]cty.Type{
					"null": cty.DynamicPseudoType,
				}),
			}),
		},
		{
			`{"a": "hello", "a": "world"}`,
			cty.Object(map[string]cty.Type{
				"a": cty.String,
			}),
		},
	}

	for _, test := range tests {
		t.Run(test.Input, func(t *testing.T) {
			got, err := ImpliedType([]byte(test.Input))

			if err != nil {
				t.Fatalf("unexpected error: %s", err)
			}

			if !got.Equals(test.Want) {
				t.Errorf(
					"wrong type\ninput: %s\ngot:   %#v\nwant:  %#v",
					test.Input, got, test.Want,
				)
			}
		})
	}
}

func TestImpliedTypeErrors(t *testing.T) {
	tests := []struct {
		Input     string
		WantError string
	}{
		{
			`{"a": "hello", "a": true}`,
			`duplicate "a" property in JSON object`,
		},
		{
			`{}boop`,
			`extraneous data after JSON object`,
		},
		{
			`[!]`,
			`invalid character '!' looking for beginning of value`,
		},
		{
			`[}`,
			`invalid character '}' looking for beginning of value`,
		},
		{
			`{true: null}`,
			`invalid character 't'`,
		},
	}

	for _, test := range tests {
		t.Run(test.Input, func(t *testing.T) {
			_, err := ImpliedType([]byte(test.Input))
			if err == nil {
				t.Fatalf("unexpected success\nwant error: %s", err)
			}

			if got, want := err.Error(), test.WantError; got != want {
				t.Errorf("wrong error\ngot:  %s\nwant: %s", got, want)
			}
		})
	}
}
